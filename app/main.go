package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/valkey-io/valkey-go"
)

// Notiflex API — 알림 서비스 API 서버.
// /health  : 상태 확인
// /version : 빌드 버전
// /id      : Valkey INCR 로 클러스터 전역 순차 ID 생성 + 처리한 Pod 이름 반환

const version = "v0.3.0"

const idCounterKey = "notiflex:id:counter"

var client valkey.Client

func podName() string {
	if h := os.Getenv("HOSTNAME"); h != "" {
		return h
	}
	h, _ := os.Hostname()
	return h
}

func valkeyPassword() string {
	// ch6.2: 파일 기반 시크릿(CSI) 우선, 없으면 환경변수
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
	// DNS·Valkey 기동 지연 대비 10회 재시도 (3초 간격)
	for i := 0; i < 10; i++ {
		c, err = valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{addr},
			Password:    valkeyPassword(),
		})
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

func main() {
	client = connectValkey()
	defer client.Close()

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
		json.NewEncoder(w).Encode(map[string]any{"id": id, "pod": podName()})
	})

	addr := ":8080"
	log.Printf("notiflex-api %s listening on %s (pod=%s)", version, addr, podName())
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
