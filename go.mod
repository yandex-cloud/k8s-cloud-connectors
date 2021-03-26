module k8s-connectors

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/yandex-cloud/go-genproto v0.0.0-20210322103648-06fdc502f726
	github.com/yandex-cloud/go-sdk v0.0.0-20210316121032-42388ab29215
	google.golang.org/grpc v1.28.0
	google.golang.org/protobuf v1.26.0 // indirect
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
