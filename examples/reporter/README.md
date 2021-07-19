# Тестовый сценарий

## Описание сценария
Отправляете запрос `/report`, в `S3` кладётся файл с именем, переданным как аргумент `filename` и содержащий
сообщение, переданное в теле запроса и время обработки запроса.

- Сервер принимает POST-запросы;
- Кладёт таску в очередь;
- Рабочий мониторит таски в очереди, когда она появляется, забирает её и выполняет.

Все это реализовано внутри mk8s-кластера с помощью **YandexCloudConnectors**.

## Необходимые инструменты
* [mk8s-кластер](https://cloud.yandex.ru/services/managed-kubernetes) с установленными по [инструкции](../../README.md) **Yandex Cloud Connectors**;
* [Docker](https://www.docker.com) - для сборки образов и последующего размещения их в реестре образов;
* [kubectl](https://kubernetes.io/ru/docs/reference/kubectl/overview) - для управления объектами в кластере;
* [yc](https://cloud.yandex.ru/docs/cli/quickstart) - для управления ресурсами в Яндекс Облаке.
## Запуск

Для каждого этапа также будет указана команда из `Makefile`, которая исполняет его.

В первую очередь необходимо собрать наше приложение и положить его в какой-нибудь реестр образов, доступный кластеру (`make build-all REGISTRY=<your_registry>`):

```shell
REGISTRY=<your_registry>

docker build -t ${REGISTRY}/ycc-example/server:latest --file server.dockerfile .
docker push ${REGISTRY}/ycc-example/server:latest

docker build -t ${REGISTRY}/ycc-example/worker:latest --file worker.dockerfile .
docker push ${REGISTRY}/ycc-example/worker:latest
```

Затем в нашем фолдере новый сервисный аккаунт, который будет ответственным за это приложение и выдать ему права
`ymq.admin` и `storage.uploader` (`make create-sa FOLDER_ID=<your_folder_id>`):

```shell
FOLDER_ID=<your_folder_id>

SAID=$(yc iam service-account create ycc-example-sa --format json | jq -r '.id')
yc resource-manager folder add-access-binding --id "$FOLDER_ID" --role ymq.admin --service-account-id "$SAID"
yc resource-manager folder add-access-binding --id "$FOLDER_ID" --role storage.admin --service-account-id "$SAID"
```

Устанавливаем в кластер `yaml`-ы с нашим приложением (`make install SAID=$SAID REGISTRY=$REGISTRY`):

```shell
kubectl apply -f setup/ns.yaml
SAID=$SAID envsubst < setup/sakey.yaml.tmpl | kubectl apply -f -
sleep 1
kubectl apply -f setup/yos.yaml
kubectl apply -f setup/ycr.yaml
kubectl apply -f setup/ymq.yaml
sleep 1
REGISTRY=$REGISTRY envsubst < setup/server.yaml.tmpl | kubectl apply -f -
REGISTRY=$REGISTRY envsubst < setup/worker.yaml.tmpl | kubectl apply -f -
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

Теперь приберёмся за собой, удалив все созданные ресурсы (не забудьте очистить **Yandex Object Storage** от всех положенных в него файлов):

```shell
REGISTRY=$REGISTRY envsubst < setup/server.yaml.tmpl | kubectl delete -f -
REGISTRY=$REGISTRY envsubst < setup/worker.yaml.tmpl | kubectl delete -f -
kubectl delete -f setup/service.yaml
kubectl delete -f setup/yos.yaml
kubectl delete -f setup/ycr.yaml
kubectl delete -f setup/ymq.yaml
SAID=$SAID envsubst < setup/sakey.yaml.tmpl | kubectl delete -f -
kubectl delete -f setup/ns.yaml
```

Также это можно сделать командой `make uninstall`.

Затем удалим созданный сервисный аккаунт (`make delete-sa`):

```shell
yc iam service-account delete ycc-example-sa
```
