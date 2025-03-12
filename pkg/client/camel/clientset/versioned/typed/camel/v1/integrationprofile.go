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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	context "context"

	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	applyconfigurationcamelv1 "github.com/apache/camel-k/v2/pkg/client/camel/applyconfiguration/camel/v1"
	scheme "github.com/apache/camel-k/v2/pkg/client/camel/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// IntegrationProfilesGetter has a method to return a IntegrationProfileInterface.
// A group's client should implement this interface.
type IntegrationProfilesGetter interface {
	IntegrationProfiles(namespace string) IntegrationProfileInterface
}

// IntegrationProfileInterface has methods to work with IntegrationProfile resources.
type IntegrationProfileInterface interface {
	Create(ctx context.Context, integrationProfile *camelv1.IntegrationProfile, opts metav1.CreateOptions) (*camelv1.IntegrationProfile, error)
	Update(ctx context.Context, integrationProfile *camelv1.IntegrationProfile, opts metav1.UpdateOptions) (*camelv1.IntegrationProfile, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, integrationProfile *camelv1.IntegrationProfile, opts metav1.UpdateOptions) (*camelv1.IntegrationProfile, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*camelv1.IntegrationProfile, error)
	List(ctx context.Context, opts metav1.ListOptions) (*camelv1.IntegrationProfileList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *camelv1.IntegrationProfile, err error)
	Apply(ctx context.Context, integrationProfile *applyconfigurationcamelv1.IntegrationProfileApplyConfiguration, opts metav1.ApplyOptions) (result *camelv1.IntegrationProfile, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, integrationProfile *applyconfigurationcamelv1.IntegrationProfileApplyConfiguration, opts metav1.ApplyOptions) (result *camelv1.IntegrationProfile, err error)
	IntegrationProfileExpansion
}

// integrationProfiles implements IntegrationProfileInterface
type integrationProfiles struct {
	*gentype.ClientWithListAndApply[*camelv1.IntegrationProfile, *camelv1.IntegrationProfileList, *applyconfigurationcamelv1.IntegrationProfileApplyConfiguration]
}

// newIntegrationProfiles returns a IntegrationProfiles
func newIntegrationProfiles(c *CamelV1Client, namespace string) *integrationProfiles {
	return &integrationProfiles{
		gentype.NewClientWithListAndApply[*camelv1.IntegrationProfile, *camelv1.IntegrationProfileList, *applyconfigurationcamelv1.IntegrationProfileApplyConfiguration](
			"integrationprofiles",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *camelv1.IntegrationProfile { return &camelv1.IntegrationProfile{} },
			func() *camelv1.IntegrationProfileList { return &camelv1.IntegrationProfileList{} },
		),
	}
}
