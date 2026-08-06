// Notiflex API — B2B 알림 SaaS 플랫폼의 API 서버.
// 외부 프레임워크 없이 Go 표준 라이브러리만 사용한다.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/valkey-io/valkey-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// idCounterKey는 Valkey에서 ID 카운터를 저장하는 키다.
// 인메모리 카운터와 달리 모든 Pod이 같은 값을 공유한다.
const idCounterKey = "notiflex:id:counter"

// valkeyClient는 ID 발급에 쓰는 Valkey 연결이다.
var valkeyClient valkey.Client

// tracer는 요청 구간을 기록한다.
var tracer trace.Tracer = otel.Tracer("notiflex")

// notificationsTopic은 발급한 ID를 알림 이벤트로 보내는 토픽이다.
const notificationsTopic = "notifications"

// producer는 알림 이벤트를 Kafka로 보낸다. 브로커가 없으면 nil로 두고 동작을 건너뛴다.
var producer sarama.AsyncProducer

// podName은 응답이 어느 Pod에서 나왔는지 구분하기 위해 사용한다.
var podName = hostname()

// tier는 이 인스턴스가 어느 고객 등급을 담당하는지 나타낸다.
// SMB와 Enterprise를 별도 네임스페이스에서 운영한다.
var tier = envOr("NOTIFLEX_TIER", "smb")

// version은 배포된 코드의 버전이다. 빌드할 때 ldflags로 덮어쓸 수 있다.
var version = "0.7.1"

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// valkeyPassword는 비밀번호를 읽는다. Secret Manager CSI로 마운트한 파일이
// 있으면 그것을 먼저 쓰고, 없으면 환경변수로 넘어간다.
func valkeyPassword() string {
	if f := os.Getenv("VALKEY_PASSWORD_FILE"); f != "" {
		if data, err := os.ReadFile(f); err == nil {
			// 파일 끝의 개행이 섞이면 WRONGPASS가 난다.
			return strings.TrimSpace(string(data))
		} else {
			log.Printf("시크릿 파일을 읽지 못했다(%s): %v", f, err)
		}
	}
	return os.Getenv("VALKEY_PASSWORD")
}

// connectValkey는 연결을 재시도한다. Pod이 Valkey보다 먼저 뜨거나
// DNS가 아직 준비되지 않은 경우가 있어 즉시 종료하면 CrashLoopBackOff에 빠진다.
func connectValkey() (valkey.Client, error) {
	addr := envOr("VALKEY_ADDR", "valkey-primary.notiflex.svc.cluster.local:6379")
	pw := valkeyPassword()

	var lastErr error
	for i := 1; i <= 10; i++ {
		client, err := valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{addr},
			Password:    pw,
			// 단일 인스턴스라 복제본 조회를 시도하지 않는다.
			DisableCache: true,
		})
		if err == nil {
			return client, nil
		}
		lastErr = err
		log.Printf("Valkey 연결 재시도 %d/10: %v", i, err)
		time.Sleep(3 * time.Second)
	}
	return nil, lastErr
}

// initTracing은 OTLP gRPC로 Tempo에 트레이스를 보내도록 설정한다.
// 엔드포인트가 없으면 트레이싱 없이 동작한다.
func initTracing(ctx context.Context) func() {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		log.Printf("OTEL_EXPORTER_OTLP_ENDPOINT가 없어 트레이싱을 건너뛴다")
		return func() {}
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Printf("트레이스 exporter 생성 실패, 트레이싱 없이 계속한다: %v", err)
		return func() {}
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName("notiflex-api"),
		semconv.ServiceVersion(version),
		attribute.String("notiflex.tier", tier),
	))
	if err != nil {
		log.Printf("리소스 속성 생성 실패: %v", err)
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	tracer = tp.Tracer("notiflex")
	log.Printf("트레이싱 시작 (%s)", endpoint)

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Printf("트레이서 종료 실패: %v", err)
		}
	}
}

// traced는 핸들러를 감싸 요청마다 span을 만든다.
func traced(name string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), name)
		defer span.End()
		span.SetAttributes(
			attribute.String("notiflex.pod", podName),
			attribute.String("notiflex.tier", tier),
		)
		next(w, r.WithContext(ctx))
	}
}

// kafkaConfig는 브로커 버전에 맞춘 sarama 설정을 만든다.
func kafkaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = true
	cfg.ClientID = "notiflex-" + tier
	return cfg
}

