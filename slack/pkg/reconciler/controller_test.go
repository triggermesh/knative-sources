/*
Copyright (c) 2020 TriggerMesh Inc.

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

package reconciler

import (
	"reflect"
	"testing"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	_ "knative.dev/pkg/metrics/testing"
	rt "knative.dev/pkg/reconciler/testing"

	"github.com/stretchr/testify/assert"
	st "github.com/triggermesh/knative-sources/pkg/reconciler/testing"

	// Link fake informers accessed by our controller
	_ "github.com/triggermesh/knative-sources/pkg/client/generated/injection/informers/sources/v1alpha1/slacksource/fake"
	_ "knative.dev/pkg/client/injection/ducks/duck/v1/addressable/fake"
	_ "knative.dev/pkg/injection/clients/dynamicclient/fake"
	_ "knative.dev/serving/pkg/client/injection/informers/serving/v1/service/fake"
)

const (
	ImageEnv = "SLACKSOURCE_ADAPTER_IMAGE"
	ImageVal = "triggermesh.registry/slack-adapter:v0.0.0"
)

func TestNewController(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	ctx, informers := rt.SetupFakeContext(t)

	// expected informers: Target, Knative Service
	if expect, got := 2, len(informers); got != expect {
		t.Errorf("Expected %d injected informers, got %d", expect, got)
	}

	// required envconfig env vars
	defer st.SetEnvVar(t, ImageEnv, ImageVal)()

	cmw := configmap.NewStaticWatcher(
		st.NewConfigMap(logging.ConfigMapName(), nil),
		st.NewConfigMap(metrics.ConfigMapName(), nil),
	)

	ctrler := NewController(ctx, cmw)

	// catch unitialized fields in Reconciler struct
	ensureNoNilField(t, ctrler)
}

func TestNewControllerFailures(t *testing.T) {
	testCases := map[string]struct {
		initFn   func(**configmap.StaticWatcher) (undo func())
		assertFn func(*testing.T, assert.PanicTestFunc)
	}{
		"Fails when watching missing ConfigMaps": {
			initFn: func(cmw **configmap.StaticWatcher) func() {
				*cmw = configmap.NewStaticWatcher()

				undoEnv := st.SetEnvVar(t, ImageEnv, ImageVal)
				return func() {
					undoEnv()
				}
			},
			assertFn: func(t *testing.T, testFn assert.PanicTestFunc) {
				assert.PanicsWithValue(t, `Tried to watch unknown config with name "config-logging"`, testFn)
			},
		},
		"Fails when mandatory env var is missing": {
			initFn: func(cmw **configmap.StaticWatcher) func() {
				*cmw = configmap.NewStaticWatcher(
					st.NewConfigMap(logging.ConfigMapName(), nil),
					st.NewConfigMap(metrics.ConfigMapName(), nil),
				)

				return func() {}
			},
			assertFn: func(t *testing.T, testFn assert.PanicTestFunc) {
				assert.PanicsWithValue(t, "required key "+ImageEnv+" missing value", testFn)
			},
		},
	}

	for n, tc := range testCases {
		//nolint:scopelint
		t.Run(n, func(t *testing.T) {
			cmw := &configmap.StaticWatcher{}

			undo := tc.initFn(&cmw)
			if undo != nil {
				defer undo()
			}

			ctx, _ := rt.SetupFakeContext(t)

			tc.assertFn(t, func() {
				_ = NewController(ctx, cmw)
			})
		})
	}
}

// ensureNoNilField fails the test if the provided Impl's reconciler contains
// nil pointers or interfaces.
func ensureNoNilField(t *testing.T, impl *controller.Impl) {
	t.Helper()

	recVal := reflect.ValueOf(impl.Reconciler).Elem().
		FieldByName("reconciler"). // knative.dev/pkg/controller.Reconciler
		Elem().                    // injection/reconciler/sources/v1alpha1/<type>.Interface
		Elem()                     //*reconciler.Reconciler

	for i := 0; i < recVal.NumField(); i++ {
		f := recVal.Field(i)
		switch f.Kind() {
		case reflect.Interface, reflect.Ptr, reflect.Func:
			if f.IsNil() {
				t.Errorf("struct field %q is nil", recVal.Type().Field(i).Name)
			}
		}
	}
}
