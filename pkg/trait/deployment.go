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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	traitv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1/trait"
	"github.com/apache/camel-k/v2/pkg/util/kubernetes"
)

type deploymentTrait struct {
	BasePlatformTrait
	traitv1.DeploymentTrait `property:",squash"`
}

var _ ControllerStrategySelector = &deploymentTrait{}

func newDeploymentTrait() Trait {
	return &deploymentTrait{
		BasePlatformTrait: NewBasePlatformTrait("deployment", 1100),
	}
}

func (t *deploymentTrait) Configure(e *Environment) (bool, *TraitCondition, error) {
	if !e.IntegrationInRunningPhases() {
		return false, nil, nil
	}

	if e.IntegrationInPhase(v1.IntegrationPhaseRunning, v1.IntegrationPhaseError) {
		condition := e.Integration.Status.GetCondition(v1.IntegrationConditionDeploymentAvailable)
		return condition != nil && condition.Status == corev1.ConditionTrue, nil, nil
	}

	// Don't deploy when a different strategy is needed (e.g. Knative, Cron)
	strategy, err := e.DetermineControllerStrategy()
	if err != nil {
		return false, NewIntegrationCondition(
			v1.IntegrationConditionDeploymentAvailable,
			corev1.ConditionFalse,
			v1.IntegrationConditionDeploymentAvailableReason,
			err.Error(),
		), err
	}

	if strategy != ControllerStrategyDeployment {
		return false, NewIntegrationCondition(
			v1.IntegrationConditionDeploymentAvailable,
			corev1.ConditionFalse,
			v1.IntegrationConditionDeploymentAvailableReason,
			"controller strategy: "+string(strategy),
		), nil
	}

	return e.IntegrationInPhase(v1.IntegrationPhaseDeploying), nil, nil
}

func (t *deploymentTrait) SelectControllerStrategy(e *Environment) (*ControllerStrategy, error) {
	deploymentStrategy := ControllerStrategyDeployment
	return &deploymentStrategy, nil
}

func (t *deploymentTrait) ControllerStrategySelectorOrder() int {
	return 10000
}

func (t *deploymentTrait) Apply(e *Environment) error {
	deployment := e.Resources.GetDeploymentForIntegration(e.Integration)
	// if deployment == nil {
	// 	deployment = t.loadDeploymentFor(e)
	// }
	if deployment == nil {
		deployment = t.getDeploymentFor(e)
	}

	// create a copy to avoid sharing the underlying annotation map
	annotations := make(map[string]string)
	if e.Integration.Annotations != nil {
		for k, v := range filterTransferableAnnotations(e.Integration.Annotations) {
			annotations[k] = v
		}
	}
	deployment.Annotations = annotations

	deadline := int32(60)
	if t.ProgressDeadlineSeconds != nil {
		deadline = *t.ProgressDeadlineSeconds
	}
	deployment.Spec.ProgressDeadlineSeconds = &deadline

	switch t.Strategy {
	case appsv1.RecreateDeploymentStrategyType:
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: t.Strategy,
		}
	case appsv1.RollingUpdateDeploymentStrategyType:
		deployment.Spec.Strategy = appsv1.DeploymentStrategy{
			Type: t.Strategy,
		}

		if t.RollingUpdateMaxSurge != nil || t.RollingUpdateMaxUnavailable != nil {
			var maxSurge *intstr.IntOrString
			var maxUnavailable *intstr.IntOrString

			if t.RollingUpdateMaxSurge != nil {
				v := intstr.FromInt(*t.RollingUpdateMaxSurge)
				maxSurge = &v
			}
			if t.RollingUpdateMaxUnavailable != nil {
				v := intstr.FromInt(*t.RollingUpdateMaxUnavailable)
				maxUnavailable = &v
			}

			deployment.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
				MaxSurge:       maxSurge,
				MaxUnavailable: maxUnavailable,
			}
		}
	}
	// Reconcile the deployment replicas
	replicas := e.Integration.Spec.Replicas
	// Deployment replicas defaults to 1, so we avoid forcing
	// an update to nil that will result to another update cycle
	// back to that default value by the Deployment controller.
	if replicas == nil {
		one := int32(1)
		replicas = &one
	}
	deployment.Spec.Replicas = replicas

	if e.Integration.Spec.ServiceAccountName != "" {
		deployment.Spec.Template.Spec = corev1.PodSpec{
			ServiceAccountName: e.Integration.Spec.ServiceAccountName,
		}
	}

	e.Integration.Status.SetCondition(
		v1.IntegrationConditionDeploymentAvailable,
		corev1.ConditionTrue,
		v1.IntegrationConditionDeploymentAvailableReason,
		fmt.Sprintf("deployment name is %s", deployment.Name),
	)

	e.Resources.Add(deployment)
	return nil
}

func toDeployment(u *unstructured.Unstructured) (*appsv1.Deployment, error) {
	d := &appsv1.Deployment{}
	data, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (t *deploymentTrait) loadDeploymentFor(e *Environment) *appsv1.Deployment {
	if t.Name != "" {
		unstr, err := kubernetes.GetUnstructured(
			e.Ctx, e.Client,
			schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			t.Name,
			e.Integration.Namespace,
		)
		if err != nil {
			t.L.ForIntegration(e.Integration).Errorf(err, "Integration %s/%s tried to load Deployment by name %s", e.Integration.Namespace, e.Integration.Name, t.Name)
		}
		deployment, err := toDeployment(unstr)
		if err != nil {
			t.L.ForIntegration(e.Integration).Errorf(err, "Integration %s/%s tried to unmarshal Deployment by name %s", e.Integration.Namespace, e.Integration.Name, t.Name)
		}
		deployment.SetManagedFields(nil)
		return deployment
	}

	return nil
}

func (t *deploymentTrait) getDeploymentFor(e *Environment) *appsv1.Deployment {
	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.Integration.Name,
			Namespace: e.Integration.Namespace,
			Labels: map[string]string{
				v1.IntegrationLabel: e.Integration.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: e.Integration.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					v1.IntegrationLabel: e.Integration.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						v1.IntegrationLabel: e.Integration.Name,
					},
				},
			},
		},
	}

	// TODO, fix according to the import feature, hardcoded for POC
	if t.Name != "" {
		deployment.Name = t.Name
		deployment.Labels["app"] = "my-camel-sb-svc"
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "my-camel-sb-svc",
			},
		}
		deployment.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "my-camel-sb-svc",
				},
			},
		}
	}

	return &deployment
}

func (t *deploymentTrait) Reverse(e *Environment, traits *v1.Traits) error {
	deploymentName := e.Integration.Annotations["camel.apache.org/imported-by"]
	dpls, err := e.Client.AppsV1().Deployments(e.Integration.Namespace).List(e.Ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentName),
	},
	)
	if err != nil {
		return err
	}
	if dpls != nil {
		dpl := dpls.Items[0]
		if traits.Deployment == nil {
			traits.Deployment = &traitv1.DeploymentTrait{}
		}
		traits.Deployment.Name = deploymentName

		e.Resources.Add(&dpl)
		// clearManagedFields := `[{"op": "replace", "path": "/metadata/managedFields", "value": [{}]}]`
		// patched, err := e.Client.AppsV1().Deployments(e.Integration.Namespace).Patch(
		// 	e.Ctx,
		// 	dpl.GetName(),
		// 	types.JSONPatchType,
		// 	[]byte(clearManagedFields),
		// 	metav1.PatchOptions{FieldManager: "camel-k-operator"},
		// )
		// if err != nil {
		// 	return err
		// }

		// e.Resources.Add(patched)
	}

	return nil
}
