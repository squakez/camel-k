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

// SourceSpecApplyConfiguration represents a declarative configuration of the SourceSpec type for use
// with apply.
type SourceSpecApplyConfiguration struct {
	DataSpecApplyConfiguration `json:",inline"`
	Language                   *camelv1.Language   `json:"language,omitempty"`
	Loader                     *string             `json:"loader,omitempty"`
	Interceptors               []string            `json:"interceptors,omitempty"`
	Type                       *camelv1.SourceType `json:"type,omitempty"`
	PropertyNames              []string            `json:"property-names,omitempty"`
	FromKamelet                *bool               `json:"from-kamelet,omitempty"`
}

// SourceSpecApplyConfiguration constructs a declarative configuration of the SourceSpec type for use with
// apply.
func SourceSpec() *SourceSpecApplyConfiguration {
	return &SourceSpecApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithName(value string) *SourceSpecApplyConfiguration {
	b.DataSpecApplyConfiguration.Name = &value
	return b
}

// WithPath sets the Path field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Path field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithPath(value string) *SourceSpecApplyConfiguration {
	b.DataSpecApplyConfiguration.Path = &value
	return b
}

// WithContent sets the Content field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Content field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithContent(value string) *SourceSpecApplyConfiguration {
	b.DataSpecApplyConfiguration.Content = &value
	return b
}

// WithRawContent adds the given value to the RawContent field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the RawContent field.
func (b *SourceSpecApplyConfiguration) WithRawContent(values ...byte) *SourceSpecApplyConfiguration {
	for i := range values {
		b.DataSpecApplyConfiguration.RawContent = append(b.DataSpecApplyConfiguration.RawContent, values[i])
	}
	return b
}

// WithContentRef sets the ContentRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ContentRef field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithContentRef(value string) *SourceSpecApplyConfiguration {
	b.DataSpecApplyConfiguration.ContentRef = &value
	return b
}

// WithContentKey sets the ContentKey field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ContentKey field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithContentKey(value string) *SourceSpecApplyConfiguration {
	b.DataSpecApplyConfiguration.ContentKey = &value
	return b
}

// WithContentType sets the ContentType field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ContentType field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithContentType(value string) *SourceSpecApplyConfiguration {
	b.DataSpecApplyConfiguration.ContentType = &value
	return b
}

// WithCompression sets the Compression field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Compression field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithCompression(value bool) *SourceSpecApplyConfiguration {
	b.DataSpecApplyConfiguration.Compression = &value
	return b
}

// WithLanguage sets the Language field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Language field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithLanguage(value camelv1.Language) *SourceSpecApplyConfiguration {
	b.Language = &value
	return b
}

// WithLoader sets the Loader field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Loader field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithLoader(value string) *SourceSpecApplyConfiguration {
	b.Loader = &value
	return b
}

// WithInterceptors adds the given value to the Interceptors field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Interceptors field.
func (b *SourceSpecApplyConfiguration) WithInterceptors(values ...string) *SourceSpecApplyConfiguration {
	for i := range values {
		b.Interceptors = append(b.Interceptors, values[i])
	}
	return b
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithType(value camelv1.SourceType) *SourceSpecApplyConfiguration {
	b.Type = &value
	return b
}

// WithPropertyNames adds the given value to the PropertyNames field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the PropertyNames field.
func (b *SourceSpecApplyConfiguration) WithPropertyNames(values ...string) *SourceSpecApplyConfiguration {
	for i := range values {
		b.PropertyNames = append(b.PropertyNames, values[i])
	}
	return b
}

// WithFromKamelet sets the FromKamelet field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FromKamelet field is set to the value of the last call.
func (b *SourceSpecApplyConfiguration) WithFromKamelet(value bool) *SourceSpecApplyConfiguration {
	b.FromKamelet = &value
	return b
}
