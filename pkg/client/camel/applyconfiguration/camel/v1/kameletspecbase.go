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
)

// KameletSpecBaseApplyConfiguration represents a declarative configuration of the KameletSpecBase type for use
// with apply.
type KameletSpecBaseApplyConfiguration struct {
	Definition   *JSONSchemaPropsApplyConfiguration                   `json:"definition,omitempty"`
	Sources      []SourceSpecApplyConfiguration                       `json:"sources,omitempty"`
	Template     *TemplateApplyConfiguration                          `json:"template,omitempty"`
	Types        map[camelv1.TypeSlot]EventTypeSpecApplyConfiguration `json:"types,omitempty"`
	DataTypes    map[camelv1.TypeSlot]DataTypesSpecApplyConfiguration `json:"dataTypes,omitempty"`
	Dependencies []string                                             `json:"dependencies,omitempty"`
}

// KameletSpecBaseApplyConfiguration constructs a declarative configuration of the KameletSpecBase type for use with
// apply.
func KameletSpecBase() *KameletSpecBaseApplyConfiguration {
	return &KameletSpecBaseApplyConfiguration{}
}

// WithDefinition sets the Definition field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Definition field is set to the value of the last call.
func (b *KameletSpecBaseApplyConfiguration) WithDefinition(value *JSONSchemaPropsApplyConfiguration) *KameletSpecBaseApplyConfiguration {
	b.Definition = value
	return b
}

// WithSources adds the given value to the Sources field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Sources field.
func (b *KameletSpecBaseApplyConfiguration) WithSources(values ...*SourceSpecApplyConfiguration) *KameletSpecBaseApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithSources")
		}
		b.Sources = append(b.Sources, *values[i])
	}
	return b
}

// WithTemplate sets the Template field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Template field is set to the value of the last call.
func (b *KameletSpecBaseApplyConfiguration) WithTemplate(value *TemplateApplyConfiguration) *KameletSpecBaseApplyConfiguration {
	b.Template = value
	return b
}

// WithTypes puts the entries into the Types field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Types field,
// overwriting an existing map entries in Types field with the same key.
func (b *KameletSpecBaseApplyConfiguration) WithTypes(entries map[camelv1.TypeSlot]EventTypeSpecApplyConfiguration) *KameletSpecBaseApplyConfiguration {
	if b.Types == nil && len(entries) > 0 {
		b.Types = make(map[camelv1.TypeSlot]EventTypeSpecApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.Types[k] = v
	}
	return b
}

// WithDataTypes puts the entries into the DataTypes field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the DataTypes field,
// overwriting an existing map entries in DataTypes field with the same key.
func (b *KameletSpecBaseApplyConfiguration) WithDataTypes(entries map[camelv1.TypeSlot]DataTypesSpecApplyConfiguration) *KameletSpecBaseApplyConfiguration {
	if b.DataTypes == nil && len(entries) > 0 {
		b.DataTypes = make(map[camelv1.TypeSlot]DataTypesSpecApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.DataTypes[k] = v
	}
	return b
}

// WithDependencies adds the given value to the Dependencies field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Dependencies field.
func (b *KameletSpecBaseApplyConfiguration) WithDependencies(values ...string) *KameletSpecBaseApplyConfiguration {
	for i := range values {
		b.Dependencies = append(b.Dependencies, values[i])
	}
	return b
}
