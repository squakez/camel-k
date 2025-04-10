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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

import (
	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IntegrationProfileBuildSpecApplyConfiguration represents a declarative configuration of the IntegrationProfileBuildSpec type for use
// with apply.
type IntegrationProfileBuildSpecApplyConfiguration struct {
	RuntimeVersion  *string                         `json:"runtimeVersion,omitempty"`
	RuntimeProvider *camelv1.RuntimeProvider        `json:"runtimeProvider,omitempty"`
	BaseImage       *string                         `json:"baseImage,omitempty"`
	Registry        *RegistrySpecApplyConfiguration `json:"registry,omitempty"`
	Timeout         *metav1.Duration                `json:"timeout,omitempty"`
	Maven           *MavenSpecApplyConfiguration    `json:"maven,omitempty"`
}

// IntegrationProfileBuildSpecApplyConfiguration constructs a declarative configuration of the IntegrationProfileBuildSpec type for use with
// apply.
func IntegrationProfileBuildSpec() *IntegrationProfileBuildSpecApplyConfiguration {
	return &IntegrationProfileBuildSpecApplyConfiguration{}
}

// WithRuntimeVersion sets the RuntimeVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RuntimeVersion field is set to the value of the last call.
func (b *IntegrationProfileBuildSpecApplyConfiguration) WithRuntimeVersion(value string) *IntegrationProfileBuildSpecApplyConfiguration {
	b.RuntimeVersion = &value
	return b
}

// WithRuntimeProvider sets the RuntimeProvider field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RuntimeProvider field is set to the value of the last call.
func (b *IntegrationProfileBuildSpecApplyConfiguration) WithRuntimeProvider(value camelv1.RuntimeProvider) *IntegrationProfileBuildSpecApplyConfiguration {
	b.RuntimeProvider = &value
	return b
}

// WithBaseImage sets the BaseImage field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BaseImage field is set to the value of the last call.
func (b *IntegrationProfileBuildSpecApplyConfiguration) WithBaseImage(value string) *IntegrationProfileBuildSpecApplyConfiguration {
	b.BaseImage = &value
	return b
}

// WithRegistry sets the Registry field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Registry field is set to the value of the last call.
func (b *IntegrationProfileBuildSpecApplyConfiguration) WithRegistry(value *RegistrySpecApplyConfiguration) *IntegrationProfileBuildSpecApplyConfiguration {
	b.Registry = value
	return b
}

// WithTimeout sets the Timeout field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Timeout field is set to the value of the last call.
func (b *IntegrationProfileBuildSpecApplyConfiguration) WithTimeout(value metav1.Duration) *IntegrationProfileBuildSpecApplyConfiguration {
	b.Timeout = &value
	return b
}

// WithMaven sets the Maven field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Maven field is set to the value of the last call.
func (b *IntegrationProfileBuildSpecApplyConfiguration) WithMaven(value *MavenSpecApplyConfiguration) *IntegrationProfileBuildSpecApplyConfiguration {
	b.Maven = value
	return b
}
