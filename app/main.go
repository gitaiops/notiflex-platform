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

	"github.com/valkey-io/valkey-go"
)

// idCounterKey는 Valkey에서 ID 카운터를 저장하는 키다.
// 인메모리 카운터와 달리 모든 Pod이 같은 값을 공유한다.
const idCounterKey = "notiflex:id:counter"

// valkeyClient는 ID 발급에 쓰는 Valkey 연결이다.
var valkeyClient valkey.Client

// podName은 응답이 어느 Pod에서 나왔는지 구분하기 위해 사용한다.
var podName = hostname()

// tier는 이 인스턴스가 어느 고객 등급을 담당하는지 나타낸다.
// SMB와 Enterprise를 별도 네임스페이스에서 운영한다.
var tier = envOr("NOTIFLEX_TIER", "smb")

// version은 배포된 코드의 버전이다. 빌드할 때 ldflags로 덮어쓸 수 있다.
var version = "0.4.0"

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

	// INCR은 원자적이라 여러 Pod이 동시에 호출해도 번호가 겹치지 않는다.
	n, err := valkeyClient.Do(ctx, valkeyClient.B().Incr().Key(idCounterKey).Build()).AsInt64()
	if err != nil {
		log.Printf("ID 발급 실패: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, map[string]string{"error": "ID를 발급하지 못했다"})
		return
	}

	writeJSON(w, map[string]string{
		"id":           strconv.FormatInt(n, 10),
		"generated_by": podName,
		"tier":         tier,
		"source":       "valkey",
	})
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"version": version,
		"pod":     podName,
	})
}

func main() {
	client, err := connectValkey()
	if err != nil {
		log.Fatalf("Valkey 연결 실패: %v", err)
	}
	valkeyClient = client
	defer valkeyClient.Close()
	log.Printf("Valkey 연결 완료")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", countRequest("/health", handleHealth))
	mux.HandleFunc("GET /id", countRequest("/id", handleID))
	mux.HandleFunc("GET /version", countRequest("/version", handleVersion))
	mux.HandleFunc("GET /metrics", handleMetrics)

	addr := ":8080"
	log.Printf("Notiflex API 시작: %s (pod=%s, tier=%s, version=%s)", addr, podName, tier, version)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("서버 종료: %v", err)
	}
}
