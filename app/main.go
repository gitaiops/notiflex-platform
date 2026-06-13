package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

// Notiflex API — 알림 서비스의 최소 API 서버.
// /health : 상태 확인 (readiness/liveness probe 용)
// /id     : 인메모리 카운터로 순차 ID 생성 + 처리한 Pod 이름 반환

var counter uint64

func podName() string {
	if h := os.Getenv("HOSTNAME"); h != "" {
		return h
	}
	h, _ := os.Hostname()
	return h
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/id", func(w http.ResponseWriter, r *http.Request) {
		id := atomic.AddUint64(&counter, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":  id,
			"pod": podName(),
		})
	})

	addr := ":8080"
	log.Printf("notiflex-api listening on %s (pod=%s)", addr, podName())
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
