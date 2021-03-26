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

- Установка kustomize
```shell
 sudo curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
 sudo mv kustomize /usr/local/kubebuilder/bin
```
(я закидываю его к kubebuilder-у чтобы потом и удалять вместе + он так попадает в PATH)

- Установка controller-gen:

После установки kubebuilder выполняем `make controller-gen`.

## Проверяем работоспособность
Перед тем как все делать надо сходить в веб-интерфейс
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
/bin/bash scripts/push-controller.sh

# Устанавливаем в кластер CRD-шки
kustomize build config/crd | kubectl apply -f -

# Создаём в кластере роль и привязываем её к сервисному аккаунту
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/role_binding.yaml

# Запускаем контроллер в кластере, смотрим его логи (опять же, надо поменять ссылку на образ)
/bin/bash scripts/run-controller.sh

# Эту команду запускаем в другом окне терминала, в первом пишутся логи
# В этом файле надо по понятной причине поменять folderId на свой 
kubectl apply -f test-registry.yaml
# Смотрим на логи, видим, что реестр создался, ходим в веб-интерфейс, 
# смотрим, что он создался.
kubectl delete -f test-registry.yaml
# Опять смотрим в логи, смотрим что все счастливо удалилось.
# Останавливаем контроллер, удаляем CRD-шки и роли:
kustomize build config/crd | kubectl delete -f -
kubectl delete -f config/rbac/role.yaml
kubectl delete -f config/rbac/role_binding.yaml
```