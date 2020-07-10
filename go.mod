module github.com/triggermesh/knative-sources

go 1.14

// Top-level module control over the exact version used for important direct dependencies.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8
)

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/nukosuke/go-zendesk v0.7.5
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/text v0.3.3 // indirect
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.15.1
	knative.dev/pkg v0.0.0-20200619214325-bfb590781363
	knative.dev/serving v0.15.1
)
