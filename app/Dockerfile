# 빌드 단계: 정적 바이너리를 만든다.
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notiflex .

# 실행 단계: 바이너리 하나만 담는다.
FROM scratch

COPY --from=builder /notiflex /notiflex

EXPOSE 8080
ENTRYPOINT ["/notiflex"]
