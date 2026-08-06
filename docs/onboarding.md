# Notiflex 온보딩

새로 합류한 엔지니어가 이 플랫폼을 파악하고 첫 작업을 하기까지 필요한 내용을 모았다.
아키텍처 전체 그림은 `claude-context/architecture.md`, 결정의 이유는 `docs/architecture-decisions.md`를 본다.

## 0. 먼저 알아둘 것

이 저장소는 **Git이 유일한 진실**이다. 클러스터를 `kubectl apply`나 `kubectl edit`으로 직접
고치면 ArgoCD가 몇 분 안에 Git 상태로 되돌린다. 무엇을 바꾸든 매니페스트를 고치고 커밋한다.

여기서 예외인 것은 두 가지뿐이다. 조사를 위한 읽기 명령, 그리고 Secret Manager의 시크릿 값
갱신이다. 시크릿 값은 원래 Git에 없다.

## 1. Day 0 준비

### 필요한 도구

| 도구 | 용도 | 확인 |
|------|------|------|
| `gcloud` | 클러스터 인증, 시크릿, 노드풀 | `gcloud version` |
| `kubectl` | 클러스터 조회 | `kubectl version --client` |
| `kubectl argo rollouts` 플러그인 | Canary 진행 상황 확인 | `kubectl argo rollouts version` |
| Go 1.25 이상 | `app/` 빌드 | `go version` |
| Docker | 이미지 로컬 빌드, 로컬 Valkey | `docker version` |
| `jq` | 알림, JSON 응답 확인 | `jq --version` |

Rollouts 플러그인이 없으면 이렇게 넣는다.

```bash
gcloud components install kubectl
brew install argoproj/tap/kubectl-argo-rollouts   # macOS
```

### 접근 준비

```bash
gcloud auth login
gcloud config set project project-9d5c279f-44bf-42c9-af2
gcloud container clusters get-credentials notiflex-cluster --zone=asia-northeast3-a
kubectl config rename-context "$(kubectl config current-context)" gke-sysnet4admin_book_gitaiops
```

모든 kubectl 명령에 `--context gke-sysnet4admin_book_gitaiops`를 붙인다. 다른 클러스터에
명령이 들어가는 사고를 막기 위한 규칙이다.

