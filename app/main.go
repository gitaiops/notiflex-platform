// Notiflex API — B2B 알림 SaaS 플랫폼의 API 서버.
// 외부 프레임워크 없이 Go 표준 라이브러리만 사용한다.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
)

// counter는 /id 요청마다 증가하는 인메모리 카운터다.
// Pod을 재시작하면 0부터 다시 시작한다.
var counter atomic.Uint64

// podName은 응답이 어느 Pod에서 나왔는지 구분하기 위해 사용한다.
var podName = hostname()

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

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("응답 인코딩 실패: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"status": "ok",
		"pod":    podName,
	})
}

func handleID(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"id":           strconv.FormatUint(counter.Add(1), 10),
		"generated_by": podName,
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /id", handleID)

	addr := ":8080"
	log.Printf("Notiflex API 시작: %s (pod=%s)", addr, podName)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("서버 종료: %v", err)
	}
}
