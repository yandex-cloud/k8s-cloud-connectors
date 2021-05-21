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

Теперь у нод кластера есть права администрировать Container Registry в облаке. Теперь устанавливаем контроллер в кластер:

```shell
# Добавляем в кластер все необходимые сущности и контроллер
make install

# Эту команду запускаем в другом окне терминала, в первом пишутся логи
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