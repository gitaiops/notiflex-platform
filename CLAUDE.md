# Notiflex Platform — 운영 컨텍스트

이 저장소는 「AI 시대에 개발자가 알아야 하는 인프라 구성 배포 with 클로드 코드」 실습으로
Claude Code와 함께 구축하는 Notiflex 플랫폼이다. 진행 이력은 `JOURNEY.md`에 쌓인다.

> **언어 규칙**: 한국어로 진행한다.

## 프로젝트 개요

Notiflex는 B2B 알림 SaaS 플랫폼이다. 고객사가 API로 알림 발송을 요청하면
Notiflex가 대신 처리한다. 고객 등급을 SMB와 Enterprise로 나누어 각각 별도
네임스페이스에서 운영한다.

## 기술 스택

| 영역 | 선택 |
|------|------|
| 애플리케이션 | Go 표준 라이브러리 (net/http), 외부 웹 프레임워크 없음 |
| 컨테이너 | 멀티스테이지 빌드, scratch 베이스 이미지 |
| 인프라 | GKE Standard (Zonal), Spot VM |

## 환경

| 항목 | 값 |
|------|-----|
| GCP 프로젝트 | `project-9d5c279f-44bf-42c9-af2` |
| 리전 / 존 | `asia-northeast3` / `asia-northeast3-a` |
| GKE 클러스터 | `notiflex-cluster` (Standard, Spot, Gateway API standard) |
| kubectl 컨텍스트 | `gke-sysnet4admin_book_gitaiops` |
| Artifact Registry | `asia-northeast3-docker.pkg.dev/project-9d5c279f-44bf-42c9-af2/notiflex` |

## 저장소 구조

```
app/          # Notiflex API (Go, net/http)
k8s/smb/      # SMB 티어 매니페스트
k8s/monitoring/ # 대시보드, 데이터소스, 알림 규칙
argocd/       # ArgoCD Application 정의
helm-values/  # 서드파티 차트 values (버전·값 고정)
.github/      # CI 파이프라인
JOURNEY.md    # 챕터별 진행 이력 + 도구 선택 기록
```

## 배포 원칙

- 매니페스트 변경은 Git 커밋과 push를 거쳐 ArgoCD가 반영한다. `kubectl apply`로 직접 바꾸지 않는다.
- 이미지 태그는 CI가 `sha-<커밋 앞 7자리>`로 갱신한다. 손으로 올릴 때는 `vMAJOR.MINOR.PATCH`를 쓴다.
- 이미 올린 태그는 덮어쓰지 않는다.

## 행동 규칙

- kubectl 명령에는 항상 `--context gke-sysnet4admin_book_gitaiops`를 붙인다.
- 리소스를 변경하기 전에 현재 상태를 먼저 확인한다.
- 삭제나 되돌릴 수 없는 작업은 실행 전에 대상과 범위를 알린다.
