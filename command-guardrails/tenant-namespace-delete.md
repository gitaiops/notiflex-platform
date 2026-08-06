# 테넌트 Namespace 삭제

고객 하나를 통째로 내리는 작업이다. 실수하면 서비스가 중단되고 데이터가 사라진다.

## 사전 확인

1. 해당 네임스페이스의 모든 리소스를 본다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops get all,pvc,secret,cm -n <TENANT>
   ```
2. 남겨야 할 영구 자원이 있는지 판단한다. 백업이 필요하면 먼저 받는다.
3. 다른 네임스페이스에서 이 테넌트를 참조하는지 확인한다. 지금 구조에서는 두 테넌트가 `notiflex` 네임스페이스의 Valkey를 함께 쓴다. SMB(`notiflex`)를 지우면 Enterprise도 ID를 발급하지 못한다.
4. 어느 ArgoCD Application이 이 네임스페이스를 관리하는지 확인한다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops get app -n argocd \
     -o custom-columns=NAME:.metadata.name,NS:.spec.destination.namespace
   ```

## 실행

`kubectl delete namespace`를 직접 쓰지 않는다. ArgoCD가 관리하는 리소스라 selfHeal이 되살리거나, Application만 남아 계속 오류를 낸다.

1. Application 정의를 지우고 커밋한다. root-app이 이를 감지해 하위 Application을 정리한다.
   ```bash
   git rm argocd/apps/notiflex-<TENANT>.yaml
   git commit -m "chore: <TENANT> 테넌트 제거"
   git push origin main
   ```
2. ArgoCD가 정리를 끝낼 때까지 기다린다. `prune: true`라 Application에 속한 리소스가 함께 사라진다.
3. 매니페스트 디렉터리도 지운다.
   ```bash
   git rm -r k8s/<TENANT>
   git commit -m "chore: <TENANT> 매니페스트 제거"
   git push origin main
   ```
4. Workload Identity 바인딩도 정리한다.
   ```bash
   gcloud iam service-accounts remove-iam-policy-binding \
     notiflex-secrets@project-9d5c279f-44bf-42c9-af2.iam.gserviceaccount.com \
     --role=roles/iam.workloadIdentityUser \
     --member="serviceAccount:project-9d5c279f-44bf-42c9-af2.svc.id.goog[<TENANT>/notiflex-api]"
   ```

## 사후 검증

1. Application이 사라졌는지 확인한다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops get app -n argocd
   ```
2. 네임스페이스에 남은 리소스가 없는지 본다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops get all -n <TENANT>
   ```
3. 남은 테넌트의 API가 정상 응답하는지 확인한다.
   ```bash
   curl http://35.216.104.31/health
   ```
