# Тестовый сценарий

## Описание сценария
Отправляете приложению `POST`-запрос `/report`, в `S3` кладётся файл, содержащий
сообщение, переданное как аргумент запроса `message` и время обработки запроса.

- Сервер принимает POST-запросы;
- Кладёт таску в очередь;
- Рабочий мониторит таски в очереди, когда она появляется, забирает её и выполняет.

Все это реализовано внутри k8s-кластера с помощью **YandexCloudConnectors**.

## Запуск

В первую очередь необходимо собрать наше приложение и положить его в какой-нибудь реестр образов, доступный кластеру:

```shell
REGISTRY=cr.yandex/crptp7j81e7caog8r6gq

docker build -t ${REGISTRY}/ycc-example-server:latest --file server.dockerfile .
docker push ${REGISTRY}/ycc-example-server:latest

docker build -t ${REGISTRY}/ycc-example-worker:latest --file worker.dockerfile .
docker push ${REGISTRY}/ycc-example-worker:latest
```

*Не забудьте подставить этот реестр в `server.yaml` и `worker.yaml`, чтобы `Deployment`-ы брали его откуда нужно.*

Затем в нашем фолдере новый сервисный аккаунт, который будет ответственным за это приложение и выдать ему права `ymq.editor` и `storage.uploader`.

После чего устанавливаем в кластер (предполагается, что он уже настроен, и его данные лежат в `~/.kube/config`) **YandexCloudConnectors**:

```shell
TODO когда у нас появится итоговая инструкция, надо вставить её сюда
```

Ждём, пока *pod* с менеджером перейдёт в состояние `Running`, затем применяем к кластеру `yaml`-ы с нашим приложением (не забывайте проставить свой реестр в подах):

```shell
kubectl apply -f setup/ns.yaml

kubectl apply -f setup/sakey.yaml
kubectl apply -f setup/ycr.yaml
kubectl apply -f setup/yos.yaml
kubectl apply -f setup/ymq.yaml

kubectl apply -f setup/service.yaml
kubectl apply -f setup/server.yaml
kubectl apply -f setup/worker.yaml
```

Теперь можно сделать 