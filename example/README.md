# Тестовый сценарий

## Описание сценария
Отправляете запрос `/report`, в `S3` кладётся файл с именем, переданным как аргумент `filename` и содержащий
сообщение, переданное в теле запроса и время обработки запроса.

- Сервер принимает POST-запросы;
- Кладёт таску в очередь;
- Рабочий мониторит таски в очереди, когда она появляется, забирает её и выполняет.

Все это реализовано внутри k8s-кластера с помощью **YandexCloudConnectors**.

## Запуск

В первую очередь необходимо собрать наше приложение и положить его в какой-нибудь реестр образов, доступный кластеру (`make build-all REGISTRY=cr.yandex/crptp7j81e7caog8r6gq`):

```shell
REGISTRY=cr.yandex/crptp7j81e7caog8r6gq

docker build -t ${REGISTRY}/ycc-example/server:latest --file server.dockerfile .
docker push ${REGISTRY}/ycc-example/server:latest

docker build -t ${REGISTRY}/ycc-example/worker:latest --file worker.dockerfile .
docker push ${REGISTRY}/ycc-example/worker:latest
```

*P.S.: в дальнейшем эти образы будут лежать в общем реестре Yandex Cloud*

*Не забудьте подставить этот реестр в `server.yaml` и `worker.yaml`, чтобы `Deployment`-ы брали образ откуда нужно.*

Затем в нашем фолдере новый сервисный аккаунт, который будет ответственным за это приложение и выдать ему права `ymq.editor` и `storage.uploader`.

После чего устанавливаем в кластер (предполагается, что он уже настроен) **YandexCloudConnectors**:

```shell
TODO когда у нас появится итоговая инструкция, надо вставить её сюда
```

Ждём, пока *pod* с менеджером перейдёт в состояние `Running`, например, так (`make wait-for-ycc`):

```shell
#!/bin/bash

until [ "$(kubectl -n yandex-cloud-connectors get pod -l control-plane=connector-manager --output=json | jq '.items[0].status.phase')" = '"Running"' ]; do
  echo "Not yet Running"
  sleep 1
done
```

Затем применяем к кластеру `yaml`-ы с нашим приложением, не забывая проставить свой реестр в подах (`make install`):

```shell
kubectl apply -f setup/ns.yaml
kubectl apply -f setup/sakey.yaml
sleep 1
kubectl apply -f setup/yos.yaml
kubectl apply -f setup/ycr.yaml
kubectl apply -f setup/ymq.yaml
sleep 1
kubectl apply -f setup/server.yaml
kubectl apply -f setup/worker.yaml
kubectl apply -f setup/service.yaml
```

Теперь можно проверить работоспособность этого подхода. Сначала узнаем внешний **IP** вашего кластера, например,
такой командой:

```shell
CLUSTER_ENDPOINT=$(kubectl -n yandex-cloud-connectors-example get service/image-reporter --output=json | jq '.status.loadBalancer.ingress[0].ip' | tr -d '"')
```

Отправим нашему серверу запрос (`make make-sample-request`):

```shell
curl -X POST -d "Hello Yandex Cloud Connectors!" "${CLUSTER_ENDPOINT}/report?filename=greetings.txt"
```

Заглядываем в веб-интерфейс **Yandex Object Storage** и видим появившийся файл!
