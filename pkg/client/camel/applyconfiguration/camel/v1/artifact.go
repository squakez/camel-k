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

// ArtifactApplyConfiguration represents an declarative configuration of the Artifact type for use
// with apply.
type ArtifactApplyConfiguration struct {
	ID         *string `json:"id,omitempty"`
	Location   *string `json:"location,omitempty"`
	Target     *string `json:"target,omitempty"`
	Checksum   *string `json:"checksum,omitempty"`
	Executable *bool   `json:"executable,omitempty"`
}

// ArtifactApplyConfiguration constructs an declarative configuration of the Artifact type for use with
// apply.
func Artifact() *ArtifactApplyConfiguration {
	return &ArtifactApplyConfiguration{}
}

// WithID sets the ID field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ID field is set to the value of the last call.
func (b *ArtifactApplyConfiguration) WithID(value string) *ArtifactApplyConfiguration {
	b.ID = &value
	return b
}

// WithLocation sets the Location field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Location field is set to the value of the last call.
func (b *ArtifactApplyConfiguration) WithLocation(value string) *ArtifactApplyConfiguration {
	b.Location = &value
	return b
}

// WithTarget sets the Target field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Target field is set to the value of the last call.
func (b *ArtifactApplyConfiguration) WithTarget(value string) *ArtifactApplyConfiguration {
	b.Target = &value
	return b
}

// WithChecksum sets the Checksum field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Checksum field is set to the value of the last call.
func (b *ArtifactApplyConfiguration) WithChecksum(value string) *ArtifactApplyConfiguration {
	b.Checksum = &value
	return b
}

// WithExecutable sets the Executable field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Executable field is set to the value of the last call.
func (b *ArtifactApplyConfiguration) WithExecutable(value bool) *ArtifactApplyConfiguration {
	b.Executable = &value
	return b
}
