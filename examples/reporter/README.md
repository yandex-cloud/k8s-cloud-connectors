# Reporter

Тестовый сценарий для демонстрации работы **YandexCloudConnectors** внутри mk8s-кластера.

## Описание сценария

При помощи **Yandex CLoud Connectors** в облаке создаются:
* Очередь сообщений **Yandex Message Queue**; 
* Файловое хранилище **Yandex Object Storage**;
* **Static Access Key** для доступа к ним.

Для каждого из вышеперечисленных ресурсов **YCC** создает `Secret` или `ConfigMap`, хранящий необходимые для доступа к этому ресурсу
данные. Они используются внутри приложения и подключаются к нему как `Volume`-ы.

В кластере поднимается сервис, который при получении запроса `/report`, кладёт в **Yandex Object Storage**
файл с именем, переданным как аргумент `filename` и содержащий сообщение, переданное в теле запроса и время обработки запроса.

Внутри сервис состоит из сервера и рабочего, которые работают следующим образом:
- Сервер принимает POST-запросы;
- Кладёт таску в очередь;
- Рабочий мониторит таски в очереди, когда она появляется, забирает её и выполняет.

## Необходимые инструменты
* [mk8s-кластер](https://cloud.yandex.ru/services/managed-kubernetes) с установленными по [инструкции](../../README.md) **Yandex Cloud Connectors**;
* [Docker](https://www.docker.com) - для сборки образов и последующего размещения их в реестре образов;
* [kubectl](https://kubernetes.io/ru/docs/reference/kubectl/overview) - для управления объектами в кластере;
* [yc](https://cloud.yandex.ru/docs/cli/quickstart) - для управления ресурсами в Яндекс Облаке.

## Запуск

В первую очередь необходимо собрать наше приложение и положить его в какой-нибудь реестр образов, доступный кластеру:

```shell
REGISTRY=cr.yandex/<your_registry_id>

docker build -t ${REGISTRY}/ycc-example/server:latest --file server.dockerfile .
docker push ${REGISTRY}/ycc-example/server:latest

docker build -t ${REGISTRY}/ycc-example/worker:latest --file worker.dockerfile .
docker push ${REGISTRY}/ycc-example/worker:latest
```

Затем создадим в нашем фолдере новый сервисный аккаунт, который будет ответственным за это приложение и выдать ему права
`ymq.admin` и `storage.uploader`:

```shell
FOLDER_ID=<your_folder_id>

SAID=$(yc iam service-account create ycc-example-sa --format json | jq -r '.id')
yc resource-manager folder add-access-binding --id "$FOLDER_ID" --role ymq.admin --service-account-id "$SAID"
yc resource-manager folder add-access-binding --id "$FOLDER_ID" --role storage.admin --service-account-id "$SAID"
```

Устанавливаем в кластер `yaml`-ы с нашим приложением:

```shell
kubectl apply -f setup/ns.yaml
SAID=$SAID envsubst < setup/sakey.yaml.tmpl | kubectl apply -f -
sleep 1
kubectl apply -f setup/yos.yaml
kubectl apply -f setup/ymq.yaml
sleep 1
REGISTRY=$REGISTRY envsubst < setup/server.yaml.tmpl | kubectl apply -f -
REGISTRY=$REGISTRY envsubst < setup/worker.yaml.tmpl | kubectl apply -f -
kubectl apply -f setup/service.yaml
```

Проверим, что все работает как надо, получив список ресурсов в тестовом `Namespace`:

```shell
kubectl -n yandex-cloud-connectors-example get all
```

Вы должны увидеть, что поды обоих `Deployment`-ов, `server` и `worker`, перешли в состояние `Running`. 

Теперь можно проверить работоспособность всей системы. Сначала узнаем внешний **IP** вашего кластера, например,
такой командой:

```shell
CLUSTER_ENDPOINT=$(kubectl -n yandex-cloud-connectors-example get service/image-reporter --output=json | jq -r '.status.loadBalancer.ingress[0].ip')
```

Отправим нашему серверу запрос:

```shell
curl -X POST -d 'Hello Yandex Cloud Connectors!' "${CLUSTER_ENDPOINT}/report?filename=greetings.txt"
```

Заглядываем в веб-интерфейс **Yandex Object Storage** и видим появившийся файл!

Теперь приберёмся за собой, удалив все созданные ресурсы (не забудьте очистить **Yandex Object Storage** от всех положенных в него файлов):

```shell
REGISTRY=$REGISTRY envsubst < setup/server.yaml.tmpl | kubectl delete -f -
REGISTRY=$REGISTRY envsubst < setup/worker.yaml.tmpl | kubectl delete -f -
kubectl delete -f setup/service.yaml
kubectl delete -f setup/yos.yaml
kubectl delete -f setup/ymq.yaml
SAID=$SAID envsubst < setup/sakey.yaml.tmpl | kubectl delete -f -
kubectl delete -f setup/ns.yaml
```

Затем удалим созданный сервисный аккаунт:

```shell
yc iam service-account delete ycc-example-sa
```
