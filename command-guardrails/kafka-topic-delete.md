# Kafka Topic 삭제

되돌릴 수 없는 작업이다. 삭제한 토픽의 메시지는 복구되지 않는다.

## 사전 확인

1. 토픽에 아직 처리하지 않은 메시지가 있는지 본다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops -n kafka exec notiflex-kafka-dual-role-0 -- \
     bin/kafka-get-offsets.sh --bootstrap-server localhost:9092 --topic <TOPIC>
   ```
2. Consumer가 끝까지 읽었는지 확인한다. Notiflex API Pod 로그에서 마지막 offset을 본다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops logs -l app=notiflex-api -n notiflex --tail=50 | grep "Kafka 수신"
   ```
3. 이 토픽으로 보내는 Producer를 모두 찾는다. 현재는 `notiflex`와 `enterprise` 두 네임스페이스의 API가 보낸다.

## 실행

1. Producer를 먼저 멈춘다. Rollout의 `KAFKA_BROKER` 환경변수를 지우고 Git에 커밋해 ArgoCD가 반영하게 한다.
2. Consumer가 남은 메시지를 다 읽을 때까지 기다린다.
3. 매니페스트에서 KafkaTopic을 지우고 커밋한다. `kubectl delete`로 직접 지우지 않는다. ArgoCD selfHeal이 되돌린다.
   ```bash
   git rm k8s/kafka/notifications-topic.yaml
   git commit -m "chore: notifications 토픽 제거"
   git push origin main
   ```

## 사후 검증

1. 토픽이 사라졌는지 확인한다.
   ```bash
   kubectl --context gke-sysnet4admin_book_gitaiops get kafkatopic -n kafka
   ```
2. API Pod 로그에 Kafka 관련 오류가 없는지 본다.
3. ArgoCD `kafka` Application이 Synced/Healthy인지 확인한다.
