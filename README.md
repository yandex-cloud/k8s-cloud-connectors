# Описание процесса

```shell
go mod init k8s-connectors
kubebuilder init --domain yandex.cloud.ru
kubebuilder create api --group yandex.cloud.ru --kind YandexObjectStorage --version v1
```
