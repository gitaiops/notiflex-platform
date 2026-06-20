# Notiflex 여정 기록

이 파일은 독자가 실제로 진행한 내용을 기록한다. AI가 각 챕터 완료 시 자동으로 업데이트한다.

## 진행 현황

| 챕터 | 서브챕터 | 상태 | 완료일 | 비고 |
|------|---------|------|--------|------|
| ch2 | 2.2 설치 확인 | ✅ | 2026-06-13 | Claude Code 설치 확인 |
| ch2 | 2.3 gcloud 설정 | ✅ | 2026-06-13 | project, asia-northeast3, AR 인증 |
| ch2 | 2.4 GitHub 저장소 | ✅ | 2026-06-13 | gitaiops/notiflex-platform |
| ch2 | 2.5 GKE 클러스터 | ✅ | 2026-06-13 | notiflex-cluster (Spot, Gateway API) |
| ch2 | 2.6 빌드/배포 | ✅ | 2026-06-13 | api:v0.1.0, Deployment replicas 2 |
| ch2 | 2.7 첫 커밋 | ✅ | 2026-06-13 | Initial commit |
| ch3 | 3.2 GitOps 도구 | ✅ | 2026-06-13 | ArgoCD v3.4.3, notiflex-smb auto-sync |
| ch3 | 3.3 기능 추가 | ✅ | 2026-06-13 | /version 추가, v0.1.1 롤링 업데이트 |
| ch3 | 3.4 CI | ✅ | 2026-06-13 | GitHub Actions + WIF 키리스 인증 |
| ch3 | 3.5 CI-CD 연결 | ✅ | 2026-06-13 | CI(코드→이미지) + GitOps 커밋(매니페스트→배포) |
| ch4 | 4.2 메트릭 모니터링 | ✅ | 2026-06-13 | kube-prometheus-stack (Prometheus/Grafana/Alertmanager) |
| ch4 | 4.3 로그 수집 | ✅ | 2026-06-13 | Loki SingleBinary + Fluent Bit |
| ch4 | 4.4 알림 | ✅ | 2026-06-13 | PrometheusRule pod-restart-alert |
| ch5 | 5.2 트래픽 관리 | ✅ | 2026-06-13 | Gateway API + HTTPRoute + HealthCheckPolicy (IP 35.216.23.130) |
| ch5 | 5.3 무중단 배포 | ✅ | 2026-06-13 | Argo Rollouts Blue/Green, v0.2.0 auto-promote |
| ch6 | 6.1 캐시 | ✅ | 2026-06-13 | Valkey INCR, api:v0.3.0 |
| ch6 | 6.2 시크릿 관리 | ✅ | 2026-06-13 | GKE Secret Manager CSI + Workload Identity |
| ch6 | 6.3 Canary 전환 | ✅ | 2026-06-13 | Blue/Green → Canary(20/50/80%) |
| ch7 | 7.2 멀티 노드풀 | ✅ | 2026-06-13 | api/worker/ops-pool (GKE_METADATA) |
| ch7 | 7.3 App of Apps | ✅ | 2026-06-13 | root-app → smb/enterprise (sync-wave) |
| ch7 | 7.4 멀티테넌시 | ✅ | 2026-06-13 | enterprise 네임스페이스, cross-ns Valkey 공유 |
| ch8 | 8.1 메시징 | ✅ | 2026-06-13 | Strimzi Kafka(KRaft) + notifications 토픽, Producer/Consumer |
| ch8 | 8.2 트레이싱 | ✅ | 2026-06-13 | Tempo + OTel SDK, /id span 수집 확인 |
| ch8 | 8.3 CronJob | ✅ | 2026-06-20 | healthcheck-cronjob, 5분 주기, ops-pool 배치 |
| ch9 | 9.1 저장소 분석 | ⬜ | | |
| ch9 | 9.2 회고 | ⬜ | | |
| ch9 | 9.3 온보딩 문서 | ⬜ | | |
| ch9 | 9.4 GitAIOps 분석 | ⬜ | | |
| ch9 | 9.5 마무리 | ⬜ | | |

