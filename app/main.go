// Notiflex API — B2B 알림 SaaS 플랫폼의 API 서버.
// 외부 프레임워크 없이 Go 표준 라이브러리만 사용한다.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

// counter는 /id 요청마다 증가하는 인메모리 카운터다.
// Pod을 재시작하면 0부터 다시 시작한다.
var counter atomic.Uint64

// podName은 응답이 어느 Pod에서 나왔는지 구분하기 위해 사용한다.
var podName = hostname()

// version은 배포된 코드의 버전이다. 빌드할 때 ldflags로 덮어쓸 수 있다.
var version = "0.2.0"

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
	fmt.Fprintln(w, "# HELP notiflex_ids_generated_total 발급한 ID 총 개수")
	fmt.Fprintln(w, "# TYPE notiflex_ids_generated_total counter")
	fmt.Fprintf(w, "notiflex_ids_generated_total{pod=%q} %d\n", podName, counter.Load())
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
	})
}

func handleID(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"id":           strconv.FormatUint(counter.Add(1), 10),
		"generated_by": podName,
	})
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"version": version,
		"pod":     podName,
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", countRequest("/health", handleHealth))
	mux.HandleFunc("GET /id", countRequest("/id", handleID))
	mux.HandleFunc("GET /version", countRequest("/version", handleVersion))
	mux.HandleFunc("GET /metrics", handleMetrics)

	addr := ":8080"
	log.Printf("Notiflex API 시작: %s (pod=%s)", addr, podName)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("서버 종료: %v", err)
	}
}
