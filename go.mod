module github.com/yandex-cloud/k8s-cloud-connectors

go 1.15

exclude (
	github.com/yandex-cloud/k8s-cloud-connectors/examples/reporter v0.0.0
	github.com/yandex-cloud/k8s-cloud-connectors/scaffolder v0.0.0
)

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/aws/aws-sdk-go v1.38.21
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.2.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/jinzhu/copier v0.2.9
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/yandex-cloud/go-genproto v0.0.0-20210326132454-24349c492ce9
	github.com/yandex-cloud/go-sdk v0.0.0-20210326140609-dcebefcc0553
	go.uber.org/multierr v1.6.0
	go.uber.org/zap v1.17.0
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.0
)
