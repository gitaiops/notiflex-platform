# Notiflex 여정 기록

이 파일은 독자가 실제로 진행한 내용을 기록한다. AI가 각 챕터 완료 시 자동으로 업데이트한다.

## 진행 현황

| 챕터 | 서브챕터 | 상태 | 완료일 | 비고 |
|------|---------|------|--------|------|
| ch2 | 2.2 설치 확인 | ✅ | 2026-06-13 | Claude Code 설치 확인 |
| ch2 | 2.3 gcloud 설정 | ✅ | 2026-06-13 | project, asia-northeast3, AR 인증 |
| ch2 | 2.4 GitHub 저장소 | ✅ | 2026-06-13 | gitaiops/notiflex-platform |
| ch2 | 2.5 GKE 클러스터 | ✅ | 2026-06-13 | notiflex-cluster (Spot, Gateway API) |
| ch2 | 2.6 빌드/배포 | ✅ | 2026-06-13 | api:v0.1.0, Deployment replicas 2 |
| ch2 | 2.7 첫 커밋 | ✅ | 2026-06-13 | Initial commit |
| ch3 | 3.2 GitOps 도구 | ✅ | 2026-06-13 | ArgoCD v3.4.3, notiflex-smb auto-sync |
| ch3 | 3.3 기능 추가 | ✅ | 2026-06-13 | /version 추가, v0.1.1 롤링 업데이트 |
| ch3 | 3.4 CI | ✅ | 2026-06-13 | GitHub Actions + WIF 키리스 인증 |
| ch3 | 3.5 CI-CD 연결 | ✅ | 2026-06-13 | CI(코드→이미지) + GitOps 커밋(매니페스트→배포) |
| ch4 | 4.2 메트릭 모니터링 | ✅ | 2026-06-13 | kube-prometheus-stack (Prometheus/Grafana/Alertmanager) |
| ch4 | 4.3 로그 수집 | ✅ | 2026-06-13 | Loki SingleBinary + Fluent Bit |
| ch4 | 4.4 알림 | ✅ | 2026-06-13 | PrometheusRule pod-restart-alert |
| ch5 | 5.2 트래픽 관리 | ✅ | 2026-06-13 | Gateway API + HTTPRoute + HealthCheckPolicy (IP 35.216.45.48) |
| ch5 | 5.3 무중단 배포 | ✅ | 2026-06-13 | Argo Rollouts Blue/Green, v0.2.0 auto-promote |
| ch6 | 6.1 캐시 | ⬜ | | |
| ch6 | 6.2 시크릿 관리 | ⬜ | | |
| ch6 | 6.3 Canary 전환 | ⬜ | | |
| ch7 | 7.2 멀티 노드풀 | ⬜ | | |
| ch7 | 7.3 App of Apps | ⬜ | | |
| ch7 | 7.4 멀티테넌시 | ⬜ | | |
| ch8 | 8.1 메시징 | ⬜ | | |
| ch8 | 8.2 트레이싱 | ⬜ | | |
| ch8 | 8.3 CronJob | ⬜ | | |
| ch9 | 9.1 저장소 분석 | ⬜ | | |
| ch9 | 9.2 회고 | ⬜ | | |
| ch9 | 9.3 온보딩 문서 | ⬜ | | |
| ch9 | 9.4 GitAIOps 분석 | ⬜ | | |
| ch9 | 9.5 마무리 | ⬜ | | |

## 도구 선택 기록

독자가 3-프롬프트 패턴(탐색→비교→실행)에서 실제로 선택한 도구와 이유를 기록한다.

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-----------|----------|
| GitOps (ch3.2) | ArgoCD | Flux | UI·App of Apps·selfHeal, 멀티테넌시 확장 용이 |
| CI (ch3.4) | GitHub Actions | Jenkins, GitLab CI | 저장소 통합, WIF 키리스 인증 |
| 메트릭 (ch4.2) | Prometheus+Grafana | Datadog, New Relic | 오픈소스, kube-prometheus-stack 일괄 설치 |
| 로깅 (ch4.3) | Loki+Fluent Bit | ELK | 경량, Grafana 통합 조회 |
| 외부 노출 (ch5.2) | Gateway API | Ingress(NGINX), Istio | 역할 분리, K8s 표준, GKE 네이티브 L7 |
| 무중단 배포 (ch5.3) | Argo Rollouts Blue/Green | Flagger, Deployment RollingUpdate | preview 검증 후 일괄 전환, argoproj 통합 |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25 | ch2 |
| Notiflex 이미지 | api:v0.2.0 | v0.1.0 → v0.1.1 → ch5 v0.2.0 |
| Argo Rollouts | v1.8.3 | ch5.3 |
| ArgoCD | v3.4.3 | ch3.2 |
| Kafka | | |
| OTel SDK | | |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot) | 2 | notiflex-api |

## 트러블슈팅 이력

독자가 겪은 문제와 해결 방법을 기록한다. 같은 문제를 다시 겪지 않도록 한다.

| 챕터 | 문제 | 해결 |
|------|------|------|
| | | |
