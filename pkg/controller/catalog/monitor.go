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

package catalog

import (
	"context"
	"strings"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// NewMonitorAction returns an action that monitors the catalog after it's fully initialized.
func NewMonitorAction() Action {
	return &monitorAction{}
}

type monitorAction struct {
	baseAction
}

func (action *monitorAction) Name() string {
	return "monitor"
}

func (action *monitorAction) CanHandle(catalog *v1.CamelCatalog) bool {
	return catalog.Status.Phase == v1.CamelCatalogPhaseReady || catalog.Status.Phase == v1.CamelCatalogPhaseError
}

func (action *monitorAction) Handle(ctx context.Context, catalog *v1.CamelCatalog) (*v1.CamelCatalog, error) {
	runtimeSpec := v1.RuntimeSpec{
		Version:  catalog.Spec.GetRuntimeVersion(),
		Provider: v1.RuntimeProviderPlainQuarkus,
	}
	cat, err := loadCatalog(ctx, action.client, catalog.Namespace, runtimeSpec)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		// Clone the catalog to enable Quarkus Plain runtime
		clonedCatalog := catalog.DeepCopy()
		clonedCatalog.Status = v1.CamelCatalogStatus{}
		clonedCatalog.ObjectMeta = metav1.ObjectMeta{
			Namespace:   catalog.Namespace,
			Name:        strings.ReplaceAll(catalog.Name, "camel-catalog", "camel-catalog-quarkus"),
			Labels:      catalog.Labels,
			Annotations: catalog.Annotations,
		}
		clonedCatalog.Spec.Runtime.Provider = v1.RuntimeProviderPlainQuarkus
		clonedCatalog.Spec.Runtime.Dependencies = []v1.MavenArtifact{
			v1.MavenArtifact{
				GroupID:    "org.apache.camel.quarkus",
				ArtifactID: "camel-quarkus-core",
			},
		}

		if err = action.client.Create(ctx, clonedCatalog); err != nil {
			return nil, err
		}
	}

	return catalog, nil
}

func loadCatalog(ctx context.Context, c client.Client, namespace string, runtimeSpec v1.RuntimeSpec) (*v1.CamelCatalog, error) {
	options := []k8sclient.ListOption{
		k8sclient.InNamespace(namespace),
	}
	list := v1.NewCamelCatalogList()
	if err := c.List(ctx, &list, options...); err != nil {
		return nil, err
	}
	for _, cc := range list.Items {
		if cc.Spec.Runtime.Provider == runtimeSpec.Provider && cc.Spec.Runtime.Version == runtimeSpec.Version {
			return &cc, nil
		}
	}

	return nil, nil
}