잘 붙었는지 확인한다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get nodes
kubectl --context gke-sysnet4admin_book_gitaiops get app -n argocd
```

Application 4개가 모두 `Synced` / `Healthy`면 정상이다.

## 2. 클러스터 구성

| 노드풀 | 머신 타입 | 대수 | 올라가는 것 |
|--------|----------|------|-----------|
| `default-pool` | e2-medium (Spot) | 2 | ArgoCD, Prometheus, Grafana, Loki, Valkey |
| `api-pool` | e2-medium (Spot) | 1 | notiflex-api (SMB, Enterprise) |
| `worker-pool` | e2-standard-2 (Spot) | 1 | Strimzi, Kafka |
| `ops-pool` | e2-small (Spot) | 1 | Tempo, 헬스체크 CronJob |

| 네임스페이스 | 역할 |
|-------------|------|
| `notiflex` | SMB 테넌트. API, Valkey, Gateway, CronJob |
| `enterprise` | Enterprise 테넌트. API |
| `argocd` | ArgoCD |
| `argo-rollouts` | Canary 배포 컨트롤러 |
| `monitoring` | Prometheus, Grafana, Alertmanager, Loki, Fluent Bit, Tempo |
| `kafka` | Strimzi Operator, Kafka 브로커 |

노드는 전부 Spot VM이다. Pod이 갑자기 죽어 있으면 코드 문제가 아니라 Spot 회수일 수 있다.
`kubectl get events`로 `Preempted`가 있는지 먼저 본다.

## 3. 저장소 구조

```
app/                  Notiflex API (Go, 표준 라이브러리만)
k8s/smb/              SMB 테넌트 매니페스트 + Gateway + CronJob
k8s/enterprise/       Enterprise 테넌트 매니페스트
k8s/kafka/            Kafka 클러스터와 토픽
k8s/monitoring/       대시보드, 데이터소스, 알림 규칙
helm-values/          서드파티 차트 values (값을 여기에 고정한다)
argocd/root-app.yaml  App of Apps 루트
argocd/apps/          각 Application 정의
docs/                 ADR과 이 문서
claude-context/       현재 아키텍처 스냅샷
command-guardrails/   위험 작업 절차서
.claude/commands/     /update-docs 커스텀 명령
JOURNEY.md            진행 이력과 도구 선택 기록
```

## 4. 애플리케이션

`app/main.go` 한 파일이다. 외부 웹 프레임워크가 없고 Go 표준 라이브러리의 `net/http`만 쓴다.

### 엔드포인트

| 경로 | 하는 일 |
|------|--------|
| `GET /health` | 상태, Pod 이름, 버전, 티어 반환. readiness와 liveness 프로브가 함께 쓴다 |
| `GET /id` | Valkey `INCR`로 클러스터 전역 ID를 발급하고 Kafka에 알림 이벤트를 보낸다 |
| `GET /version` | 배포된 코드 버전과 배포 전략 반환 |
| `GET /metrics` | Prometheus 노출 형식. 클라이언트 라이브러리 없이 직접 출력한다 |

### 환경변수

| 변수 | 없을 때 동작 |
|------|-------------|
| `VALKEY_ADDR` | 클러스터 내부 DNS로 기본값 사용 |
| `VALKEY_PASSWORD_FILE` | `VALKEY_PASSWORD` 환경변수로 넘어간다 |
| `KAFKA_BROKER` | 이벤트 발행을 건너뛰고 계속 동작한다 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | 트레이싱 없이 계속 동작한다 |
| `NOTIFLEX_TIER` | `smb` |
| `POD_NAME` | OS 호스트명 |

Kafka와 트레이싱은 없어도 동작하지만 **Valkey는 필수다**. 연결에 실패하면 3초 간격으로 10번
재시도한 뒤 프로세스가 종료된다. 로컬에서 `go run`이 30초쯤 뒤에 죽는다면 대개 이 경우다.

### 로컬 실행

가장 간단한 방법은 Valkey를 로컬에 띄우는 것이다.

```bash
docker run -d --name valkey-local -p 6379:6379 valkey/valkey:8-alpine
cd app
VALKEY_ADDR=127.0.0.1:6379 NOTIFLEX_TIER=smb go run .
```

다른 터미널에서 확인한다.

```bash
curl -s localhost:8080/health | jq
curl -s localhost:8080/id | jq
curl -s localhost:8080/metrics
```

클러스터의 Valkey를 그대로 쓰고 싶으면 port-forward하고 비밀번호를 Secret Manager에서 가져온다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops port-forward svc/valkey-primary -n notiflex 6379:6379
VALKEY_ADDR=127.0.0.1:6379 \
VALKEY_PASSWORD="$(gcloud secrets versions access latest --secret=valkey-password \
  --project=project-9d5c279f-44bf-42c9-af2)" go run .
```

이 경우 로컬에서 발급한 ID가 운영 카운터를 올린다. 카운터를 건드리고 싶지 않으면 첫 번째
방법을 쓴다.

커밋 전에는 이 정도만 돌린다. 지금 테스트 코드는 없다.

```bash
cd app && go vet ./... && go build ./...
```

## 5. 접속 방법

### API

```bash
curl http://35.216.104.31/health
curl http://35.216.104.31/id
curl http://35.216.104.31/version
```

Gateway IP가 바뀌었으면 이렇게 확인한다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get gateway notiflex-gateway -n notiflex \
  -o jsonpath='{.status.addresses[0].value}'
```

### ArgoCD UI

```bash
kubectl --context gke-sysnet4admin_book_gitaiops -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d; echo
kubectl --context gke-sysnet4admin_book_gitaiops port-forward svc/argocd-server -n argocd 8443:443
```

`https://localhost:8443`, 계정은 `admin`.

### Grafana

```bash
kubectl --context gke-sysnet4admin_book_gitaiops port-forward svc/kube-prometheus-grafana -n monitoring 3000:80
```

`http://localhost:3000`, 계정은 `admin` / `notiflex-admin`.

데이터소스는 셋이다.

- **Prometheus**: 메트릭. 대시보드 "Notiflex API"에 요청 수, ID 발급 수, CPU, 메모리, 재시작 횟수가 있다
- **Loki**: 로그. Explore에서 `{namespace="notiflex"}`
- **Tempo**: 트레이스. Explore에서 Search로 서비스 `notiflex-api` 조회

