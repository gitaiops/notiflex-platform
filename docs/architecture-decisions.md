# Architecture Decision Records

## ADR-001: GitOps 도구로 ArgoCD 채택 (3장)
**시점**: 2026-06 / **결정**: 클러스터 상태를 Git으로 선언·동기화하는 GitOps 도구로 ArgoCD를 채택한다. CLI 푸시 방식(`kubectl apply` 수동)은 쓰지 않는다.
**이유**:
- Git이 단일 진실 공급원(SSOT)이 되어 변경 이력·롤백이 `git revert`로 일관된다
- selfHeal·prune으로 드리프트를 자동 교정한다
- App of Apps 패턴으로 멀티테넌시(ch7) 확장이 용이하다
- UI로 동기화 상태·헬스를 한눈에 본다

## ADR-002: CI로 GitHub Actions + WIF 채택 (3장)
**시점**: 2026-06 / **결정**: 이미지 빌드·푸시 CI로 GitHub Actions를 쓰고, GCP 인증은 Workload Identity Federation(키리스)로 한다. 서비스 계정 키 파일은 쓰지 않는다.
**이유**:
- 저장소와 동일 플랫폼에 통합되어 별도 CI 인프라가 불필요하다
- WIF로 장기 SA 키 노출 위험을 제거한다(조직 정책상 키 생성 차단과도 부합)
- `app/**` 경로 트리거로 코드 변경 시에만 빌드한다

## ADR-003: 메트릭으로 Prometheus + Grafana 채택 (4장)
**시점**: 2026-06 / **결정**: 메트릭 수집·시각화로 kube-prometheus-stack(Prometheus + Grafana + Alertmanager)을 채택한다.
**이유**:
- 오픈소스로 벤더 종속이 없다
- Helm 차트 하나로 Operator·kube-state-metrics·ServiceMonitor가 일괄 구성된다
- PrometheusRule로 알림을 코드로 관리한다
- Grafana에서 메트릭·로그·트레이스를 한 화면에서 본다

## ADR-004: 로깅으로 Loki + Fluent Bit 채택 (4장)
**시점**: 2026-06 / **결정**: 로그 수집·조회로 Loki(SingleBinary) + Fluent Bit를 채택한다. ELK는 쓰지 않는다.
**이유**:
- 인덱스 대신 레이블 기반이라 리소스가 가볍다(e2-medium 노드에 적합)
- Grafana 데이터소스로 통합되어 메트릭과 같은 UI에서 조회한다
- Fluent Bit는 DaemonSet으로 노드 로그를 낮은 오버헤드로 전송한다

## ADR-005: 외부 노출로 Gateway API 채택 (5장)
**시점**: 2026-06 / **결정**: 외부 접근 경로로 GKE Gateway API(`gke-l7-regional-external-managed`)를 채택한다. 기존 Ingress는 쓰지 않는다.
**이유**:
- 역할 분리(Gateway = 인프라, HTTPRoute = 라우팅)로 멀티테넌시에 맞다
- Kubernetes 표준 API로 이식성이 높다
- HealthCheckPolicy로 `/health` 기반 정밀 헬스체크를 한다
- GKE 네이티브 L7 로드밸런서와 직접 통합된다

## ADR-006: 무중단 배포로 Argo Rollouts Blue/Green 채택 (5장)
**시점**: 2026-06 / **결정**: 무중단 배포 전략으로 Argo Rollouts의 Blue/Green을 채택한다. 기본 Deployment의 RollingUpdate는 쓰지 않는다.
**이유**:
- preview 환경에서 신버전을 검증한 뒤 트래픽을 한 번에 전환해 배포 중 끊김을 없앤다
- autoPromotionSeconds로 검증 시간을 둔 자동 승격이 가능하다
- ArgoCD와 같은 argoproj 생태계로 GitOps와 자연스럽게 맞물린다
- ch6에서 Canary 전략으로 전환할 여지를 남긴다

## ADR-007: 캐시로 Valkey 채택 (6장)
**시점**: 2026-06 / **결정**: 분산 카운터·캐시로 Valkey(standalone)를 채택한다. 인메모리 카운터는 폐기한다.
**이유**:
- Redis 프로토콜 호환으로 클라이언트·운영 지식이 그대로 통한다
- 오픈 거버넌스(라이선스 리스크 회피)
- standalone + 최소 리소스로 e2-medium 환경에 적합
- INCR로 클러스터 전역 순차 ID를 원자적으로 보장한다

## ADR-008: 시크릿 관리로 GKE Secret Manager CSI + Workload Identity 채택 (6장)
**시점**: 2026-06 / **결정**: 비밀번호를 Google Secret Manager에 저장하고 GKE managed CSI로 파일 마운트한다. K8s Secret 평문 보관과 SA 키 파일은 쓰지 않는다.
**이유**:
- Workload Identity로 SA 키 없이(키리스) GCP 시크릿에 접근한다
- 파일 마운트가 K8s Secret 자동 생성보다 회전·감사에 안정적이다
- GCP 네이티브 통합으로 별도 Vault 운영 부담이 없다
- 조직 정책(SA 키 생성 차단)과도 부합한다

## ADR-009: 배포 전략을 Canary로 전환 (6장)
**시점**: 2026-06 / **결정**: Blue/Green에서 Canary(20→50→80→100%)로 전환한다.
**이유**:
- 점진적 트래픽 전환으로 새 버전 결함의 영향 범위를 단계적으로 제한한다
- preview 서비스를 canaryService로 재사용해 추가 인프라가 없다
- 단계별 pause로 메트릭 관찰 창을 둔다

## ADR-010: 역할별 멀티 노드풀 채택 (7장)
**시점**: 2026-06 / **결정**: 단일 노드풀 대신 api-pool·worker-pool·ops-pool로 워크로드를 분리한다. nodeSelector 키는 `cloud.google.com/gke-nodepool`로 고정한다.
**이유**:
- 워크로드 격리로 한 컴포넌트의 자원 폭주가 다른 워크로드를 침범하지 않는다
- 워크로드별 머신 타입 최적화(Kafka=standard-2, ops=small)
- GKE 자동 라벨 사용으로 nodeSelector 키 혼선을 없앤다
- 노드풀 단위 확장·교체가 독립적이다
