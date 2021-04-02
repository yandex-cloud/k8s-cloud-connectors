# Описание процесса

## Устанавливаем инструменты разработчика
Важно: эта часть не нужна конечному пользователю. Это вообще нужно только для того, чтобы 

- Установка kubebuilder
```shell
# Скачиваем себе kubebuilder и распаковываем его в нужную директорию
os=$(go env GOOS)
arch=$(go env GOARCH)

# download kubebuilder and extract it to tmp
curl -L https://go.kubebuilder.io/dl/2.3.1/${os}/${arch} | tar -xz -C /tmp/

# move to a long-term location and put it on your path
# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
sudo mkdir /usr/bin/kubebuilder
sudo mv /tmp/kubebuilder_2.3.1_${os}_${arch} /usr/local/kubebuilder

# Работает на время этой сессии, лучше дописать эту команду в ~/.profile
export PATH=$PATH:/usr/local/kubebuilder/bin
```

К сожалению, для версии 3.0.0, которую я использую, пока что нет способа загрузки через командную строку. Релизы
можно найти [здесь](https://github.com/kubernetes-sigs/kubebuilder/releases), но стоит принять во внимание, 
что в пакеты поставки для beta-версий не включены дополнительные компоненты для тестирования.
Их можно использовать из старой версии (в папке со старой версией заменить только бинарник kubebuilder-а). 

**P.S.:** для того, чтобы использовать существующее решение, установка kubebuilder не требуется. Он используется только для тестов,
а они пока нам не очень нужны. По этой же причине я отключил (временно) тестирование в Makefile. Конечному пользователю
они тоже не нужны, так что он вообще может обойтись без kubebuilder-а.

## Проверяем работоспособность
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
/bin/bash scripts/push-controller.sh ycr cr.yandex/crptp7j81e7caog8r6gq

# Устанавливаем в кластер CRD-шки
make install CONNECTOR=ycr

# Создаём в кластере роль и привязываем её к сервисному аккаунту
kubectl apply -f ./connectors/ycr/config/rbac/role.yaml
kubectl apply -f ./connectors/ycr/config/rbac/role_binding.yaml

# Запускаем контроллер в кластере, смотрим его логи (опять же, надо подставить свой реестр)
/bin/bash scripts/run-controller.sh ycr cr.yandex/crptp7j81e7caog8r6gq

# Эту команду запускаем в другом окне терминала, в первом пишутся логи
folder_id=$folder_id envsubst < ./connectors/ycr/examples/test-registry.yaml.tmpl | kubectl apply -f -

# Смотрим на логи, видим, что реестр создался, ходим в веб-интерфейс, 
# смотрим, что он создался.
folder_id=$folder_id envsubst < ./connectors/ycr/examples/test-registry.yaml.tmpl | kubectl delete -f -
# Опять смотрим в логи, смотрим что все счастливо удалилось.
# Останавливаем контроллер, удаляем CRD-шки и роли:
make uninstall CONNECTOR=ycr
kubectl delete -f ./connectors/ycr/config/rbac/role.yaml
kubectl delete -f ./connectors/ycr/config/rbac/role_binding.yaml
```