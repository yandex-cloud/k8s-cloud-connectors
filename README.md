# Описание процесса

## Установка на свой кластер

Проверим работоспособность коннекторов на Container Registry. Для начала нам надо дать права на работу с Container Registry
сервисному аккаунту, под управлением которого работают ноды в нашем кластере.

Это можно сделать через UI:

```Кластер -> Обзор -> Сервисный аккаунт для узлов -> Редактировать аккаунт -> Роли в каталоге -> '+' -> 'container-registry.admin'```

Того же самого результата можно добиться из консоли:

```shell
folder_id=b1g7jvgmf06eel94s22d # Подставить свой
nodegroup_id=catnmbc81hdag58trmgi # Подставить свой свой

instance_group_id=$(yc managed-kubernetes node-group get --id ${nodegroup_id} --format json | jq -r ".instance_group_id")
service_account_id=$(yc compute instance-group get --id ${instance_group_id} --format json | jq -r ".instance_template.service_account_id")
yc resource-manager folder add-access-binding --id $folder_id --role container-registry.admin --service-account-id $service_account_id
```

Теперь у нод кластера есть права администрировать Container Registry в облаке. Теперь установим контроллер в кластер.

```shell
# Добавляем в кластер все необходимые сущности и контроллер
make install
```
Можно заметить, что все сущности появились, однако под повис в состоянии Container Creating. Это происходит из-за того,
что он не может получить TLS сертификат, чтобы контроллер мог использовать механизм вебхуков. Скрипт, выписывающий этот
сертификат, взят из [этого репозитория](https://github.com/morvencao/kube-mutating-webhook-tutorial).

```shell
./scripts/webhook-create-signed-cert.sh --namespace yandex-cloud-connectors \
                                        --service webhook-service \
                                        --secret webhook-tls-cert

kubectl -n yandex-cloud-connectors get secret webhook-tls-cert -o json | jq '.data["tls.crt"]' | tr -d '"'
```

Полученный при помощи второй команды ключ вставляем в [patch.yaml](config/webhook/patch.yaml) вместо `${CA_BUNDLE}`.
Теперь повторяем инсталляцию, kustomize сам пропатчит нужный ресурс:

```shell
make install

folder_id=$folder_id envsubst < ./connectors/ycr/examples/test-registry.yaml.tmpl | kubectl apply -f -
```

Можно зайти в UI облака и посмотреть, что там создался новый Container Registry. Аналогичную проверку можно выполнить
с помощью консольной команды `yc container registry list`. Теперь удалим Registry и деинсталлируем из кластера 
контроллер:

```shell
folder_id=$folder_id envsubst < ./connectors/ycr/examples/test-registry.yaml.tmpl | kubectl delete -f -
make uninstall
```

Повторно сходив в веб-интерфейс или исполнив команду `yc container registry list` можно увидеть, что Registry
удалён.