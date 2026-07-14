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

## ADR-011: GitOps 구조로 App of Apps 채택 (7장)
**시점**: 2026-06 / **결정**: root-app이 `argocd/apps`를 감시(recurse)하여 하위 Application을 일괄 관리한다. Application을 개별 kubectl apply 하지 않는다.
**이유**:
- 새 테넌트·서비스 추가가 매니페스트 한 파일 추가로 끝난다(선언적)
- sync-wave로 인프라→플랫폼→앱 설치 순서를 제어한다
- 루트 하나만 부트스트랩하면 나머지는 자동 동기화된다

## ADR-012: 멀티테넌시로 Namespace 분리 + per-tenant Rollout 채택 (7장)
**시점**: 2026-06 / **결정**: 테넌트(smb/enterprise)를 Namespace로 분리하고 테넌트별 독립 Rollout을 둔다. 단일 namespace 라벨 격리는 쓰지 않는다.
**이유**:
- Namespace 경계로 RBAC·리소스쿼터·정책을 강하게 격리한다
- App of Apps와 결합해 테넌트별 독립 배포·롤백이 가능하다
- 공용 자원(Valkey)은 cross-namespace DNS로 공유해 중복을 피한다
- 테넌트 추가가 디렉터리 + Application 추가로 단순하다

## ADR-013: 이벤트 드리븐 메시징으로 Strimzi Kafka 채택 (8장)
**시점**: 2026-06 / **결정**: 비동기 알림 처리를 위해 Strimzi로 Kafka(KRaft)를 도입한다. 동기 호출 일변도는 지양한다.
**이유**:
- 이벤트 드리븐으로 서비스 간 결합을 낮추고 버스트를 흡수한다
- KRaft 모드로 ZooKeeper 운영 부담이 없다
- Strimzi 오퍼레이터로 K8s 네이티브하게 선언적 운영한다
- worker-pool 격리로 브로커 자원을 분리한다

## ADR-014: 분산 트레이싱으로 Tempo + OpenTelemetry 채택 (8장)
**시점**: 2026-06 / **결정**: 분산 트레이싱 백엔드로 Grafana Tempo, 계측으로 OpenTelemetry SDK(OTLP)를 채택한다.
**이유**:
- OTLP 표준으로 특정 벤더·백엔드에 종속되지 않는다
- Grafana에서 메트릭·로그·트레이스를 한 화면에서 상관 분석한다
- monolithic 모드로 ops-pool에서 경량 운영한다
- 기존 Prometheus/Loki 관측 스택과 자연스럽게 결합한다

## ADR-015: 알림 규칙으로 PrometheusRule 채택 (4장, 소급 기록)
**시점**: 2026-06 (결정) / 2026-07 (기록 소급 — 6월 구축 시 ADR-003에 한 줄로 흡수되어
독립 기록이 누락된 것을 발견하고 뒤늦게 기록한다. 결정 근거: JOURNEY.md 4.4)
**결정**: 알림 규칙을 PrometheusRule CRD로 관리한다 (vs Grafana Alerting).
**이유**:
- Prometheus 네이티브 — PromQL 표현식으로 조건을 정밀하게 정의한다
- git으로 버전 관리 — 알림 규칙의 변경 이력을 추적한다
- kube-prometheus-stack과 자동 연동 — `labels.release` 매칭으로 즉시 활성화된다
- Alertmanager 라우팅과 결합 — 심각도별 수신 채널을 분리한다
- Grafana Alerting은 UI 중심 관리라 GitOps 흐름(선언·리뷰·이력)과 맞지 않는다

## ADR-016: 배치 자동화로 Kubernetes CronJob 채택 (8장, 소급 기록)
**시점**: 2026-06-20 (결정) / 2026-07 (기록 소급 — 8.3 작업이 ADR 일괄 기록(6/13) 이후에
진행되어 누락된 것을 발견하고 뒤늦게 기록한다. 결정 근거: JOURNEY.md 도구 선택 기록)
**결정**: 주기적 헬스체크 자동화를 Kubernetes CronJob으로 구현한다
(vs 외부 cron + 클러스터 외부 트리거, Argo Workflows).
**이유**:
- 쿠버네티스 네이티브 — 별도 스케줄러 없이 클러스터 내장 CronJob을 활용한다
- ops-pool 배치 — 배치 워크로드를 운영 전용 노드에 격리한다
- ArgoCD가 매니페스트로 관리 — git에서 스케줄을 바꾸면 ArgoCD가 자동 반영한다
- Argo Workflows는 5분 주기 단일 잡에는 과한 도구다