// startKafka는 Producer와 Consumer를 준비한다.
// 브로커 주소가 없으면 Kafka 없이 동작한다(ch8 이전 상태와 호환).
func startKafka() {
	brokers := os.Getenv("KAFKA_BROKER")
	if brokers == "" {
		log.Printf("KAFKA_BROKER가 없어 이벤트 발행을 건너뛴다")
		return
	}
	list := strings.Split(brokers, ",")

	var err error
	for i := 1; i <= 10; i++ {
		producer, err = sarama.NewAsyncProducer(list, kafkaConfig())
		if err == nil {
			break
		}
		log.Printf("Kafka Producer 연결 재시도 %d/10: %v", i, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Printf("Kafka Producer 연결 실패, 이벤트 발행 없이 계속한다: %v", err)
		producer = nil
		return
	}
	log.Printf("Kafka Producer 연결 완료 (%s)", brokers)

	go func() {
		for e := range producer.Errors() {
			log.Printf("Kafka 발행 실패: %v", e)
		}
	}()

	go consumeNotifications(list)
}

// consumeNotifications는 발행된 알림 이벤트를 받아 로그로 남긴다.
// 실제 서비스라면 여기서 발송 처리를 한다.
func consumeNotifications(brokers []string) {
	var consumer sarama.Consumer
	var err error
	for i := 1; i <= 10; i++ {
		consumer, err = sarama.NewConsumer(brokers, kafkaConfig())
		if err == nil {
			break
		}
		log.Printf("Kafka Consumer 연결 재시도 %d/10: %v", i, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Printf("Kafka Consumer 연결 실패: %v", err)
		return
	}

	partitions, err := consumer.Partitions(notificationsTopic)
	if err != nil {
		log.Printf("파티션 조회 실패: %v", err)
		return
	}
	for _, part := range partitions {
		pc, err := consumer.ConsumePartition(notificationsTopic, part, sarama.OffsetNewest)
		if err != nil {
			log.Printf("파티션 %d 구독 실패: %v", part, err)
			continue
		}
		go func(pc sarama.PartitionConsumer, part int32) {
			for msg := range pc.Messages() {
				log.Printf("Kafka 수신 [partition=%d offset=%d] %s", part, msg.Offset, string(msg.Value))
			}
		}(pc, part)
	}
	log.Printf("Kafka Consumer 구독 시작 (%s, 파티션 %d개)", notificationsTopic, len(partitions))
}

// publishNotification은 발급한 ID를 알림 이벤트로 보낸다.
func publishNotification(id string) {
	if producer == nil {
		return
	}
	payload := fmt.Sprintf(`{"id":%q,"tier":%q,"pod":%q}`, id, tier, podName)
	producer.Input() <- &sarama.ProducerMessage{
		Topic: notificationsTopic,
		Key:   sarama.StringEncoder(tier),
		Value: sarama.StringEncoder(payload),
	}
}

func hostname() string {
	if h := os.Getenv("POD_NAME"); h != "" {
		return h
	}
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// requestCounts는 경로별 요청 수다. Prometheus가 /metrics로 긁어간다.
// 외부 클라이언트 라이브러리 없이 노출 형식을 직접 만든다.
var (
	metricsMu     sync.Mutex
	requestCounts = map[string]uint64{}
)

// countRequest는 핸들러를 감싸 경로별 요청 수를 센다.
func countRequest(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsMu.Lock()
		requestCounts[path]++
		metricsMu.Unlock()
		next(w, r)
	}
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	metricsMu.Lock()
	paths := make([]string, 0, len(requestCounts))
	for p := range requestCounts {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	snapshot := make(map[string]uint64, len(requestCounts))
	for _, p := range paths {
		snapshot[p] = requestCounts[p]
	}
	metricsMu.Unlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintln(w, "# HELP notiflex_http_requests_total 경로별 누적 HTTP 요청 수")
	fmt.Fprintln(w, "# TYPE notiflex_http_requests_total counter")
	for _, p := range paths {
		fmt.Fprintf(w, "notiflex_http_requests_total{path=%q,pod=%q} %d\n", p, podName, snapshot[p])
	}
	fmt.Fprintln(w, "# HELP notiflex_ids_generated_total 클러스터 전역에서 발급한 ID 총 개수")
	fmt.Fprintln(w, "# TYPE notiflex_ids_generated_total counter")
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if n, err := valkeyClient.Do(ctx, valkeyClient.B().Get().Key(idCounterKey).Build()).AsInt64(); err == nil {
		fmt.Fprintf(w, "notiflex_ids_generated_total{tier=%q} %d\n", tier, n)
	}
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("응답 인코딩 실패: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"status":  "ok",
		"pod":     podName,
		"version": version,
		"tier":    tier,
	})
}

func handleID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	incrCtx, incrSpan := tracer.Start(ctx, "valkey.incr")
	// INCR은 원자적이라 여러 Pod이 동시에 호출해도 번호가 겹치지 않는다.
	n, err := valkeyClient.Do(incrCtx, valkeyClient.B().Incr().Key(idCounterKey).Build()).AsInt64()
	incrSpan.End()
	if err != nil {
		log.Printf("ID 발급 실패: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, map[string]string{"error": "ID를 발급하지 못했다"})
		return
	}

	id := strconv.FormatInt(n, 10)
	_, pubSpan := tracer.Start(ctx, "kafka.publish")
	// 알림 발송은 요청을 붙잡지 않고 Kafka로 넘긴다.
	publishNotification(id)
	pubSpan.End()

	writeJSON(w, map[string]string{
		"id":           id,
		"generated_by": podName,
		"tier":         tier,
		"source":       "valkey",
	})
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"version":  version,
		"pod":      podName,
		"tier":     tier,
		"strategy": "canary",
	})
}

func main() {
	shutdownTracing := initTracing(context.Background())
	defer shutdownTracing()

	client, err := connectValkey()
	if err != nil {
		log.Fatalf("Valkey 연결 실패: %v", err)
	}
	valkeyClient = client
	defer valkeyClient.Close()
	log.Printf("Valkey 연결 완료")

	startKafka()
	if producer != nil {
		defer producer.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", countRequest("/health", traced("GET /health", handleHealth)))
	mux.HandleFunc("GET /id", countRequest("/id", traced("GET /id", handleID)))
	mux.HandleFunc("GET /version", countRequest("/version", traced("GET /version", handleVersion)))
	mux.HandleFunc("GET /metrics", handleMetrics)

	addr := ":8080"
	log.Printf("Notiflex API 시작: %s (pod=%s, tier=%s, version=%s)", addr, podName, tier, version)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("서버 종료: %v", err)
	}
}
