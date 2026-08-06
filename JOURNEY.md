# Notiflex 여정 기록

이 파일은 실제로 진행한 내용을 기록한다. 각 챕터를 마칠 때마다 갱신한다.

## 진행 현황

| 챕터 | 서브챕터 | 상태 | 완료일 | 비고 |
|------|---------|------|--------|------|
| ch2 | 2.2 설치 확인 | ✅ | 2026-08-06 | Claude Code 2.1.223, statusline 구성됨 |
| ch2 | 2.3 gcloud 설정 | ✅ | 2026-08-06 | 프로젝트 project-9d5c279f-44bf-42c9-af2, asia-northeast3-a |
| ch2 | 2.4 GitHub 저장소 | ✅ | 2026-08-06 | gitaiops/notiflex-platform, CLAUDE.md 작성 |
| ch2 | 2.5 GKE 클러스터 | ✅ | 2026-08-06 | notiflex-cluster, GKE 1.35.6, e2-medium Spot x2, Gateway API standard |
| ch2 | 2.6 빌드/배포 | ✅ | 2026-08-06 | api:v0.1.0 배포, Pod 2개 Running, /health·/id 확인 |
| ch2 | 2.7 첫 커밋 | ✅ | 2026-08-06 | main을 새 히스토리로 시작 (이전 이력은 archive/2026-07) |
| ch3 | 3.2 GitOps 도구 | ✅ | 2026-08-06 | ArgoCD v3.5.0, notiflex-smb Application Synced/Healthy |
| ch3 | 3.3 기능 추가 | ✅ | 2026-08-06 | /version 추가 v0.1.1, 롤링 업데이트·롤백·복원 확인 |
| ch3 | 3.4 CI | ✅ | 2026-08-06 | GitHub Actions + WIF 키리스 인증, 빌드 59초 |
| ch3 | 3.5 CI-CD 연결 | ✅ | 2026-08-06 | push → 빌드 → 태그 갱신 → ArgoCD 배포까지 확인 |
| ch4 | 4.2 메트릭 모니터링 | ✅ | 2026-08-06 | kube-prometheus-stack, /metrics + ServiceMonitor, Notiflex 대시보드 |
| ch4 | 4.3 로그 수집 | ✅ | 2026-08-06 | Loki(SingleBinary) + Fluent Bit, Grafana에서 notiflex 로그 조회 확인 |
| ch4 | 4.4 알림 | ✅ | 2026-08-06 | PrometheusRule 2건, Alertmanager 수신 확인 |
| ch5 | 5.2 트래픽 관리 | ✅ | 2026-08-06 | Gateway API, 외부 IP 35.216.104.31, HealthCheckPolicy |
| ch5 | 5.3 무중단 배포 | ✅ | 2026-08-06 | Argo Rollouts Blue/Green, preview 30초 후 자동 승격 확인 |
| ch5 | 5.4 ADR 기록 | ✅ | 2026-08-06 | docs/architecture-decisions.md 신설, ADR-001~007 |
| ch6 | 6.1 캐시 | ✅ | 2026-08-06 | Valkey standalone, INCR로 클러스터 전역 ID |
| ch6 | 6.2 시크릿 관리 | ✅ | 2026-08-06 | Secret Manager CSI + Workload Identity, 키 파일 없음 |
| ch6 | 6.3 Canary 전환 | ✅ | 2026-08-06 | 20/50/80/100% 단계 진행 확인 |
| ch6 | 6.4 아키텍처 스냅샷 | ✅ | 2026-08-06 | claude-context/architecture.md 신설 |
| ch7 | 7.2 멀티 노드풀 | ✅ | 2026-08-06 | api/worker/ops 풀 추가, Rollout에 nodeSelector, replicas 2 복원 |
| ch7 | 7.3 App of Apps | ✅ | 2026-08-06 | root-app이 argocd/apps/ 관리, sync-wave 지정 |
| ch7 | 7.4 멀티테넌시 | ✅ | 2026-08-06 | enterprise 네임스페이스 분리, Valkey 카운터 공유 확인 |
| ch8 | 8.1 메시징 | ✅ | 2026-08-06 | Strimzi 1.1.0 + Kafka 4.3.0(KRaft), Producer/Consumer 확인 |
| ch8 | 8.2 트레이싱 | ✅ | 2026-08-06 | Tempo + OTel, GET /id 트레이스에 하위 span 2개 확인 |
| ch8 | 8.3 CronJob | ✅ | 2026-08-06 | 5분 주기 헬스체크, ops-pool 배치 |
| ch9 | 9.1 저장소 분석 | ✅ | 2026-08-06 | 커밋 36, 파일 44, Go 397줄, YAML 958줄, 문서 588줄 |
| ch9 | 9.2 회고 | ✅ | 2026-08-06 | docs/retrospective.md |
| ch9 | 9.3 온보딩 문서 | ✅ | 2026-08-06 | docs/onboarding.md, README.md |
| ch9 | 9.4 GitAIOps 분석 | ✅ | 2026-08-06 | 회고 문서의 "Git, AI, Ops가 만나는 지점" |
| ch9 | 9.5 마무리 | ✅ | 2026-08-06 | 회고 문서의 "다음에 할 만한 것" |

