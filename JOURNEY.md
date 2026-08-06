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
| ch3 | 3.2 GitOps 도구 | ⬜ | | |
| ch3 | 3.3 기능 추가 | ⬜ | | |
| ch3 | 3.4 CI | ⬜ | | |
| ch3 | 3.5 CI-CD 연결 | ⬜ | | |
| ch4 | 4.2 메트릭 모니터링 | ⬜ | | |
| ch4 | 4.3 로그 수집 | ⬜ | | |
| ch4 | 4.4 알림 | ⬜ | | |
| ch5 | 5.2 트래픽 관리 | ⬜ | | |
| ch5 | 5.3 무중단 배포 | ⬜ | | |
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

탐색 → 비교 → 실행 과정에서 실제로 선택한 도구와 이유를 기록한다.

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-----------|----------|
| | | | |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25 | ch2에서 1.25로 시작 (OTel SDK 요구 사항 대비) |
| Notiflex 이미지 | v0.1.0 | ch2 초기 빌드 |
| ArgoCD | | |
| Kafka | | |
| OTel SDK | | |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|----------|---------|-------------|
| default-pool | e2-medium (Spot) | 2 | notiflex-api |

## 트러블슈팅 이력

| 챕터 | 문제 | 해결 |
|------|------|------|
| ch2 | 새 GCP 프로젝트에서 `gcloud builds submit`이 403(`storage.objects.get denied`)으로 실패 | Compute 기본 서비스 계정(`<PROJECT_NUM>-compute@developer.gserviceaccount.com`)에 `roles/cloudbuild.builds.builder`, `roles/logging.logWriter`, `roles/artifactregistry.writer`, `roles/storage.objectAdmin` 부여 후 재시도 |
| ch2 | 클러스터 생성 직후 `kubectl get gatewayclass`가 비어 있음 | CRD는 먼저 설치되고 GatewayClass 리소스는 몇 분 뒤에 생성된다. 기다리면 4개(`gke-l7-global-external-managed` 등)가 나타난다 |
| ch2 | 같은 태그(v0.1.0)로 이미지를 다시 빌드했더니 Pod이 이전 코드로 동작 | `imagePullPolicy`가 기본값 `IfNotPresent`라 노드 캐시를 사용한다. 일시적으로 `Always`로 패치해 새로 받은 뒤 패치를 되돌렸다. 태그는 덮어쓰지 않는 것이 원칙 |
