/*
Copyright (c) 2021 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"knative.dev/pkg/injection/sharedmain"

	"github.com/triggermesh/knative-sources/pkg/reconciler/httppollersource"
	"github.com/triggermesh/knative-sources/pkg/reconciler/slacksource"
	"github.com/triggermesh/knative-sources/pkg/reconciler/webhooksource"
	"github.com/triggermesh/knative-sources/pkg/reconciler/zendesksource"
)

func main() {
	sharedmain.Main("knative-sources-controller",
		slacksource.NewController,
		zendesksource.NewController,
		webhooksource.NewController,
		httppollersource.NewController,
	)
}
