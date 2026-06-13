# Notiflex 아키텍처 스냅샷

> 이 문서는 AI가 매 대화에서 전체 그림을 빠르게 잡기 위한 **현재 상태 한눈 보기**다.
> 결정의 "왜"는 `docs/architecture-decisions.md`(ADR), 진행 이력은 `JOURNEY.md`에 둔다.

## 3층 지식 구조

- **CLAUDE.md** — 프로젝트 메타데이터(환경·규칙), 매 대화 자동 로드
- **claude-context/** — 현재 아키텍처 스냅샷(이 문서), 자동 참조
- **docs/architecture-decisions.md** — 결정 누적(ADR), 사람·AI가 함께 검토

세 층이 분리되어야 작업 컨텍스트·현재 그림·과거 결정이 섞이지 않는다.

## 클러스터 토폴로지

| 항목 | 값 |
|------|-----|
| 클러스터 | `notiflex-cluster` (GKE Standard, Zonal) |
| 리전/존 | `asia-northeast3` / `asia-northeast3-a` |
| GKE 기능 | Gateway API(standard), Workload Identity, Secret Manager CSI |

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot) | 2 | ArgoCD, 모니터링 스택 |
| api-pool | e2-medium (Spot) | 1 | notiflex-api(smb), Valkey, Grafana |
| worker-pool | e2-standard-2 (Spot) | 1 | Strimzi Kafka(KRaft) |
| ops-pool | e2-small (Spot) | 1 | Tempo (이후 CronJob 예정) |

## 컴포넌트 다이어그램

```
외부 사용자
   │  http://35.216.45.48
   ▼
Gateway(gke-l7-regional-external-managed) ── HealthCheckPolicy(/health)
   │  HTTPRoute notiflex-route
   ▼
Service notiflex-api ──(Canary)── Service notiflex-api-preview
   │
   ▼
Rollout notiflex-api (Argo Rollouts, Canary 20→50→80→100)
   │  Pod (api-pool)
   ▼
Valkey(INCR) ◀── 비밀번호: GSM ─CSI파일─▶ /mnt/secrets (Workload Identity)
```

## 배포 파이프라인

```
git push (app/**) → GitHub Actions(WIF) → 이미지 빌드 → Artifact Registry
git push (k8s/**) → ArgoCD 감지 → Rollout 동기화(Canary) → 클러스터 반영
```

- 이미지: `asia-northeast3-docker.pkg.dev/project-75fce205-dfa5-4975-a56/notiflex/api`
- GitOps: ArgoCD automated sync(prune + selfHeal)

## 관측 가능성

| 도구 | 역할 |
|------|------|
| Prometheus | 메트릭 수집 |
| Grafana | 메트릭·로그 통합 시각화 |
| Alertmanager + PrometheusRule | 알림(Pod 재시작 과다 등) |
| Loki + Fluent Bit | 로그 수집·조회 |
| Tempo + OpenTelemetry | 분산 트레이싱 (OTLP gRPC 4317) |
| Strimzi Kafka(KRaft) | 이벤트 드리븐 메시징 (notifications 토픽) |

## 주요 네임스페이스

| 네임스페이스 | 주요 워크로드 |
|-------------|-------------|
| notiflex | notiflex-api(Rollout), notiflex-api-preview, Valkey, Gateway/HTTPRoute |
| argocd | ArgoCD (Application: notiflex-smb) |
| monitoring | Prometheus, Grafana, Alertmanager, Loki, Fluent Bit |
| argo-rollouts | Argo Rollouts 컨트롤러 |
