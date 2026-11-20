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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// IntegrationBuildKind --.
	IntegrationBuildKind string = "IntegrationBuild"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=integrationbuilds,scope=Namespaced,shortName=ib,categories=kamel;camel
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// IntegrationBuild is the Schema for the IntegrationBuild API. This resource takes care of building your Camel application and transforming into a
// container pushed to a registry. The container image can be used to run the application on the cluster.
type IntegrationBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationBuildSpec   `json:"spec,omitempty"`
	Status IntegrationBuildStatus `json:"status,omitempty"`
}

// IntegrationBuildSpec defines the desired state of an IntegrationBuild.
type IntegrationBuildSpec struct {
	// The user ID the build Pod runs as (Pod securityContext.runAsUser).
	SecurityContextRunAsUser *int32 `json:"securityContextRunAsUser,omitempty"`
	// The names of the ConfigMaps containing the Camel sources, mounted under /input in the export task.
	Sources []string `json:"sources,omitempty"`
	// The task exporting the sources into a Maven project (Camel JBang).
	ExportTask *CamelJBangExporterTaskConf `json:"exportTask,omitempty"`
	// The task building the Maven project and pushing the container image (Maven JIB).
	BuildTask MavenJIBBuilderTaskConf `json:"buildTask,omitempty"`
}

// IntegrationBuildStatus defines the observed state of an IntegrationBuild.
type IntegrationBuildStatus struct {
	// The version (tag) of the container image produced.
	ImageVersion *string `json:"imageVersion,omitempty"`
	// The digest (sha) of the container image produced.
	ImageSha *string `json:"imageSha,omitempty"`
	// The current phase of the build.
	Phase IntegrationBuildPhase `json:"phase,omitempty"`
}

// IntegrationBuildPhase is the phase of an IntegrationBuild.
type IntegrationBuildPhase string

const (
	// IntegrationBuildPhaseNone means the build has not been initialized yet.
	IntegrationBuildPhaseNone IntegrationBuildPhase = ""
	// IntegrationBuildPhaseBuilding means the build Job is running.
	IntegrationBuildPhaseBuilding IntegrationBuildPhase = "Building"
	// IntegrationBuildPhaseSucceeded means the image has been built and pushed.
	IntegrationBuildPhaseSucceeded IntegrationBuildPhase = "Succeeded"
	// IntegrationBuildPhaseFailed means the build has failed.
	IntegrationBuildPhaseFailed IntegrationBuildPhase = "Failed"
	// IntegrationBuildPhaseError means the build is in Error phase (it has not either started).
	IntegrationBuildPhaseError IntegrationBuildPhase = "Error"
)

// CamelJBangExporterTaskConf configures the Camel JBang export task.
type CamelJBangExporterTaskConf struct {
	// The container image running the task (defaults to the Camel JBang image).
	ImageTask *string `json:"imageTask,omitempty"`
	// The Camel runtime to export to (i.e. quarkus).
	RuntimeProvider *string `json:"runtimeProvider,omitempty"`
	// The Camel runtime version.
	RuntimeVersion *string `json:"runtimeVersion,omitempty"`
	// The Maven GAV (groupId:artifactId:version) of the exported project.
	GAV *string `json:"gav,omitempty"`
}

// MavenJIBBuilderTaskConf configures the Maven JIB build task.
type MavenJIBBuilderTaskConf struct {
	// The container image running the task (defaults to a Maven image).
	ImageTask *string `json:"imageTask,omitempty"`
	// The container image repository the application image is pushed to.
	ImageDestination string `json:"imageDestination,omitempty"`
	// The tag of the application image.
	ImageDestinationVersion string `json:"imageDestinationVersion,omitempty"`
	// The base image the application image is built from.
	ImageBaseSource string `json:"imageBaseSource,omitempty"`
	// The user ID the application container runs as.
	ImageContainerUser *int32 `json:"imageContainerUser,omitempty"`
	// Whether pushing to insecure (plain HTTP or untrusted TLS) registries is allowed.
	AllowInsecureRegistries *bool `json:"allowInsecureRegistries,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationBuildList contains a list of IntegrationBuilds.
type IntegrationBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []IntegrationBuild `json:"items"`
}