## 도구 선택 기록

탐색 → 비교 → 실행 과정에서 실제로 선택한 도구와 이유를 기록한다.

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-----------|----------|
| GitOps 도구 (ch3.2) | ArgoCD | Flux, Jenkins X, Spinnaker | Git이 단일 진실 공급원이 되고 selfHeal·prune으로 드리프트를 자동 교정한다. Web UI로 동기화 상태를 눈으로 확인할 수 있어 학습에 유리하다. ch7의 App of Apps로 확장하기 좋다 |
| CI 도구 (ch3.4) | GitHub Actions + Workload Identity Federation | Jenkins, GitLab CI, Cloud Build 직접 호출 | 저장소와 같은 플랫폼이라 별도 CI 인프라가 필요 없다. WIF로 장기 서비스 계정 키를 만들지 않는다(공개 저장소라 특히 중요). `app/**` 경로 트리거로 코드가 바뀔 때만 빌드한다 |
| 메트릭 (ch4.2) | Prometheus + Grafana (kube-prometheus-stack) | Datadog, New Relic, Cloud Monitoring | 오픈소스라 벤더 종속이 없다. Helm 차트 하나로 Operator, kube-state-metrics, Alertmanager가 함께 구성된다. ServiceMonitor로 수집 대상을 매니페스트로 관리한다. 메트릭·로그·트레이스를 Grafana 한 화면에서 본다 |
| 로그 (ch4.3) | Loki + Fluent Bit | ELK(Elasticsearch), Cloud Logging | 전문 색인 대신 라벨 색인이라 e2-medium 노드에서 감당할 수 있다. Grafana 데이터소스로 붙어 메트릭과 같은 UI에서 조회한다. Fluent Bit는 DaemonSet으로 노드 로그를 낮은 부담으로 보낸다 |
| 알림 (ch4.4) | PrometheusRule + Alertmanager | Grafana Alerting, PagerDuty, Cloud Monitoring Alert | 규칙이 YAML로 Git에 남아 리뷰와 `git blame`이 된다. Grafana UI 알림은 Grafana 데이터베이스에만 남아 클러스터를 다시 만들면 사라진다. Alertmanager는 4.2에서 이미 설치돼 추가 비용이 없다 |
| 외부 진입점 (ch5.2) | GKE Gateway API | Ingress, NGINX Ingress Controller, Istio Gateway | GKE 관리형 L7 로드밸런서에 직접 붙어 Controller Pod이 필요 없다. Gateway와 HTTPRoute로 역할이 나뉘어 멀티테넌시에 맞는다. backendRefs weight로 Canary와 이어진다 |
| 배포 전략 (ch5.3) | Argo Rollouts Blue/Green | RollingUpdate, Flagger, Istio | preview 서비스로 새 버전을 먼저 확인한 뒤 트래픽을 넘긴다. replicas 2 규모라 Pod을 잠시 두 배 쓰는 부담이 작다. ArgoCD와 같은 생태계라 ch6 Canary 전환이 쉽다 |
| 캐시·분산 카운터 (ch6.1) | Valkey (standalone) | Redis, Memcached, 인메모리 유지 | INCR가 원자적이라 여러 Pod이 같은 카운터를 안전하게 공유한다. Redis 프로토콜 호환이라 지식이 그대로 통한다. 라이선스 위험이 없다 |
| 시크릿 관리 (ch6.2) | Google Secret Manager + GKE CSI + Workload Identity | K8s Secret 직접 사용, HashiCorp Vault, Sealed Secrets | 키 파일 없이 접근한다. K8s Secret은 base64일 뿐 암호화가 아니다. 매니페스트에 값이 없어 공개 저장소에도 안전하다 |
| 배포 전략 전환 (ch6.3) | Argo Rollouts Canary (20/50/80%) | Blue/Green 유지, 수동 전환 | 문제가 있어도 처음엔 20%만 노출된다. 단계마다 30초 관찰 창을 둔다. preview 서비스를 재사용해 추가 인프라가 없다 |
| 노드 배치 (ch7.2) | 역할별 노드풀 + nodeSelector | 단일 풀 유지, Taint/Toleration, Node Affinity | 무거운 워크로드가 API의 CPU를 잠식하지 않는다. 워크로드마다 맞는 머신 타입을 쓴다. 풀 단위로 늘리고 줄일 수 있다 |
| GitOps 구조 (ch7.3) | App of Apps (root-app + argocd/apps/) | Application 개별 관리, ApplicationSet | 새 앱을 파일 하나 커밋으로 추가한다. 배포 대상이 Git에 다 남는다. sync-wave로 설치 순서를 지정한다 |
| 멀티테넌시 (ch7.4) | Namespace 분리 + 테넌트별 Rollout | 단일 namespace + 라벨 격리, vCluster | RBAC와 쿼터를 테넌트 단위로 건다. 테넌트별로 따로 배포하고 롤백한다. App of Apps와 자연스럽게 맞물린다 |
| 메시징 (ch8.1) | Strimzi Kafka (KRaft) | RabbitMQ, NATS, Pulsar, Pub/Sub | 발송이 느려도 API 응답이 늦지 않는다. 메시지가 디스크에 남아 재처리가 된다. Operator가 CRD로 관리해 GitOps에 그대로 들어온다 |
| 트레이싱 (ch8.2) | Tempo + OpenTelemetry | Jaeger, Zipkin, Cloud Trace | 어느 구간이 느린지 span으로 나눠 본다. Grafana 한 화면에서 메트릭·로그와 함께 본다. OTel은 벤더 중립이라 저장소를 바꿔도 코드를 안 고친다 |
| 배치 자동화 (ch8.3) | Kubernetes CronJob | 클러스터 밖 cron, Argo Workflows | 매니페스트로 관리돼 ArgoCD가 함께 동기화한다. ops-pool에 몰아 API 자원을 건드리지 않는다. 한 단계 작업에 Workflows는 과하다 |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25 | ch2에서 1.25로 시작 (OTel SDK 요구 사항 대비) |
| Notiflex 이미지 | sha-2e7c43e (코드 버전 0.7.1) | v0.1.0 → v0.1.1 → (CI SHA 태그) 0.2.0 /metrics → 0.3.0 tier → 0.4.0 Valkey → 0.5.0 → 0.6.0 Kafka → 0.7.0 트레이싱 → 0.7.1 |
| ArgoCD | v3.5.0 | ch3.2 설치 (stable 매니페스트) |
| kube-prometheus-stack | Prometheus v3.13.2, Grafana 13.1.2 | ch4.2 설치 |
| Loki | 3.6.11 (chart 7.2.0) | ch4.3 설치, SingleBinary |
| Argo Rollouts | latest (ch5.3 시점) | ch5.3 설치, Blue/Green → ch6.3 Canary |
| Valkey | 9.1.1 (bitnami chart) | ch6.1 설치, standalone |
| Fluent Bit | v2.1.0 (chart 2.6.0) | ch4.3 설치 |
| Kafka | 4.3.0 (Strimzi 1.1.0) | ch8.1 설치, KRaft 단일 노드 |
| Tempo | grafana/tempo (단일 바이너리) | ch8.2 설치, ops-pool |
| Argo Rollouts | v1.9 계열 | ch5.3 설치 |
| OTel SDK | v1.45.0 | ch8.2 추가 |
| IBM/sarama | v1.60.1 | ch8.1 추가 |
| valkey-go | v1.0.73 | ch6.1 추가 |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot) | 2 | ArgoCD, 관측 스택, Valkey |
| api-pool | e2-medium (Spot) | 1 | notiflex-api (SMB, Enterprise) |
| worker-pool | e2-standard-2 (Spot) | 1 | Kafka (ch8) |
| ops-pool | e2-small (Spot) | 1 | Tempo, CronJob (ch8) |

