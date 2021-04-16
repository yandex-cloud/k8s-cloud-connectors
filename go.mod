module k8s-connectors

go 1.15

require (
	github.com/aws/aws-sdk-go v1.38.21
	github.com/go-logr/logr v0.3.0
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/jinzhu/copier v0.2.9
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/stretchr/testify v1.5.1
	github.com/yandex-cloud/go-genproto v0.0.0-20210326132454-24349c492ce9
	github.com/yandex-cloud/go-sdk v0.0.0-20210326140609-dcebefcc0553
	google.golang.org/grpc v1.28.0
	google.golang.org/protobuf v1.24.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.7.0
)
