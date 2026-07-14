# Notiflex Platform — 운영 컨텍스트

이 저장소는 「AI 시대에 개발자가 알아야 할 인프라 구성 배포 with 클로드 코드」 실습으로
Claude Code와 함께 구축하는 Notiflex 플랫폼이다. 진행 이력은 `JOURNEY.md`에 누적된다.

> **언어 규칙**: 한국어로 진행한다.

## 환경

| 항목 | 값 |
|------|-----|
| GCP 프로젝트 | `project-75fce205-dfa5-4975-a56` |
| 리전 / 존 | `asia-northeast3` / `asia-northeast3-a` |
| GKE 클러스터 | `notiflex-cluster` (Standard, Spot, Gateway API standard) |
| kubectl 컨텍스트 | `gke-notiflex` |
| Artifact Registry | `asia-northeast3-docker.pkg.dev/project-75fce205-dfa5-4975-a56/notiflex` |

> kubectl 명령은 항상 `--context gke-notiflex`를 붙여 대상 클러스터를 명확히 한다.

## 저장소 구조

```
app/                 # Notiflex API (Go, net/http)
k8s/smb/             # SMB 티어 (rollout, gateway, secret-provider, healthcheck-cronjob)
k8s/enterprise/      # Enterprise 티어 (SMB와 동일 패턴, 별도 namespace)
k8s/kafka/           # Strimzi Kafka(KRaft) + notifications 토픽
k8s/monitoring/      # PrometheusRule 알림 정의
helm-values/         # 서드파티 차트 values (버전·값 고정)
argocd/              # App of Apps (root-app + apps/)
claude-context/      # 현재 아키텍처 스냅샷 (매 대화 참조)
docs/                # ADR 16건 + 온보딩 가이드
JOURNEY.md           # 챕터별 진행 이력 + 도구 선택 기록
```

## 배포 원칙

- 매니페스트 변경은 Git 커밋 → push → ArgoCD 동기화 경로를 따른다 (ch3 이후).
- 이미지 태그는 `api:vMAJOR.MINOR.PATCH` 규칙으로 올린다.
