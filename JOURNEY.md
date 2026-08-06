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
| GitOps 도구 (ch3.2) | ArgoCD | Flux, Jenkins X, Spinnaker | Git이 단일 진실 공급원이 되고 selfHeal·prune으로 드리프트를 자동 교정한다. Web UI로 동기화 상태를 눈으로 확인할 수 있어 학습에 유리하다. ch7의 App of Apps로 확장하기 좋다 |
| CI 도구 (ch3.4) | GitHub Actions + Workload Identity Federation | Jenkins, GitLab CI, Cloud Build 직접 호출 | 저장소와 같은 플랫폼이라 별도 CI 인프라가 필요 없다. WIF로 장기 서비스 계정 키를 만들지 않는다(공개 저장소라 특히 중요). `app/**` 경로 트리거로 코드가 바뀔 때만 빌드한다 |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25 | ch2에서 1.25로 시작 (OTel SDK 요구 사항 대비) |
| Notiflex 이미지 | sha-bc504f7 (코드 버전 0.1.1) | ch2 v0.1.0 → ch3.3 v0.1.1 → ch3.5부터 CI가 SHA 태그로 갱신 |
| ArgoCD | v3.5.0 | ch3.2 설치 (stable 매니페스트) |
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
| ch3 | CI가 매니페스트를 push할 때 403 | 조직(gitaiops) 설정에서 워크플로 쓰기 권한이 꺼져 있었다. Settings > Actions > General > Workflow permissions를 "Read and write"로 바꾸면 해결된다. 저장소 설정만으로는 안 되고 조직 설정이 먼저다 |
| ch3 | ArgoCD 자동 동기화가 push 직후 바로 반영되지 않음 | 기본 재조정 주기가 3분이라 최대 3분까지 기다린다. 급하면 `argocd app sync` 또는 UI의 Refresh를 쓴다 |
