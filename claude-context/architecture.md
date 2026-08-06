# Notiflex 아키텍처 스냅샷 (7장 완료 시점)

이 문서는 **지금 어떻게 동작하는가**를 한 페이지로 정리한 것이다. AI가 매 대화에서 전체 그림을
빠르게 잡도록 돕는 것이 목적이다.

## 3층 지식 구조

이 저장소는 AI가 참조하는 지식을 세 층으로 나눠 둔다. 층이 섞이면 오래된 정보가 새 정보를
덮어쓰거나, 결정의 근거가 사라진다.

| 층 | 파일 | 역할 | 갱신 시점 |
|----|------|------|----------|
| 규칙 | `CLAUDE.md` | 항상 지켜야 할 규칙과 환경 정보. 매 대화에 자동으로 읽힌다 | 규칙이나 환경이 바뀔 때 |
| 현재 | `claude-context/architecture.md` | 지금의 아키텍처 스냅샷 | 구성이 바뀔 때마다 |
| 과거 | `docs/architecture-decisions.md` | 결정과 그 이유의 누적 기록 | 결정할 때마다, 지우지 않는다 |

`JOURNEY.md`는 사람이 읽는 진행 기록이고, 도구 선택 기록이 ADR의 초안 역할을 한다.

## 클러스터 토폴로지

| 항목 | 값 |
|------|-----|
| 클러스터 | `notiflex-cluster` (GKE Standard, Zonal) |
| 버전 | 1.35.6-gke.1250000 |
| 리전 / 존 | asia-northeast3 / asia-northeast3-a |
| 노드풀 | `default-pool` e2-medium×2, `api-pool` e2-medium×1, `worker-pool` e2-standard-2×1, `ops-pool` e2-small×1 (모두 Spot) |
| kubectl 컨텍스트 | `gke-sysnet4admin_book_gitaiops` |
| Gateway API | CHANNEL_STANDARD |
| Workload Identity | `project-9d5c279f-44bf-42c9-af2.svc.id.goog` |
| Secret Manager CSI | 활성 (`secrets-store-gke.csi.k8s.io`) |
| GKE 관리형 로깅/모니터링 | 끔. Loki와 Prometheus를 직접 운영하므로 중복이고 노드당 약 130m을 아낀다 |

## 컴포넌트 흐름

```
[인터넷]
    │
    ▼
[Gateway: notiflex-gateway]  35.216.104.31  (gke-l7-regional-external-managed)
    │  HTTPRoute: / → notiflex-api:80
    │  HealthCheckPolicy: /health:8080
    ▼
[Service: notiflex-api]  (stable)        [Service: notiflex-api-preview]  (canary)
    │                                          │
    └──────────────┬───────────────────────────┘
                   ▼
        [Rollout: notiflex-api]  Canary 20 → 50 → 80 → 100%   (api-pool에 배치)
                   │
                   ├── Pod: api (Go, scratch)
                   │     ├── ServiceAccount: notiflex-api (Workload Identity)
                   │     ├── CSI 볼륨 /mnt/secrets/valkey-password
                   │     │     └── Google Secret Manager: valkey-password
                   │     └── /health  /id  /version  /metrics
                   ▼
        [StatefulSet: valkey-primary]  ID 카운터(INCR)를 모든 Pod이 공유
```

## 배포 파이프라인

```
개발자 git push (app/**)
    │
    ▼
GitHub Actions (WIF 키리스 인증)
    ├── docker build → Artifact Registry: notiflex/api:sha-<7자리>
    └── k8s/smb/rollout.yaml 태그 갱신 후 git push
    │
    ▼
ArgoCD (auto-sync, prune, selfHeal, 재조정 주기 3분)
    │
    ▼
Argo Rollouts Canary 진행 → 100% 전환
```

수동 배포는 하지 않는다. 클러스터를 직접 고치면 ArgoCD selfHeal이 Git 상태로 되돌린다.

## 관측 가능성

| 도구 | 역할 | 위치 |
|------|------|------|
| Prometheus | 메트릭 수집. ServiceMonitor로 `/metrics` 수집 | `monitoring` |
| Grafana | 메트릭·로그 조회. Notiflex 대시보드 등록 | `monitoring` |
| Alertmanager | 알림 라우팅. PrometheusRule 2건 | `monitoring` |
| Loki | 로그 저장 (SingleBinary) | `monitoring` |
| Fluent Bit | 노드 로그 수집 후 Loki로 전송 (DaemonSet) | `monitoring` |

앱이 내보내는 지표: `notiflex_http_requests_total{path,pod}`, `notiflex_ids_generated_total{tier}`

## 주요 네임스페이스

| 네임스페이스 | 주요 워크로드 |
|-------------|-------------|
| `notiflex` | SMB 테넌트. notiflex-api(Rollout, api-pool), valkey-primary(StatefulSet), Gateway, HTTPRoute |
| `enterprise` | Enterprise 테넌트. notiflex-api(Rollout, api-pool). Valkey는 notiflex의 것을 함께 쓴다 |
| `argocd` | ArgoCD v3.5.0 (server, repo-server, application-controller 등) |
| `argo-rollouts` | Argo Rollouts 컨트롤러 |
| `monitoring` | Prometheus, Grafana, Alertmanager, Loki, Fluent Bit |
| `kube-system` | GKE 시스템, Secret Manager CSI DaemonSet |

## ArgoCD Application

App of Apps 구조다. `root-app` 하나가 `argocd/apps/` 아래 Application들을 관리한다.

| 이름 | 감시 경로 | 대상 네임스페이스 | sync-wave |
|------|----------|-----------------|-----------|
| `root-app` | `argocd/apps` | `argocd` | — |
| `notiflex-smb` | `k8s/smb` | `notiflex` | 2 |
| `notiflex-enterprise` | `k8s/enterprise` | `enterprise` | 2 |
