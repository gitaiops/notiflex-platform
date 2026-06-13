# 멀티스테이지 빌드: golang:1.25-alpine 에서 정적 바이너리 → scratch 로 복사
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notiflex-api .

FROM scratch
COPY --from=build /notiflex-api /notiflex-api
EXPOSE 8080
ENTRYPOINT ["/notiflex-api"]
