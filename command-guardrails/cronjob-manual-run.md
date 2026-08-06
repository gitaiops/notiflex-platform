# CronJob 수동 실행

정기 실행과 겹치면 같은 작업이 두 번 돌 수 있다.

## 사전 확인

1. 다음 정기 실행까지 시간이 남았는지 본다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops get cronjob notiflex-healthcheck -n notiflex \
     -o custom-columns=NAME:.metadata.name,SCHEDULE:.spec.schedule,LAST:.status.lastScheduleTime
   ```
2. 이 Job이 무엇을 건드리는지 확인한다. `notiflex-healthcheck`는 읽기만 하므로 안전하지만, 외부 API를 호출하거나 데이터를 바꾸는 Job이면 영향 범위를 먼저 파악한다.
3. `concurrencyPolicy`가 `Forbid`인지 본다. 수동 Job은 CronJob과 별개라 이 정책이 적용되지 않는다.

## 실행

```bash
kubectl --context gke-sysnet4admin_book_gitaiops create job healthcheck-manual-$(date +%s) \
  --from=cronjob/notiflex-healthcheck -n notiflex
```

Pod 로그를 따라간다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops logs -f job/<JOB_NAME> -n notiflex
```

## 사후 검증

1. Job이 `Complete` 상태인지 확인한다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops get job -n notiflex
   ```
2. 로그에 "정상 (200)"이 찍혔는지 본다.
3. 수동으로 만든 Job은 CronJob의 히스토리 정리 대상이 아니므로 직접 지운다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops delete job <JOB_NAME> -n notiflex
   ```
