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
