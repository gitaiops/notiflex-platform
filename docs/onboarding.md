# Notiflex Platform — 온보딩 가이드

이 문서는 `notiflex-platform`에 새로 합류하는 엔지니어가 첫 주에 필요한 내용을 정리한 것이다.
"왜 이렇게 결정했는지"는 `docs/architecture-decisions.md`(ADR), "어떻게 여기까지 왔는지"는
`JOURNEY.md`를 참조한다. 이 문서는 그 둘을 요약한 입구다.

## 1. 한눈에 보기

Notiflex는 알림 API를 GitOps로 운영하는 멀티테넌트(SMB/Enterprise) 플랫폼이다.

| 항목 | 값 |
|---|---|
| GCP 프로젝트 | `project-75fce205-dfa5-4975-a56` |
| 리전 / 존 | `asia-northeast3` / `asia-northeast3-a` |
| GKE 클러스터 | `notiflex-cluster` (Standard, Spot, Gateway API) |
| kubectl 컨텍스트 | `gke-notiflex` (모든 명령에 `--context gke-notiflex` 명시) |
| Artifact Registry | `asia-northeast3-docker.pkg.dev/project-75fce205-dfa5-4975-a56/notiflex` |
| 외부 진입점 | Gateway API, 현재 IP는 `claude-context/architecture.md` 참조 |

## 2. 저장소 구조

```
app/                       Notiflex API (Go, net/http)
  main.go                  /health, /version, /id 핸들러
  Dockerfile                multistage: golang:1.25-alpine → scratch

k8s/smb/                   SMB 티어 매니페스트
  namespace.yaml
  rollout.yaml             Argo Rollouts (Canary)
  service.yaml / service-preview.yaml
  gateway.yaml             Gateway + HTTPRoute
  healthcheckpolicy.yaml
  secret-provider.yaml     SecretProviderClass (GSM CSI)
  healthcheck-cronjob.yaml 5분 주기 헬스체크 배치

k8s/enterprise/            Enterprise 티어 매니페스트 (SMB와 동일 패턴, 별도 namespace)
k8s/kafka/                 Strimzi KafkaNodePool/Kafka/KafkaTopic
k8s/monitoring/            PrometheusRule 등 알림 정의

helm-values/                서드파티 차트 values (kube-prometheus, loki, fluent-bit, strimzi, tempo)
argocd/
  root-app.yaml             App of Apps 루트, argocd/apps를 recurse 감시
  apps/notiflex-smb.yaml
  apps/notiflex-enterprise.yaml

.github/workflows/ci.yaml   GitHub Actions CI (WIF 키리스 인증)

CLAUDE.md                   환경 메타데이터, AI 작업 컨텍스트
JOURNEY.md                  챕터별 진행 이력 + 도구 선택 기록 + 버전 표
claude-context/architecture.md   현재 아키텍처 스냅샷
docs/architecture-decisions.md   ADR (결정의 "왜")
```

## 3. 배포 방법

매니페스트 변경은 **항상 Git을 거쳐서만** 클러스터에 반영한다. `kubectl apply`로 직접 바꾸지 않는다
(CronJob 최초 적용처럼 ArgoCD 인식 지연 시 임시로 직접 적용하는 예외가 있으나, 곧바로 Git에도 반영해야 한다).

```
git push (app/**)  → GitHub Actions(WIF) → 이미지 빌드 → Artifact Registry
git push (k8s/**)  → ArgoCD 감지          → Rollout 동기화(Canary) → 클러스터 반영
```

- GitHub Actions는 **이미지만** 빌드/푸시한다. 조직 정책상 워크플로 쓰기 권한이 없어 매니페스트 자동 커밋은 하지 않는다 — 이미지 태그를 manifest에 반영하는 커밋은 사람(또는 Claude)이 수동으로 한다.
- 이미지 태그 규칙: `api:vMAJOR.MINOR.PATCH`.
- ArgoCD는 `automated sync(prune + selfHeal)`이므로 클러스터를 직접 고치면 곧 되돌아간다 — 항상 Git을 먼저 고친다.
- root-app(App of Apps)이 `argocd/apps`를 재귀 감시하므로 새 Application은 디렉터리에 파일 추가만으로 등록된다.

## 4. 주요 아키텍처 결정 (요약)

전체 근거는 `docs/architecture-decisions.md`(ADR-001~014)에 있다. 핵심만 추리면:

| 영역 | 선택 | 이유 (한 줄) |
|---|---|---|
| GitOps | ArgoCD | Git이 SSOT, selfHeal/prune, App of Apps로 멀티테넌시 확장 |
| CI | GitHub Actions + WIF | 키리스 인증, SA 키 생성이 막힌 조직 정책과 부합 |
| 메트릭 | Prometheus + Grafana | 오픈소스, kube-prometheus-stack 일괄 설치 |
| 로깅 | Loki + Fluent Bit | 레이블 기반 경량, Grafana 통합 |
| 외부 노출 | Gateway API | 인프라/라우팅 역할 분리, GKE 네이티브 L7 |
| 배포 전략 | Canary (Blue/Green → 전환) | 점진적 트래픽 전환으로 결함 영향 범위 단계적 제한 |
| 캐시 | Valkey | Redis 호환, 오픈 거버넌스, INCR로 원자적 ID 생성 |
| 시크릿 | GKE Secret Manager CSI + Workload Identity | 키리스, 파일 마운트, GCP 네이티브 |
| 노드풀 | 역할별 분리 (api/worker/ops) | 워크로드 격리, 머신타입 최적화 |
| 멀티테넌시 | Namespace 분리 + 테넌트별 Rollout | 강한 격리, App of Apps와 자연 결합 |
| 메시징 | Strimzi Kafka (KRaft) | 이벤트 드리븐, ZooKeeper 불필요 |
| 트레이싱 | Tempo + OpenTelemetry | OTLP 표준, Grafana 통합 |
| 배치 자동화 | K8s CronJob | 쿠버네티스 네이티브, ArgoCD가 매니페스트로 관리 |

