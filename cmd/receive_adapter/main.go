package main

import (
	"knative.dev/eventing/pkg/adapter/v2"

	slackadapter "github.com/triggermesh/knative-slack-source/pkg/adapter"
)

func main() {
	adapter.Main("slack", slackadapter.EnvAccesor, slackadapter.New)
}