## 6. 배포 흐름

```
app/ 코드 수정 → git push
    → GitHub Actions가 이미지 빌드 (Workload Identity, 키 파일 없음)
    → Artifact Registry에 notiflex/api:sha-<커밋 7자리> 푸시
    → CI가 k8s/smb/rollout.yaml 태그를 갱신하고 다시 push
    → ArgoCD가 감지 (최대 3분)
    → Argo Rollouts Canary 20% → 50% → 80% → 100% (단계마다 30초)
```

여기서 새로 온 사람이 자주 놓치는 것이 둘 있다.

**CI는 `app/**`와 CI 파일이 바뀔 때만 돈다.** 매니페스트만 고쳤다면 빌드는 없고 ArgoCD가
바로 반영한다.

**CI는 두 테넌트의 Rollout을 함께 갱신한다.** `k8s/smb/rollout.yaml`과
`k8s/enterprise/rollout.yaml`이 같은 커밋에서 같은 태그로 올라간다. 테넌트마다 배포 시점을
다르게 가져가고 싶다면 `.github/workflows/ci.yaml`의 `TARGETS`에서 대상을 줄인다.
지금 두 태그가 같은지는 이렇게 본다.

```bash
grep -h "image:" k8s/smb/rollout.yaml k8s/enterprise/rollout.yaml
```

진행 상황은 이렇게 본다.

```bash
kubectl argo rollouts get rollout notiflex-api -n notiflex --context gke-sysnet4admin_book_gitaiops -w
```

## 7. 이 저장소에서 Claude Code와 일하는 방식

이 저장소는 코드와 매니페스트만이 아니라 결정의 이유까지 함께 담는다. 지식을 세 층으로 나눈다.

| 층 | 파일 | 무엇을 담는가 | 언제 고치는가 |
|----|------|-------------|-------------|
| 규칙 | `CLAUDE.md` | 항상 지켜야 할 것. 매 대화에 자동으로 읽힌다 | 환경이나 원칙이 바뀔 때 |
| 현재 | `claude-context/architecture.md` | 지금의 아키텍처 | 구성이 바뀔 때 덮어쓴다 |
| 과거 | `docs/architecture-decisions.md` | 결정과 그 이유 | 새 결정을 추가만 한다. 지우지 않는다 |

작업을 마치면 `/update-docs`로 이 문서들과 `JOURNEY.md`를 함께 갱신한다. 구성을 바꿔 놓고
문서를 두면 다음 사람과 Claude가 모두 틀린 정보를 읽게 된다.

되돌릴 수 없는 작업은 `command-guardrails/`에 절차서가 있다. 실행 전에 해당 파일을 먼저 읽는다.

- `kafka-topic-delete.md`: Kafka 토픽 삭제
- `tenant-namespace-delete.md`: 테넌트 네임스페이스 삭제
- `cronjob-manual-run.md`: CronJob 수동 실행

## 8. 자주 묻는 것

### Canary를 중단하고 되돌리려면

```bash
kubectl argo rollouts abort notiflex-api -n notiflex --context gke-sysnet4admin_book_gitaiops
```

중단은 임시 조치다. Git에는 새 태그가 남아 있으므로 ArgoCD가 다시 배포하려 한다. 확실히
되돌리려면 커밋을 되돌린다.

```bash
git revert <문제 커밋> --no-edit && git push origin main
```

반대로 관찰을 건너뛰고 바로 100%로 올리려면 `kubectl argo rollouts promote --full notiflex-api -n notiflex`.

### 특정 요청의 로그를 찾으려면

Grafana Explore에서 Loki 데이터소스를 고르고 LogQL로 조회한다.

```
{namespace="notiflex"} |= "ID 발급 실패"
{namespace="enterprise"} |= "Kafka"
```

### 어느 구간이 느린지 보려면

Grafana Explore에서 Tempo를 고르고 Search로 서비스 `notiflex-api`를 조회한다.
`GET /id` 트레이스를 열면 하위 span으로 `valkey.incr`와 `kafka.publish`가 나뉘어 있어
어디서 시간이 쓰였는지 바로 보인다.

### Kafka 토픽을 추가하려면

`k8s/kafka/`에 KafkaTopic 매니페스트를 추가하고 커밋한다.

