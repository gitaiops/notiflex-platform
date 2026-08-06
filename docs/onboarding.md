# Notiflex 온보딩

새로 합류한 엔지니어가 이 플랫폼을 파악하고 첫 작업을 하기까지 필요한 내용을 모았다.
아키텍처 전체 그림은 `claude-context/architecture.md`, 결정의 이유는 `docs/architecture-decisions.md`를 본다.

## 0. 먼저 알아둘 것

이 저장소는 **Git이 유일한 진실**이다. 클러스터를 `kubectl apply`나 `kubectl edit`으로 직접
고치면 ArgoCD가 몇 분 안에 Git 상태로 되돌린다. 무엇을 바꾸든 매니페스트를 고치고 커밋한다.

## 1. 접근 준비

```bash
gcloud container clusters get-credentials notiflex-cluster --zone=asia-northeast3-a
kubectl config rename-context "$(kubectl config current-context)" gke-sysnet4admin_book_gitaiops
```

모든 kubectl 명령에 `--context gke-sysnet4admin_book_gitaiops`를 붙인다. 다른 클러스터에
명령이 들어가는 사고를 막기 위한 규칙이다.

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

## 4. 접속 방법

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

- **Prometheus** — 메트릭. 대시보드 "Notiflex API"에 요청 수, ID 발급 수, CPU, 메모리, 재시작 횟수가 있다
- **Loki** — 로그. Explore에서 `{namespace="notiflex"}`
- **Tempo** — 트레이스. Explore에서 Search로 서비스 `notiflex-api` 조회

## 5. 배포 흐름

```
app/ 코드 수정 → git push
    → GitHub Actions가 이미지 빌드 (Workload Identity, 키 파일 없음)
    → Artifact Registry에 notiflex/api:sha-<커밋 7자리> 푸시
    → CI가 k8s/smb/rollout.yaml 태그를 갱신하고 다시 push
    → ArgoCD가 감지 (최대 3분)
    → Argo Rollouts Canary 20% → 50% → 80% → 100% (단계마다 30초)
```

매니페스트만 바꾸는 경우 CI는 돌지 않고 ArgoCD가 바로 반영한다.

진행 상황은 이렇게 본다.

```bash
kubectl argo rollouts get rollout notiflex-api -n notiflex --context gke-sysnet4admin_book_gitaiops -w
```

## 6. 자주 묻는 것

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

## 7. 첫 작업으로 해볼 만한 것

1. `app/main.go`에 엔드포인트를 하나 추가하고 push해서 Canary가 도는 것을 지켜본다
2. Grafana에서 방금 만든 요청의 트레이스를 찾아본다
3. `docs/architecture-decisions.md`를 읽고 왜 지금 구조가 됐는지 파악한다
