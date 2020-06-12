package adapter

import (
	"knative.dev/eventing/pkg/adapter/v2"
)

// EnvAccesor for configuration parameters
func EnvAccesor() adapter.EnvConfigAccessor {
	return &envAccessor{}
}

type envAccessor struct {
	adapter.EnvConfig

	Threadiness int    `envconfig:"THREADINESS" default:"1"`
	Token       string `envconfig:"SLACK_TOKEN" required:"true"` //TODO::

}