## 도구 선택 기록

독자가 3-프롬프트 패턴(탐색→비교→실행)에서 실제로 선택한 도구와 이유를 기록한다.

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-----------|----------|
| GitOps (ch3.2) | ArgoCD | Flux | UI·App of Apps·selfHeal, 멀티테넌시 확장 용이 |
| CI (ch3.4) | GitHub Actions | Jenkins, GitLab CI | 저장소 통합, WIF 키리스 인증 |
| 메트릭 (ch4.2) | Prometheus+Grafana | Datadog, New Relic | 오픈소스, kube-prometheus-stack 일괄 설치 |
| 로깅 (ch4.3) | Loki+Fluent Bit | ELK | 경량, Grafana 통합 조회 |
| 외부 노출 (ch5.2) | Gateway API | Ingress(NGINX), Istio | 역할 분리, K8s 표준, GKE 네이티브 L7 |
| 무중단 배포 (ch5.3) | Argo Rollouts Blue/Green | Flagger, Deployment RollingUpdate | preview 검증 후 일괄 전환, argoproj 통합 |
| 캐시 (ch6.1) | Valkey | Redis, Memcached | Redis 호환·오픈 거버넌스, 경량 standalone |
| 시크릿 (ch6.2) | GKE Secret Manager CSI + WI | K8s Secret, 외부 Vault | 키리스(WI), 파일 마운트, GCP 네이티브 |
| 배포 전략 (ch6.3) | Canary | Blue/Green 유지 | 점진 트래픽 전환으로 위험 분산 |
| 멀티 노드풀 (ch7.2) | 역할별 노드풀 | 단일 노드풀 | 워크로드 격리(api/worker/ops), 리소스 예측성 |
| App of Apps (ch7.3) | root-app 패턴 | Application 개별 관리 | 선언적 일괄 관리, sync-wave 순서 제어 |
| 멀티테넌시 (ch7.4) | Namespace 분리 + per-tenant Rollout | 단일 namespace + 라벨 격리, vCluster | 강한 격리, App of Apps와 자연 결합, 테넌트별 독립 배포 |
| 메시징 (ch8.1) | Strimzi Kafka (KRaft) | RabbitMQ, Pub/Sub | 이벤트 드리븐 표준, K8s 네이티브 운영, ZooKeeper 불필요 |
| 트레이싱 (ch8.2) | Tempo + OpenTelemetry | Jaeger, Zipkin | Grafana 통합, OTLP 표준, 경량 monolithic |
| 배치 자동화 (ch8.3) | K8s CronJob | 외부 cron + 쿠버네티스 외부 트리거, Argo Workflows | 쿠버네티스 네이티브, ops-pool 배치, ArgoCD가 매니페스트로 관리 |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25 | ch2 |
| Notiflex 이미지 | api:v0.5.0 | … v0.3.0(Valkey) → v0.4.0(Kafka) → v0.5.0(OTel) |
| Argo Rollouts | v1.8.3 | ch5.3 |
| Valkey | bitnami standalone | ch6.1 |
| Kafka | 4.1.0 (Strimzi 1.0.0, KRaft) | ch8.1 |
| OTel SDK | 1.43.0 (Tempo) | ch8.2 |
| ArgoCD | v3.4.3 | ch3.2 |
| Kafka | | |
| OTel SDK | | |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot) | 2 | ArgoCD, 모니터링 |
| api-pool | e2-medium (Spot) | 1 | notiflex-api, Valkey |
| worker-pool | e2-standard-2 (Spot) | 1 | (ch8 Kafka) |
| ops-pool | e2-small (Spot) | 1 | (ch8 Tempo/CronJob) |

## 트러블슈팅 이력

독자가 겪은 문제와 해결 방법을 기록한다. 같은 문제를 다시 겪지 않도록 한다.

| 챕터 | 문제 | 해결 |
|------|------|------|
| | | |
