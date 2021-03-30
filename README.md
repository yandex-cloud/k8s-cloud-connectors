# Описание процесса

## Устанавливаем инструменты
- Установка kubebuilder
```shell
# Скачиваем себе kubebuilder и распаковываем его в нужную директорию
sudo mkdir /usr/local/kubebuilder
sudo curl -L https://go.kubebuilder.io/dl/2.3.1/$(go env GOOS)/$(go env GOARCH) | sudo tar -xz -C /usr/local/kubebuilder

# TODO: remove tar archive

# Работает на время этой сессии, лучше дописать эту команду в ~/.profile
export PATH=$PATH:/usr/local/kubebuilder/bin
```

EDIT: мигрировали на версию kubebuilder@3.0.0 (пока что она в бете, но тут очевидно пофикшены некоторые прошлый проблемы).
Чтобы последовать инструкции теперь, надо в ссылке заменить `2.3.1` на `latest`.

- Установка kustomize
```shell
 sudo curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
 sudo mv kustomize /usr/local/kubebuilder/bin
```
(я закидываю его к kubebuilder-у чтобы потом и удалять вместе + он так попадает в PATH)

Место где нужно немного уличной магии: kubebuilder сейчас использует `controller-gen@v0.2.5`, это очень устаревшая версия.
Она генерирует много Deprecated штук, и хотя они все рабочие, мы же хотим быть в тренде, верно? Идем в Makefile нашего проекта
(ну, уже после того, как создали проект) и меняем в директиве `controller-gen` версию на `controller-gen@v0.5.0`.

- Установка controller-gen:

После инициализации проекта выполняем `make controller-gen`.

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

# Собираем контроллер и пушим его в реестр (там настроено на мой реестр, надо поменять)
/bin/bash scripts/push-controller.sh ycr

# Устанавливаем в кластер CRD-шки
make install CONNECTOR="ycr"

# Создаём в кластере роль и привязываем её к сервисному аккаунту
kubectl apply -f ./connectors/ycr/config/rbac/role.yaml
kubectl apply -f ./connectors/ycr/config/rbac/role_binding.yaml

# Запускаем контроллер в кластере, смотрим его логи (опять же, надо поменять ссылку на образ)
/bin/bash scripts/run-controller.sh ycr

# Эту команду запускаем в другом окне терминала, в первом пишутся логи
# В этом файле надо по понятной причине поменять folderId на свой 
kubectl apply -f ./connectors/ycr/test-registry.yaml
# Смотрим на логи, видим, что реестр создался, ходим в веб-интерфейс, 
# смотрим, что он создался.
kubectl delete -f ./connectors/ycr/test-registry.yaml
# Опять смотрим в логи, смотрим что все счастливо удалилось.
# Останавливаем контроллер, удаляем CRD-шки и роли:
make uninstall CONNECTOR=ycr
kubectl delete -f ./connectors/ycr/config/rbac/role.yaml
kubectl delete -f ./connectors/ycr/config/rbac/role_binding.yaml
```