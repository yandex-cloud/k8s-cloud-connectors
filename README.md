# Yandex Cloud Connectors

**Yandex Cloud Connectors** это инструмент, позволяющий интегрировать взаимодействие с ресурсами **Yandex Cloud** 
в процесс работы с **kubernetes**, позволяя отказаться от использования дополнительных инструментов и расширяя возможности
по автоматизации процессов в кластере.

**Yandex Cloud Connectors** используют встроенный в **Kubernetes** control-loop, сохраняя желаемые состояния
облачных ресурсов как объекты в **k8s api**.

## Установка на свой кластер

#### Необходимые инструменты
1. Настроенный кластер [**Kubernetes**](https://kubernetes.io) - [mk8s](https://cloud.yandex.ru/services/managed-kubernetes) в Яндекс Облаке.
2. Установленный [__Helm__](https://helm.sh). Для установки из реестра необходима версия 3.7+

Для установки **Yandex Cloud Connectors** из реестра:
```
export HELM_EXPERIMENTAL_OCI=1
export HELM_REGISTRY_CONFIG="$HOME/.docker/config.json"
helm install yandex-cloud-connectors oci://cr.yandex/yc/cloud-connectors/chart/yandex-cloud-connectors --version <version>
```
список доступных версий можно найти здесь: https://github.com/yandex-cloud/k8s-cloud-connectors/tags

Для установки **Yandex Cloud Connectors** из репозитория (latest версия образов):

```shell
helm install yandex-cloud-connectors helm/yandex-cloud-connectors
```

## Пример использования

*Для этого примера помимо вышеуказанных зависимостей необходимо установить следующие командные утилиты:*
* [*kubectl*](https://kubernetes.io/ru/docs/reference/kubectl/overview)
* [*yc*](https://cloud.yandex.ru/docs/cli/quickstart)

_Более развернутый пример, демонстрирующий больше возможностей **YCC**, находится в этом репозитории, в [папке "examples"](./examples)._

Покажем пример работы **YCC** на [Yandex Container Registry](https://cloud.yandex.ru/services/container-registry).
Для начала нам надо дать права на работу с Container Registry сервисному аккаунту,
под управлением которого работают ноды в нашем кластере:

```shell
FOLDER_ID=<your_folder_id>
CLUSTER_ID=<your_cluster_id>

SERVICE_ACCOUNT_ID=$(yc k8s cluster get "$CLUSTER_ID" --format json | jq -r '.node_service_account_id')
yc resource-manager folder add-access-binding --id "$FOLDER_ID" --role container-registry.admin --service-account-id "$SERVICE_ACCOUNT_ID"
```

Теперь у нод кластера есть права администрировать Yandex Container Registry в облаке. Теперь установим контроллер в кластер, добавляем в кластер все необходимые сущности и контроллер:

```shell
helm install yandex-cloud-connectors helm/yandex-cloud-connectors
```

Теперь попробуем создать какой-нибудь облачный ресурс, например, **Yandex Container Registry**:

```shell
FOLDER_ID="$FOLDER_ID" envsubst < ./examples/test-registry.yaml.tmpl | kubectl apply -f -
```

Можно зайти в UI облака и посмотреть, что там создался новый Container Registry. Аналогичную проверку можно выполнить
с помощью консольной команды `yc container registry list`. Теперь удалим Registry из кластера:

```shell
FOLDER_ID="$FOLDER_ID" envsubst < ./examples/test-registry.yaml.tmpl | kubectl delete -f -
```

Повторно сходив в веб-интерфейс или исполнив команду `yc container registry list` можно увидеть, что реестр удалён.

Чтобы удалить **YCC** из кластера, достаточно выполнить команду:

```shell
helm uninstall yandex-cloud-connectors
```
