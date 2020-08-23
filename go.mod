module github.com/triggermesh/knative-sources

go 1.14

// Top-level module control over the exact version used for important direct dependencies.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
)

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/nukosuke/go-zendesk v0.7.7
	github.com/stretchr/testify v1.6.1
	github.com/triggermesh/pkg v0.0.0-20200823085210-71cdd7af8c1a
	go.uber.org/zap v1.15.0
	k8s.io/api v0.18.7-rc.0
	k8s.io/apimachinery v0.18.7-rc.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.17.0
	knative.dev/pkg v0.0.0-20200812224206-44c860147a87
	knative.dev/serving v0.17.0
)
