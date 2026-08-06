# Architecture Decision Records

이 파일은 Notiflex 플랫폼을 만들면서 내린 결정을 시간 순서대로 쌓는다.
한 결정에 한 항목이고, 형식은 모두 같다. "무엇을 골랐는가"보다 "왜 골랐는가"를 남기는 것이 목적이다.

`JOURNEY.md`가 진행 기록이라면 이 파일은 결정 기록이다. 도구를 바꿀지 검토할 때 먼저 여기를 읽는다.

## ADR-001: GitOps 도구로 ArgoCD 채택 (3장)
**시점**: 2026-08 / **결정**: 클러스터 상태를 Git으로 선언하고 동기화하는 GitOps 도구로 ArgoCD를 채택한다. `kubectl apply` 수동 배포는 쓰지 않는다.
**이유**:
- Git이 단일 진실 공급원이 되어 변경 이력과 롤백이 `git revert` 하나로 일관된다
- selfHeal과 prune으로 클러스터를 직접 고친 흔적을 자동으로 되돌린다
- Web UI로 동기화 상태와 헬스를 눈으로 확인할 수 있어 학습 과정에 유리하다
- 7장의 App of Apps 패턴으로 멀티테넌시까지 확장할 수 있다

## ADR-002: CI로 GitHub Actions + Workload Identity Federation 채택 (3장)
**시점**: 2026-08 / **결정**: 이미지 빌드와 푸시는 GitHub Actions로 하고, GCP 인증은 Workload Identity Federation(키리스)으로 한다. 서비스 계정 키 파일은 만들지 않는다.
**이유**:
- 저장소와 같은 플랫폼에 붙어 있어 별도 CI 인프라를 운영하지 않아도 된다
- WIF는 장기 서비스 계정 키를 만들지 않으므로 키 유출 위험이 없다. 이 저장소가 공개 저장소라 특히 중요하다
- `app/**` 경로 트리거로 코드가 바뀔 때만 빌드해 불필요한 실행을 줄인다
- 빌드 후 매니페스트 태그를 갱신해 push하면 ArgoCD가 이어받아 배포까지 자동으로 이어진다

## ADR-003: 메트릭으로 Prometheus + Grafana 채택 (4장)
**시점**: 2026-08 / **결정**: 메트릭 수집과 시각화로 kube-prometheus-stack(Prometheus + Grafana + Alertmanager + Operator)을 채택한다. Datadog 같은 SaaS는 쓰지 않는다.
**이유**:
- 오픈소스라 벤더에 묶이지 않고 비용이 사용량에 따라 늘지 않는다
- Helm 차트 하나로 Operator, kube-state-metrics, node-exporter, Alertmanager가 함께 구성된다
- ServiceMonitor CRD로 수집 대상을 매니페스트로 관리해 GitOps 흐름과 맞물린다
- 메트릭, 로그, 트레이스를 Grafana 한 화면에서 볼 수 있다

## ADR-004: 로그 수집으로 Loki + Fluent Bit 채택 (4장)
**시점**: 2026-08 / **결정**: 로그 수집과 조회로 Loki(SingleBinary) + Fluent Bit를 채택한다. ELK(Elasticsearch)는 쓰지 않는다.
**이유**:
- 로그 본문을 전부 색인하지 않고 라벨만 색인해서 e2-medium 노드에서도 감당할 수 있다
- Grafana 데이터소스로 붙어 메트릭과 같은 UI에서 조회한다. 별도 Kibana를 띄우지 않는다
- Fluent Bit는 DaemonSet으로 노드 로그를 낮은 부담으로 전송한다
- Elasticsearch는 JVM 힙과 마스터 노드 구성이 필요해 이 규모에 과하다

## ADR-005: 알림으로 PrometheusRule + Alertmanager 채택 (4장)
**시점**: 2026-08 / **결정**: 알림 규칙은 `PrometheusRule` CRD로 작성하고 Alertmanager가 라우팅한다. Grafana Alerting은 쓰지 않는다.
**이유**:
- 규칙이 YAML로 Git 저장소에 남는다. Grafana Alerting은 규칙이 Grafana 데이터베이스에만 있어 클러스터를 다시 만들면 사라지고 인계도 어렵다
- 알림 변경이 Pull Request로 올라와 리뷰를 거친다. "이 임계값이 왜 5분이지?"를 `git blame`으로 답할 수 있다
- Alertmanager는 4장에서 kube-prometheus-stack을 설치할 때 이미 함께 들어왔다. 추가로 설치하거나 운영할 것이 없다
- ArgoCD로 GitOps를 깔아온 흐름에서 알림만 UI 설정으로 빠지면 그동안 쌓은 패턴이 끊긴다

