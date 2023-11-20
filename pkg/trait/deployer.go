/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trait

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	traitv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1/trait"
	"github.com/apache/camel-k/v2/pkg/util/patch"
)

type deployerTrait struct {
	BasePlatformTrait
	traitv1.DeployerTrait `property:",squash"`
}

var _ ControllerStrategySelector = &deployerTrait{}

var hasServerSideApply = true

func newDeployerTrait() Trait {
	return &deployerTrait{
		BasePlatformTrait: NewBasePlatformTrait("deployer", 900),
	}
}

func (t *deployerTrait) Configure(e *Environment) (bool, *TraitCondition, error) {
	return e.Integration != nil, nil, nil
}

func (t *deployerTrait) Apply(e *Environment) error {
	// Register a post action that patches the resources generated by the traits
	e.PostActions = append(e.PostActions, func(env *Environment) error {
		for _, resource := range env.Resources.Items() {
			// We assume that server-side apply is enabled by default.
			// It is currently convoluted to check proactively whether server-side apply
			// is enabled. This is possible to fetch the OpenAPI endpoint, which returns
			// the entire server API document, then lookup the resource PATCH endpoint, and
			// check its list of accepted MIME types.
			// As a simpler solution, we fall back to client-side apply at the first
			// 415 error, and assume server-side apply is not available globally.
			if hasServerSideApply && pointer.BoolDeref(t.UseSSA, true) {
				err := t.serverSideApply(env, resource)
				switch {
				case err == nil:
					continue
				case isIncompatibleServerError(err):
					t.L.Info("Fallback to client-side apply to patch resources")
					hasServerSideApply = false
				default:
					// Keep server-side apply unless server is incompatible with it
					return err
				}
			}
			if err := t.clientSideApply(env, resource); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

func (t *deployerTrait) serverSideApply(env *Environment, resource ctrl.Object) error {
	target, err := patch.ApplyPatch(resource)
	if err != nil {
		return err
	}
	fmt.Println("********* Applying patch", target, "on resource", resource)
	err = env.Client.Patch(env.Ctx, target, ctrl.Apply, ctrl.ForceOwnership, ctrl.FieldOwner("camel-k-operator"))
	if err != nil {
		return fmt.Errorf("error during apply resource: %s/%s: %w", resource.GetNamespace(), resource.GetName(), err)
	}
	// Update the resource with the response returned from the API server
	return t.unstructuredToRuntimeObject(target, resource)
}

func (t *deployerTrait) clientSideApply(env *Environment, resource ctrl.Object) error {
	err := env.Client.Create(env.Ctx, resource)
	if err == nil {
		return nil
	} else if !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("error during create resource: %s/%s: %w", resource.GetNamespace(), resource.GetName(), err)
	}
	object := &unstructured.Unstructured{}
	object.SetNamespace(resource.GetNamespace())
	object.SetName(resource.GetName())
	object.SetGroupVersionKind(resource.GetObjectKind().GroupVersionKind())
	err = env.Client.Get(env.Ctx, ctrl.ObjectKeyFromObject(object), object)
	if err != nil {
		return err
	}
	p, err := patch.MergePatch(object, resource)
	if err != nil {
		return err
	} else if len(p) == 0 {
		// Update the resource with the object returned from the API server
		return t.unstructuredToRuntimeObject(object, resource)
	}
	err = env.Client.Patch(env.Ctx, resource, ctrl.RawPatch(types.MergePatchType, p))
	if err != nil {
		return fmt.Errorf("error during patch %s/%s: %w", resource.GetNamespace(), resource.GetName(), err)
	}
	return nil
}

func (t *deployerTrait) unstructuredToRuntimeObject(u *unstructured.Unstructured, obj ctrl.Object) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}

func isIncompatibleServerError(err error) bool {
	// First simpler check for older servers (i.e. OpenShift 3.11)
	if strings.Contains(err.Error(), "415: Unsupported Media Type") {
		return true
	}

	// 415: Unsupported media type means we're talking to a server which doesn't
	// support server-side apply.
	var serr *k8serrors.StatusError
	if errors.As(err, &serr) {
		return serr.Status().Code == http.StatusUnsupportedMediaType
	}

	// Non-StatusError means the error isn't because the server is incompatible.
	return false
}

func (t *deployerTrait) SelectControllerStrategy(e *Environment) (*ControllerStrategy, error) {
	if t.Kind != "" {
		strategy := ControllerStrategy(t.Kind)
		return &strategy, nil
	}
	return nil, nil
}

func (t *deployerTrait) ControllerStrategySelectorOrder() int {
	return 0
}

// RequiresIntegrationPlatform overrides base class method.
func (t *deployerTrait) RequiresIntegrationPlatform() bool {
	return false
}

func (t *deployerTrait) Reverse(e *Environment, traits *v1.Traits) error {
	// Register a post action that patches the resources generated by the traits
	e.PostActions = append(e.PostActions, func(env *Environment) error {
		for _, resource := range env.Resources.Items() {
			err := env.Client.Update(env.Ctx, resource, ctrl.FieldOwner("camel-k-operator"))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}
