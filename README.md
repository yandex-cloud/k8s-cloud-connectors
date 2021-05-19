# Описание процесса

## Установка на свой кластер

Проверяем работоспособность контроллеров на Container Registry.

```shell
# Для начала надо дать сервисному аккаунту, который менеджит инстансы, права делать вещи с реджистри
# На самом деле можно давать не админские, но пока что так будет прощу, потом пофиксим
folder_id=b1g7jvgmf06eel94s22d # Подставь свой
nodegroup_id=catnmbc81hdag58trmgi # Подставь свой

instance_group_id=$(yc managed-kubernetes node-group get --id ${nodegroup_id} --format json | jq -r ".instance_group_id")
service_account_id=$(yc compute instance-group get --id ${instance_group_id} --format json | jq -r ".instance_template.service_account_id")
yc resource-manager folder add-access-binding --id $folder_id --role container-registry.admin --service-account-id $service_account_id

# Выполняем эти команды из корня репозитория

# Собираем контроллер и пушим его в реестр (стоит подставить свой реестр)
make docker-push REGISTRY=cr.yandex/crptp7j81e7caog8r6gq

# Добавляем в кластер все необходимые сущности и контроллер
make install

# Эту команду запускаем в другом окне терминала, в первом пишутся логи
folder_id=$folder_id envsubst < ./connectors/ycr/examples/test-registry.yaml.tmpl | kubectl apply -f -

# Смотрим на логи, видим, что реестр создался, ходим в веб-интерфейс, 
# смотрим, что он создался.
folder_id=$folder_id envsubst < ./connectors/ycr/examples/test-registry.yaml.tmpl | kubectl delete -f -
# Опять смотрим в логи, смотрим что все счастливо удалилось.
# Останавливаем контроллер, удаляем CRD-шки и роли:
make uninstall
```