## ADR-006: 외부 진입점으로 Gateway API 채택 (5장)
**시점**: 2026-08 / **결정**: 외부 접근 경로로 GKE Gateway API(`gke-l7-regional-external-managed`)를 채택한다. Ingress와 NGINX Ingress Controller는 쓰지 않는다.
**이유**:
- GKE가 관리하는 L7 로드밸런서와 직접 연결되어 별도 Controller Pod을 띄우지 않는다. NGINX Ingress는 100m 이상을 상시로 쓴다
- 역할이 분리된다. Gateway는 인프라 담당이 관리하고 HTTPRoute는 서비스 담당이 관리한다. 7장 멀티테넌시에서 이 구분이 그대로 쓰인다
- HTTPRoute의 `backendRefs` weight로 트래픽을 비율로 나눌 수 있어 Canary(6장)와 맞물린다
- Kubernetes 공식 표준이고 Ingress의 후속이다. Ingress는 어노테이션으로 벤더별 기능을 넣는 방식이라 이식성이 떨어진다

## ADR-007: 무중단 배포로 Argo Rollouts Blue/Green 채택 (5장)
**시점**: 2026-08 / **결정**: 배포 전략으로 Argo Rollouts의 Blue/Green을 채택한다. 기본 Deployment의 RollingUpdate는 쓰지 않는다.
**이유**:
- RollingUpdate는 배포와 검증이 섞인다. 새 Pod이 뜨는 순간 바로 실제 트래픽을 받는다
- preview 서비스로 새 버전을 먼저 붙여 확인한 뒤 트래픽을 한 번에 넘긴다. `autoPromotionSeconds: 30`으로 확인 시간을 둔다
- replicas가 2라 Blue/Green이 잠시 Pod을 두 배로 쓰는 부담이 크지 않다
- ArgoCD와 같은 argoproj 생태계라 GitOps와 자연스럽게 맞물리고, 6장에서 Canary로 넘어갈 때 같은 Rollout 리소스를 그대로 쓴다

## ADR-008: 분산 카운터로 Valkey 채택 (6장)
**시점**: 2026-08 / **결정**: ID 발급 카운터를 Valkey(standalone)로 옮긴다. 프로세스 안 인메모리 카운터는 버린다.
**이유**:
- 인메모리 카운터는 Pod마다 값이 따로 돌아 같은 번호가 여러 번 나온다. Pod을 재시작하면 1부터 다시 시작한다
- Valkey의 `INCR`는 원자적이라 여러 Pod이 동시에 호출해도 번호가 겹치지 않는다
- Redis 프로토콜과 호환되어 클라이언트와 운영 지식이 그대로 통한다
- Redis가 2024년 라이선스를 바꾼 뒤 Linux Foundation으로 옮겨간 포크라 라이선스 위험이 없다

## ADR-009: 시크릿 관리로 Secret Manager CSI + Workload Identity 채택 (6장)
**시점**: 2026-08 / **결정**: 비밀번호를 Google Secret Manager에 두고 GKE managed CSI로 파일 마운트한다. Kubernetes Secret에 평문으로 두거나 서비스 계정 키 파일을 쓰지 않는다.
**이유**:
- Workload Identity로 키 파일 없이 GCP 시크릿에 접근한다. 유출될 장기 자격 증명 자체가 없다
- Kubernetes Secret은 base64 인코딩일 뿐 암호화가 아니다. `kubectl get secret -o yaml`로 누구나 읽는다
- 시크릿을 바꿀 때 Secret Manager에서 새 버전만 올리면 되고, 회전과 감사 기록이 GCP 쪽에 남는다
- 매니페스트에 값이 들어가지 않으므로 Git 저장소가 공개여도 문제가 없다

## ADR-010: 배포 전략을 Blue/Green에서 Canary로 전환 (6장)
**시점**: 2026-08 / **결정**: Argo Rollouts 전략을 Canary(20 → 50 → 80 → 100%)로 바꾼다. Blue/Green은 더 쓰지 않는다.
**이유**:
- Blue/Green은 전환이 한 번에 일어나 문제가 있으면 전체 사용자가 동시에 겪는다. Canary는 처음 20%만 노출된다
- 각 단계에 30초 관찰 시간을 둬서 메트릭과 로그로 이상을 확인할 창을 만든다
- 5장에서 만든 preview 서비스를 `canaryService`로 그대로 재사용하므로 추가 인프라가 없다
- 전략만 바꾸면 되고 Rollout 리소스와 GitOps 흐름은 그대로다

## ADR-011: 역할별 노드풀 분리 채택 (7장)
**시점**: 2026-08 / **결정**: 워크로드 성격에 따라 노드풀을 `api-pool`, `worker-pool`, `ops-pool`로 나누고 `nodeSelector`로 배치한다. 모든 워크로드를 한 풀에 섞지 않는다.
**이유**:
- 한 풀에 다 올리면 무거운 워크로드(Kafka)가 API Pod의 CPU를 잠식한다. 실제로 6장에서 노드 CPU가 100%에 닿아 Valkey가 Pending에 걸렸다
- 워크로드마다 필요한 머신이 다르다. API는 e2-medium, Kafka는 e2-standard-2, 운영 도구는 e2-small로 각각 맞춘다
- 풀 단위로 늘리고 줄일 수 있어 비용을 필요한 곳에만 쓴다
- `nodeSelector` 키는 GKE가 자동으로 붙이는 `cloud.google.com/gke-nodepool`만 쓴다. 커스텀 라벨을 만들면 어떤 키를 써야 하는지 헷갈려 Pod이 영구 Pending에 빠지기 쉽다

