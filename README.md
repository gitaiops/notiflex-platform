# Notiflex Platform

B2B 알림 SaaS 플랫폼을 GKE 위에 올린 저장소다. 「AI 시대에 개발자가 알아야 하는 인프라 구성 배포
with 클로드 코드」의 실습 결과물이고, 코드와 매니페스트뿐 아니라 **결정의 이유와 진행 이력까지
저장소 안에 남긴다**.

처음 보는 사람이라면 다음 순서로 읽으면 된다.

1. [`docs/architecture-decisions.md`](docs/architecture-decisions.md) — 왜 이 구조가 됐는가 (ADR 16건)
2. [`claude-context/architecture.md`](claude-context/architecture.md) — 지금 어떻게 동작하는가
3. [`docs/onboarding.md`](docs/onboarding.md) — 어떻게 손대는가

## 무엇이 올라가 있나

```
[인터넷] → Gateway API → Service(stable/canary) → Rollout(Canary) → Pod
                                                        ├── Valkey     ID 발급 (INCR)
                                                        ├── Kafka      알림 이벤트 비동기 처리
                                                        └── Secret Manager (CSI 마운트)
```

| 영역 | 선택 | 기록 |
|------|------|------|
| GitOps | ArgoCD (App of Apps) | ADR-001, ADR-012 |
| CI | GitHub Actions + Workload Identity Federation | ADR-002 |
| 관측 | Prometheus, Grafana, Loki, Fluent Bit, Tempo | ADR-003, ADR-004, ADR-015 |
| 알림 | PrometheusRule + Alertmanager | ADR-005 |
| 외부 진입점 | GKE Gateway API | ADR-006 |
| 배포 전략 | Argo Rollouts (Blue/Green → Canary) | ADR-007, ADR-010 |
| 캐시 | Valkey | ADR-008 |
| 시크릿 | Google Secret Manager + CSI + Workload Identity | ADR-009 |
| 노드 배치 | 역할별 노드풀 4개 | ADR-011 |
| 멀티테넌시 | Namespace 분리 (SMB / Enterprise) | ADR-013 |
| 메시징 | Strimzi Kafka (KRaft) | ADR-014 |
| 배치 | Kubernetes CronJob | ADR-016 |

애플리케이션은 Go 표준 라이브러리로만 짰고 컨테이너는 scratch 베이스라 이미지가 2.4MB다.

## 3층 지식 구조

이 저장소의 핵심은 인프라 구성 자체가 아니라, AI와 사람이 함께 쓰는 지식을 세 층으로 나눠
누적한다는 점이다.

| 층 | 파일 | 무엇을 담는가 |
|----|------|-------------|
| 규칙 | [`CLAUDE.md`](CLAUDE.md) | 항상 지켜야 할 것. 매 대화에 자동으로 읽힌다 |
| 현재 | [`claude-context/architecture.md`](claude-context/architecture.md) | 지금의 아키텍처 |
| 과거 | [`docs/architecture-decisions.md`](docs/architecture-decisions.md) | 결정과 그 이유. 지우지 않는다 |

이 구조가 있으면 "왜 Grafana Alerting 대신 Alertmanager를 썼나" 같은 질문에 AI가 저장소만
읽고 답할 수 있다. 사람이 기억하고 있을 필요가 없다.

곁들여서 [`JOURNEY.md`](JOURNEY.md)에 진행 이력과 겪은 문제를,
[`command-guardrails/`](command-guardrails/)에 되돌릴 수 없는 작업의 절차를 남긴다.

## 배포

```
app/ 수정 → git push → GitHub Actions 빌드 → 매니페스트 태그 갱신 → ArgoCD → Canary 20/50/80/100%
```

클러스터를 직접 고치지 않는다. ArgoCD selfHeal이 Git 상태로 되돌린다.

## 환경

| 항목 | 값 |
|------|-----|
| 클러스터 | `notiflex-cluster` (GKE Standard, Zonal, Spot VM) |
| 리전 / 존 | asia-northeast3 / asia-northeast3-a |
| 노드풀 | default / api / worker / ops |
| 언어 | Go 1.25 |

## 실습 가이드

이 저장소를 직접 만들어보려면 [gitaiops/_Book_GitAIOps](https://github.com/gitaiops/_Book_GitAIOps)의
가드레일을 따라간다. 자연어로 요청하면 Claude Code가 각 장의 절차를 참조해 실행한다.
