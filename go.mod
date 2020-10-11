module github.com/triggermesh/knative-sources

go 1.15

// Top-level module control over the exact version used for important direct dependencies.
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
replace (
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
)

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/google/go-cmp v0.5.2
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/nukosuke/go-zendesk v0.7.7
	github.com/oracle/oci-go-sdk v24.3.0+incompatible
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.18.0
	knative.dev/pkg v0.0.0-20200922164940-4bf40ad82aab
	knative.dev/serving v0.18.0
)
