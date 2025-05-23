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

// JSONSchemaPropsApplyConfiguration represents a declarative configuration of the JSONSchemaProps type for use
// with apply.
type JSONSchemaPropsApplyConfiguration struct {
	ID           *string                                     `json:"id,omitempty"`
	Description  *string                                     `json:"description,omitempty"`
	Title        *string                                     `json:"title,omitempty"`
	Properties   map[string]JSONSchemaPropApplyConfiguration `json:"properties,omitempty"`
	Required     []string                                    `json:"required,omitempty"`
	Example      *JSONApplyConfiguration                     `json:"example,omitempty"`
	ExternalDocs *ExternalDocumentationApplyConfiguration    `json:"externalDocs,omitempty"`
	Schema       *camelv1.JSONSchemaURL                      `json:"$schema,omitempty"`
	Type         *string                                     `json:"type,omitempty"`
}

// JSONSchemaPropsApplyConfiguration constructs a declarative configuration of the JSONSchemaProps type for use with
// apply.
func JSONSchemaProps() *JSONSchemaPropsApplyConfiguration {
	return &JSONSchemaPropsApplyConfiguration{}
}

// WithID sets the ID field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ID field is set to the value of the last call.
func (b *JSONSchemaPropsApplyConfiguration) WithID(value string) *JSONSchemaPropsApplyConfiguration {
	b.ID = &value
	return b
}

// WithDescription sets the Description field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Description field is set to the value of the last call.
func (b *JSONSchemaPropsApplyConfiguration) WithDescription(value string) *JSONSchemaPropsApplyConfiguration {
	b.Description = &value
	return b
}

// WithTitle sets the Title field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Title field is set to the value of the last call.
func (b *JSONSchemaPropsApplyConfiguration) WithTitle(value string) *JSONSchemaPropsApplyConfiguration {
	b.Title = &value
	return b
}

// WithProperties puts the entries into the Properties field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Properties field,
// overwriting an existing map entries in Properties field with the same key.
func (b *JSONSchemaPropsApplyConfiguration) WithProperties(entries map[string]JSONSchemaPropApplyConfiguration) *JSONSchemaPropsApplyConfiguration {
	if b.Properties == nil && len(entries) > 0 {
		b.Properties = make(map[string]JSONSchemaPropApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.Properties[k] = v
	}
	return b
}

// WithRequired adds the given value to the Required field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Required field.
func (b *JSONSchemaPropsApplyConfiguration) WithRequired(values ...string) *JSONSchemaPropsApplyConfiguration {
	for i := range values {
		b.Required = append(b.Required, values[i])
	}
	return b
}

// WithExample sets the Example field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Example field is set to the value of the last call.
func (b *JSONSchemaPropsApplyConfiguration) WithExample(value *JSONApplyConfiguration) *JSONSchemaPropsApplyConfiguration {
	b.Example = value
	return b
}

// WithExternalDocs sets the ExternalDocs field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ExternalDocs field is set to the value of the last call.
func (b *JSONSchemaPropsApplyConfiguration) WithExternalDocs(value *ExternalDocumentationApplyConfiguration) *JSONSchemaPropsApplyConfiguration {
	b.ExternalDocs = value
	return b
}

// WithSchema sets the Schema field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Schema field is set to the value of the last call.
func (b *JSONSchemaPropsApplyConfiguration) WithSchema(value camelv1.JSONSchemaURL) *JSONSchemaPropsApplyConfiguration {
	b.Schema = &value
	return b
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *JSONSchemaPropsApplyConfiguration) WithType(value string) *JSONSchemaPropsApplyConfiguration {
	b.Type = &value
	return b
}