신규 결정을 추가할 때는 **검토한 대안**도 함께 ADR에 남긴다. "왜 이것을 안 썼는지"가 "왜 이것을 썼는지"만큼 중요하다.

## 5. 관측 가능성 스택

| 도구 | 역할 | 접근 방법 |
|---|---|---|
| Prometheus | 메트릭 수집 | Grafana 데이터소스로 조회 |
| Grafana | 메트릭/로그/트레이스 통합 시각화 | `kubectl port-forward -n monitoring svc/kube-prometheus-grafana 3000:80` |
| Alertmanager + PrometheusRule | 알림 (예: Pod 재시작 과다) | `k8s/monitoring/pod-restart-alert.yaml` |
| Loki + Fluent Bit | 로그 수집/조회 | Grafana Explore에서 LogQL |
| Tempo + OpenTelemetry | 분산 트레이싱 (OTLP gRPC 4317) | Grafana Explore에서 TraceQL, 앱은 `OTEL_EXPORTER_OTLP_ENDPOINT`로 전송 |

애플리케이션 쪽 계측은 `app/main.go`의 `initTracer()`를 참고 — `/id` 핸들러가 `generate-id` 스팬을 생성하고
`notiflex.id`, `notiflex.pod` 속성을 붙인다.

## 6. 로컬에서 첫 확인 (Day 1 체크리스트)

```bash
# 1. kubectl 컨텍스트 확인
kubectl config use-context gke-notiflex

# 2. ArgoCD 상태 확인 — 전부 Synced/Healthy 여야 정상
kubectl get application -n argocd

# 3. 핵심 워크로드 확인
kubectl get pods -n notiflex
kubectl get pods -n enterprise
kubectl get rollout -n notiflex

# 4. 외부 접근 확인
GATEWAY_IP=$(kubectl get gateway notiflex-gateway -n notiflex -o jsonpath='{.status.addresses[0].value}')
curl http://$GATEWAY_IP/health
curl http://$GATEWAY_IP/version

# 5. CronJob 동작 확인
kubectl get cronjob -n notiflex
kubectl get jobs -n notiflex
```

## 7. 트러블슈팅 가이드

`JOURNEY.md`의 트러블슈팅 이력 테이블이 비어 있다면 아직 기록된 사례가 없다는 뜻이다 — 새로 겪는 문제는
반드시 그 테이블에 추가해서 같은 문제를 반복하지 않게 한다. 지금까지 실제로 겪은 사례(구축 과정 기록):

| 증상 | 원인 | 해결 |
|---|---|---|
| Loki Pod CrashLoopBackOff (`mkdir /var/loki: read-only file system`) | 기본 values가 root 파일시스템에 쓰기 시도 | `helm-values/loki.yaml`에 `extraVolumes`(emptyDir) + `extraVolumeMounts`를 `/var/loki`로 추가 |
| Tempo Pod Pending (ops-pool) | e2-small 노드 메모리 요청이 포화 (다른 Pod가 ops-pool로 드리프트) | 드리프트한 Pod를 삭제해 올바른 노드풀로 재스케줄 유도 |
| Strimzi Kafka 매니페스트 "no matches for kind" | `kafka.strimzi.io/v1beta2` 사용 — Strimzi 1.0.0은 `v1`만 서빙 | `kubectl get crd <name> -o jsonpath=...spec.versions[*].name`로 실제 서빙 버전 확인 후 매니페스트 수정 |
| ArgoCD가 새 매니페스트(CronJob 등)를 즉시 인식하지 못함 | ArgoCD 파일 감지 주기(기본 3분), sync 호출이 "already synced" 반환 | `kubectl apply`로 직접 적용 → Git에도 반드시 동일 내용 커밋 → ArgoCD가 곧 자동 인식 (또는 hard refresh annotation) |
| GKE Gateway 리소스 누수 (클러스터 삭제 후) | forwarding-rule/proxy/url-map/backend-service/주소가 클러스터 삭제로 자동 정리되지 않음 | 의존성 역순으로 직접 삭제: forwarding-rules → target-http-proxies → url-maps → backend-services → addresses → orphaned PVC 디스크 |

## 8. 자주 헷갈리는 부분

- **nodeSelector 키는 항상 `cloud.google.com/gke-nodepool`** (GKE 자동 라벨). 커스텀 라벨/taint를 쓰지 않는다 — ADR-010 참조.
- **enterprise 네임스페이스는 자체 Valkey가 없다.** `notiflex` 네임스페이스의 `valkey-primary`를 cross-namespace DNS(`valkey-primary.notiflex.svc.cluster.local`)로 공유한다.
- **시크릿은 K8s Secret이 아니라 GSM + CSI 파일 마운트**다. `VALKEY_PASSWORD_FILE` 환경변수가 가리키는 파일을 읽는 식이지, 환경변수에 비밀번호 값이 직접 들어가지 않는다.
- **CI가 매니페스트를 자동으로 고치지 않는다.** 이미지가 빌드된 뒤 `k8s/**`의 이미지 태그를 갱신하는 커밋은 별도로 필요하다.

## 9. 더 읽을 것

- `JOURNEY.md` — 챕터별 진행 이력, 도구 선택 시 검토했던 대안, 현재 버전 표
- `docs/architecture-decisions.md` — 전체 ADR
- `claude-context/architecture.md` — 현재 아키텍처 스냅샷 (다이어그램 포함)