## 트러블슈팅 이력

| 챕터 | 문제 | 해결 |
|------|------|------|
| ch2 | 새 GCP 프로젝트에서 `gcloud builds submit`이 403(`storage.objects.get denied`)으로 실패 | Compute 기본 서비스 계정(`<PROJECT_NUM>-compute@developer.gserviceaccount.com`)에 `roles/cloudbuild.builds.builder`, `roles/logging.logWriter`, `roles/artifactregistry.writer`, `roles/storage.objectAdmin` 부여 후 재시도 |
| ch2 | 클러스터 생성 직후 `kubectl get gatewayclass`가 비어 있음 | CRD는 먼저 설치되고 GatewayClass 리소스는 몇 분 뒤에 생성된다. 기다리면 4개(`gke-l7-global-external-managed` 등)가 나타난다 |
| ch2 | 같은 태그(v0.1.0)로 이미지를 다시 빌드했더니 Pod이 이전 코드로 동작 | `imagePullPolicy`가 기본값 `IfNotPresent`라 노드 캐시를 사용한다. 일시적으로 `Always`로 패치해 새로 받은 뒤 패치를 되돌렸다. 태그는 덮어쓰지 않는 것이 원칙 |
| ch3 | CI가 매니페스트를 push할 때 403 | 조직(gitaiops) 설정에서 워크플로 쓰기 권한이 꺼져 있었다. Settings > Actions > General > Workflow permissions를 "Read and write"로 바꾸면 해결된다. 저장소 설정만으로는 안 되고 조직 설정이 먼저다 |
| ch3 | ArgoCD 자동 동기화가 push 직후 바로 반영되지 않음 | 기본 재조정 주기가 3분이라 최대 3분까지 기다린다. 급하면 `argocd app sync` 또는 UI의 Refresh를 쓴다 |
| ch4 | Loki가 `mkdir /var/loki: read-only file system`으로 CrashLoopBackOff | `singleBinary.persistence.enabled: false`로 두면 `/var/loki` 볼륨이 생기지 않는데 루트 파일시스템은 읽기 전용이다. `extraVolumes`/`extraVolumeMounts`로 emptyDir을 직접 붙인다 |
| ch4 | TargetDown 알림이 계속 발생 (coredns) | GKE는 kube-dns를 쓰고 9153 포트를 열지 않는다. `coreDns.enabled: false`로 수집 대상에서 제외한다 |
| ch4 | KubeCPUOvercommit 알림 발생 | e2-medium 2노드의 실제 allocatable은 노드당 949m(합계 1880m)으로, 책 예산표의 3200m보다 작다. ch6 진입 전 관측 스택 requests를 줄이고 ch7에서 노드풀을 늘려 해소한다 |
| ch5 | Gateway가 `An active proxy-only subnetwork is required`로 IP를 못 받음 | 리전 외부 Gateway는 proxy-only 서브넷이 필요한데 자동 생성되지 않는다. `gcloud compute networks subnets create proxy-only-subnet --purpose=REGIONAL_MANAGED_PROXY --role=ACTIVE --region=asia-northeast3 --network=default --range=172.16.0.0/23`로 만든다 |
| ch6 | CSI DaemonSet 추가 후 노드 CPU 100%, Valkey까지 Pending | GKE 관리형 로깅/모니터링 에이전트(fluentbit-gke, gke-metrics-agent, gmp)가 노드당 약 130m을 쓰는데 Loki·Prometheus와 중복이다. `gcloud container clusters update --logging=NONE --monitoring=NONE`으로 끄면 약 260m이 확보된다 |
| ch6 | Blue/Green에서 Canary로 바꿔도 전략이 그대로 유지됨 | `kubectl apply`만으로는 전환되지 않는다. Git push를 먼저 하고 `kubectl delete rollout` 후 ArgoCD가 새 정의로 다시 만들게 한다. preview 서비스는 canaryService로 계속 쓰므로 지우면 안 된다 |
| ch6 | Rollout을 지운 뒤 ArgoCD가 OutOfSync로 남음 | `kubectl annotate application notiflex-smb -n argocd argocd.argoproj.io/refresh=hard --overwrite`로 즉시 재조정한다 |
| ch7 | 노드풀 생성 시 `Cluster is running incompatible operation` | ch6.2의 클러스터 업데이트가 아직 실행 중이면 노드풀을 만들 수 없다. `gcloud container operations list --zone=<ZONE> --filter="status=RUNNING"`이 비어 있는지 먼저 확인한다 |
| ch7 | Pod을 api-pool로 옮긴 직후 Gateway가 `no healthy upstream` | 로드밸런서가 새 노드의 NEG를 다시 등록하는 데 1~2분 걸린다. 기다리면 복구된다 |
| ch8 | 가드레일이 적은 Strimzi 0.51 / Kafka 4.1.0 조합이 설치되지 않음 | 차트가 1.1.0으로 올라가 Kafka 4.2~4.3만 지원한다. 지원 목록은 `kubectl get deploy strimzi-cluster-operator -n kafka -o jsonpath='{...STRIMZI_KAFKA_IMAGES...}'`로 확인한다. 4.3.0과 metadataVersion 4.3-IV0을 썼다 |
| ch8 | Strimzi 1.x는 KafkaNodePool이 필수 | v1 API에서 `spec.kafka.replicas`와 `spec.kafka.resources`가 사라졌다. replicas, storage, resources는 KafkaNodePool에 쓰고 Kafka에는 `strimzi.io/node-pools: enabled` 어노테이션을 단다 |
| ch8 | Grafana 데이터소스 기본값 충돌 | Loki와 Tempo 데이터소스를 직접 ConfigMap으로 등록하면서 둘 다 `isDefault: false`로 고정했다. 기본값이 두 개면 Grafana가 기동에 실패한다 |
| ch8 | Strimzi 매니페스트가 계속 OutOfSync | Strimzi의 pod 템플릿은 `nodeSelector`를 받지 않는다. 지정하면 API 서버가 값을 잘라내 ArgoCD가 영구 OutOfSync로 본다. `affinity.nodeAffinity`로 바꾼다. CRD 스키마는 `kubectl get crd <name> -o json | jq '...template.properties.pod.properties'`로 확인한다 |
| ch9 | CI가 SMB 매니페스트만 갱신해 두 테넌트의 이미지가 어긋남 | `ci.yaml`의 sed 대상이 `k8s/smb/rollout.yaml` 하나였다. `TARGETS` 변수로 enterprise까지 포함하도록 고쳤다. 테넌트를 추가하면 이 목록도 함께 늘려야 한다 |