```yaml
apiVersion: kafka.strimzi.io/v1
kind: KafkaTopic
metadata:
  name: <토픽 이름>
  namespace: kafka
  labels:
    strimzi.io/cluster: notiflex-kafka
spec:
  partitions: 3
  replicas: 1
```

토픽을 지우는 것은 되돌릴 수 없다. `command-guardrails/kafka-topic-delete.md`를 먼저 읽는다.

### 테넌트를 추가하려면

1. `k8s/<테넌트>/`에 매니페스트를 만든다. `k8s/enterprise/`를 복사해 네임스페이스와
   `NOTIFLEX_TIER` 값을 바꾸면 된다
2. Workload Identity 바인딩을 추가한다
   ```bash
   gcloud iam service-accounts add-iam-policy-binding \
     notiflex-secrets@project-9d5c279f-44bf-42c9-af2.iam.gserviceaccount.com \
     --role=roles/iam.workloadIdentityUser \
     --member="serviceAccount:project-9d5c279f-44bf-42c9-af2.svc.id.goog[<테넌트>/notiflex-api]"
   ```
3. `argocd/apps/notiflex-<테넌트>.yaml`을 만든다. `sync-wave: "2"`, `CreateNamespace=true`
4. 커밋하면 root-app이 알아서 Application을 만든다

### 지금 어떤 알림이 떠 있는지 보려면

```bash
kubectl --context gke-sysnet4admin_book_gitaiops port-forward svc/kube-prometheus-kube-prome-alertmanager -n monitoring 9093:9093
curl -s http://localhost:9093/api/v2/alerts | jq -r '.[] | "\(.labels.alertname) \(.status.state)"'
```

알림 규칙은 `k8s/monitoring/pod-restart-alert.yaml`에 있다. 규칙을 바꾸려면 이 파일을
고치고 커밋한다. Grafana UI에서 만든 알림은 Git에 남지 않으므로 쓰지 않는다(ADR-005).

### 시크릿을 바꾸려면

값은 Google Secret Manager에 있고 매니페스트에는 없다.

```bash
printf '%s' "<새 값>" | gcloud secrets versions add valkey-password --data-file=- \
  --project=project-9d5c279f-44bf-42c9-af2
kubectl --context gke-sysnet4admin_book_gitaiops delete pod -l app=notiflex-api -n notiflex
```

`printf`를 쓴다. `echo`는 개행을 붙여서 인증이 실패한다.

## 9. 막혔을 때

먼저 `JOURNEY.md`의 "트러블슈팅 이력" 표를 본다. 이 클러스터를 만들면서 실제로 겪은 문제와
해결이 챕터별로 정리돼 있다. 아래는 그중 자주 다시 나오는 것들이다.

| 증상 | 확인할 것 |
|------|----------|
| push했는데 배포가 안 된다 | ArgoCD 재조정 주기가 3분이다. 급하면 UI에서 Refresh |
| ArgoCD가 계속 `OutOfSync` | API 서버가 필드를 잘라내는 경우가 있다. Strimzi는 `nodeSelector`가 아니라 `affinity`를 쓴다 |
| Gateway가 `no healthy upstream` | 노드가 바뀌면 NEG 재등록에 1~2분 걸린다. 기다린다 |
| Pod이 이전 코드로 동작한다 | 같은 태그를 다시 빌드했을 때 생긴다. 태그는 덮어쓰지 않는다 |
| Valkey 인증 실패(`WRONGPASS`) | 시크릿 값 끝에 개행이 붙었다. `echo` 대신 `printf` |

그래도 안 되면 `docs/architecture-decisions.md`에서 해당 영역의 ADR을 읽는다. 지금 구조가
왜 이런지 알면 원인 범위가 크게 줄어든다.

## 10. 첫 주에 해볼 만한 것

1. `docs/architecture-decisions.md`를 처음부터 읽는다. ADR 16건이고 한 시간이면 된다
2. 로컬에서 `app/`을 띄우고 `/id`를 호출해 Valkey 카운터가 오르는 것을 확인한다
3. `app/main.go`에 엔드포인트를 하나 추가하고 push해서 Canary가 도는 것을 지켜본다
4. Grafana에서 방금 만든 요청의 트레이스를 찾아본다
5. `/update-docs`를 한 번 돌려 문서가 어떻게 갱신되는지 본다