## ADR-012: GitOps 구조로 App of Apps 채택 (7장)
**시점**: 2026-08 / **결정**: `argocd/root-app.yaml` 하나가 `argocd/apps/` 아래의 Application들을 관리하게 한다. Application을 사람이 하나씩 `kubectl apply`하지 않는다.
**이유**:
- 새 테넌트나 컴포넌트를 추가할 때 `argocd/apps/`에 파일 하나를 커밋하면 끝난다. 클러스터에 손대지 않는다
- Application 정의 자체가 Git에 남아 무엇이 배포 대상인지 저장소만 봐도 알 수 있다
- `sync-wave`로 인프라(0), 플랫폼(1), 애플리케이션(2) 순서를 정해 의존성이 깨지지 않게 한다
- `argocd/` 직속에는 root-app만 두어 루트가 자기 자신을 다시 동기화하는 순환을 피한다

## ADR-013: 멀티테넌시로 Namespace 분리 + 테넌트별 Rollout 채택 (7장)
**시점**: 2026-08 / **결정**: 고객 등급별로 네임스페이스를 나누고(`notiflex`=SMB, `enterprise`) 각각 별도 Rollout을 둔다. 한 네임스페이스에서 라벨로만 구분하지 않는다.
**이유**:
- 네임스페이스가 나뉘면 RBAC, ResourceQuota, NetworkPolicy를 테넌트 단위로 걸 수 있다. 라벨 구분은 이런 경계를 만들지 못한다
- 테넌트별로 따로 배포하고 롤백한다. 한쪽 배포가 잘못돼도 다른 쪽은 영향받지 않는다
- App of Apps와 자연스럽게 맞물린다. 테넌트를 추가하려면 Application 파일 하나와 매니페스트 디렉터리 하나를 더하면 된다
- vCluster처럼 테넌트마다 컨트롤 플레인을 띄우는 방식은 이 규모에 과하고 노드 자원을 크게 쓴다

## ADR-014: 이벤트 처리로 Strimzi Kafka 채택 (8장)
**시점**: 2026-08 / **결정**: 알림 발송을 Kafka(KRaft 모드, Strimzi 운영)로 비동기 처리한다. API가 발송까지 직접 처리하지 않는다.
**이유**:
- 알림 발송이 느려도 API 응답이 늦어지지 않는다. 요청은 이벤트만 던지고 바로 끝난다
- 메시지가 디스크에 남아 Consumer가 죽어도 재시작하면 이어서 처리한다. RabbitMQ보다 재처리와 되감기가 자연스럽다
- Strimzi Operator가 CRD로 브로커와 토픽을 관리해 매니페스트만 커밋하면 된다. GitOps 흐름에 그대로 들어온다
- KRaft 모드라 ZooKeeper를 따로 운영하지 않는다. 단일 노드 구성에서 부담이 크게 줄어든다

## ADR-015: 분산 트레이싱으로 Tempo + OpenTelemetry 채택 (8장)
**시점**: 2026-08 / **결정**: 요청 추적은 OpenTelemetry SDK로 계측하고 Tempo에 저장한다. Jaeger는 쓰지 않는다.
**이유**:
- 메트릭은 "느리다"까지만 알려준다. 어느 구간이 느린지는 span으로 나눠야 보인다. `valkey.incr`와 `kafka.publish`를 분리해 두었다
- Tempo가 Grafana 데이터소스로 붙어 메트릭, 로그, 트레이스를 한 화면에서 오간다. Jaeger는 별도 UI를 띄워야 한다
- OpenTelemetry는 벤더 중립 표준이라 나중에 저장소를 바꿔도 앱 코드를 고치지 않는다
- 저장소로 객체 스토리지를 쓸 수 있어 색인 기반인 Elasticsearch보다 보관 비용이 낮다

## ADR-016: 배치 자동화로 Kubernetes CronJob 채택 (8장)
**시점**: 2026-08 / **결정**: 주기 실행 작업은 Kubernetes CronJob으로 만든다. 클러스터 밖의 cron이나 Argo Workflows를 쓰지 않는다.
**이유**:
- 매니페스트로 관리되어 ArgoCD가 다른 리소스와 함께 동기화한다. 스케줄 변경도 Git 커밋으로 남는다
- 클러스터 밖 cron은 클러스터에 접근할 자격 증명을 따로 관리해야 하고, 실행 기록이 클러스터 밖에 흩어진다
- `nodeSelector`로 ops-pool에 몰아 API 노드의 자원을 건드리지 않는다
- Argo Workflows는 여러 단계로 이어지는 파이프라인용이다. 헬스체크 한 단계짜리 작업에는 과하다
