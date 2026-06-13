package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/valkey-io/valkey-go"
)

// Notiflex API — 알림 서비스 API 서버.
// /health  : 상태 확인
// /version : 빌드 버전
// /id      : Valkey INCR 로 전역 순차 ID 생성 + Kafka notifications 토픽으로 이벤트 발행

const version = "v0.4.0"

const (
	idCounterKey = "notiflex:id:counter"
	kafkaTopic   = "notifications"
)

var (
	client   valkey.Client
	producer sarama.SyncProducer
)

func podName() string {
	if h := os.Getenv("HOSTNAME"); h != "" {
		return h
	}
	h, _ := os.Hostname()
	return h
}

func valkeyPassword() string {
	if f := os.Getenv("VALKEY_PASSWORD_FILE"); f != "" {
		if data, err := os.ReadFile(f); err == nil {
			return string(data)
		}
	}
	return os.Getenv("VALKEY_PASSWORD")
}

func connectValkey() valkey.Client {
	addr := os.Getenv("VALKEY_ADDR")
	if addr == "" {
		addr = "valkey-primary.notiflex.svc.cluster.local:6379"
	}
	var c valkey.Client
	var err error
	for i := 0; i < 10; i++ {
		c, err = valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}, Password: valkeyPassword()})
		if err == nil {
			if e := c.Do(context.Background(), c.B().Ping().Build()).Error(); e == nil {
				return c
			} else {
				err = e
				c.Close()
			}
		}
		log.Printf("Valkey 연결 재시도 %d/10: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("Valkey 연결 실패: %v", err)
	return nil
}

func kafkaBrokers() []string {
	b := os.Getenv("KAFKA_BROKERS")
	if b == "" {
		return nil
	}
	return strings.Split(b, ",")
}

// 이벤트 드리븐: notifications 토픽 Producer (재시도 포함)
func connectProducer(brokers []string) sarama.SyncProducer {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_1_0_0
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true
	var p sarama.SyncProducer
	var err error
	for i := 0; i < 10; i++ {
		p, err = sarama.NewSyncProducer(brokers, cfg)
		if err == nil {
			return p
		}
		log.Printf("Kafka Producer 연결 재시도 %d/10: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	log.Fatalf("Kafka Producer 연결 실패: %v", err)
	return nil
}

// 이벤트 드리븐: notifications 토픽 Consumer (백그라운드, 로그 출력)
func startConsumer(brokers []string) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_1_0_0
	consumer, err := sarama.NewConsumer(brokers, cfg)
	if err != nil {
		log.Printf("Kafka Consumer 생성 실패(스킵): %v", err)
		return
	}
	pc, err := consumer.ConsumePartition(kafkaTopic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("Kafka 파티션 구독 실패(스킵): %v", err)
		return
	}
	go func() {
		for msg := range pc.Messages() {
			log.Printf("[consumer] notifications: %s", string(msg.Value))
		}
	}()
}

func main() {
	client = connectValkey()
	defer client.Close()

	brokers := kafkaBrokers()
	if len(brokers) > 0 {
		producer = connectProducer(brokers)
		defer producer.Close()
		startConsumer(brokers)
		log.Printf("Kafka 연결됨 (brokers=%v, topic=%s)", brokers, kafkaTopic)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"version": version, "pod": podName()})
	})

	mux.HandleFunc("/id", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		id, err := client.Do(ctx, client.B().Incr().Key(idCounterKey).Build()).AsInt64()
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		// 이벤트 발행 (best-effort)
		if producer != nil {
			payload := fmt.Sprintf(`{"id":%d,"pod":%q,"ts":%q}`, id, podName(), time.Now().UTC().Format(time.RFC3339))
			if _, _, perr := producer.SendMessage(&sarama.ProducerMessage{
				Topic: kafkaTopic,
				Value: sarama.StringEncoder(payload),
			}); perr != nil {
				log.Printf("[producer] 발행 실패: %v", perr)
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"id": id, "pod": podName()})
	})

	addr := ":8080"
	log.Printf("notiflex-api %s listening on %s (pod=%s)", version, addr, podName())
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
