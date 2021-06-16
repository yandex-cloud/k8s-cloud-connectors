module k8s-connectors

go 1.15

require (
	github.com/aws/aws-sdk-go v1.38.21
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.2.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/jinzhu/copier v0.2.9
	github.com/stretchr/testify v1.6.1
	github.com/yandex-cloud/go-genproto v0.0.0-20210326132454-24349c492ce9
	github.com/yandex-cloud/go-sdk v0.0.0-20210326140609-dcebefcc0553
	go.uber.org/zap v1.15.0
	golang.org/x/tools v0.1.1 // indirect
	google.golang.org/grpc v1.28.0
	google.golang.org/protobuf v1.25.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.0
)
