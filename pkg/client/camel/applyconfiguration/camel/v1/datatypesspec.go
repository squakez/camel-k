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

// DataTypesSpecApplyConfiguration represents a declarative configuration of the DataTypesSpec type for use
// with apply.
type DataTypesSpecApplyConfiguration struct {
	Default *string                                   `json:"default,omitempty"`
	Types   map[string]DataTypeSpecApplyConfiguration `json:"types,omitempty"`
	Headers map[string]HeaderSpecApplyConfiguration   `json:"headers,omitempty"`
}

// DataTypesSpecApplyConfiguration constructs a declarative configuration of the DataTypesSpec type for use with
// apply.
func DataTypesSpec() *DataTypesSpecApplyConfiguration {
	return &DataTypesSpecApplyConfiguration{}
}

// WithDefault sets the Default field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Default field is set to the value of the last call.
func (b *DataTypesSpecApplyConfiguration) WithDefault(value string) *DataTypesSpecApplyConfiguration {
	b.Default = &value
	return b
}

// WithTypes puts the entries into the Types field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Types field,
// overwriting an existing map entries in Types field with the same key.
func (b *DataTypesSpecApplyConfiguration) WithTypes(entries map[string]DataTypeSpecApplyConfiguration) *DataTypesSpecApplyConfiguration {
	if b.Types == nil && len(entries) > 0 {
		b.Types = make(map[string]DataTypeSpecApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.Types[k] = v
	}
	return b
}

// WithHeaders puts the entries into the Headers field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Headers field,
// overwriting an existing map entries in Headers field with the same key.
func (b *DataTypesSpecApplyConfiguration) WithHeaders(entries map[string]HeaderSpecApplyConfiguration) *DataTypesSpecApplyConfiguration {
	if b.Headers == nil && len(entries) > 0 {
		b.Headers = make(map[string]HeaderSpecApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.Headers[k] = v
	}
	return b
}
