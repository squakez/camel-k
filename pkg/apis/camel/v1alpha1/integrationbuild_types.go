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
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="The build phase"
// +kubebuilder:printcolumn:name="Container SHA",type=string,JSONPath=`.status.imageSha`,description="The container build SHA"

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
	// The task exporting the sources into a Maven project (Camel JBang).
	ExportTask *CamelJBangExporterTaskConf `json:"exportTask,omitempty"`
	// The task cloning a Git project.
	GitCloneTask *GitCloneTaskConf `json:"cloneTask,omitempty"`
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
	// The names of the ConfigMaps containing the Camel sources.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:items:MinLength=1
	Sources []string `json:"sources"`
	// The container image running the task (defaults to the Camel JBang image).
	ImageTask *string `json:"imageTask,omitempty"`
	// The Camel runtime to export to (i.e. quarkus).
	RuntimeProvider *string `json:"runtimeProvider,omitempty"`
	// The Camel runtime version.
	RuntimeVersion *string `json:"runtimeVersion,omitempty"`
	// The Maven GAV (groupId:artifactId:version) of the exported project.
	GAV *string `json:"gav,omitempty"`
}

// GitCloneTaskConf configures the git clone task.
type GitCloneTaskConf struct {
	// the configuration of the project to build on Git
	Git GitConfigSpec `json:"git"`
}

// MavenJIBBuilderTaskConf configures the Maven JIB build task.
type MavenJIBBuilderTaskConf struct {
	// The container image running the task (defaults to a Maven image).
	ImageTask *string `json:"imageTask,omitempty"`
	// The container image repository the application image is pushed to.
	// +kubebuilder:validation:MinLength=1
	ImageDestination string `json:"imageDestination"`
	// The tag of the application image.
	// +kubebuilder:validation:MinLength=1
	ImageDestinationVersion string `json:"imageDestinationVersion"`
	// The base image the application image is built from.
	// +kubebuilder:validation:MinLength=1
	ImageBaseSource string `json:"imageBaseSource"`
	// The user ID the application container runs as.
	ImageContainerUser *int32 `json:"imageContainerUser,omitempty"`
	// Whether pushing to insecure (plain HTTP or untrusted TLS) registries is allowed.
	AllowInsecureRegistries *bool `json:"allowInsecureRegistries,omitempty"`
}

// GitConfigSpec defines the Git configuration of a project.
// NOTE: this must be copied from v1 API. When this package is promoted, then, we can merge both into a unique one.
type GitConfigSpec struct {
	// the URL of the project
	// +kubebuilder:validation:MinLength=1
	URL string `json:"url,omitempty"`
	// the Kubernetes secret where token is stored
	Secret string `json:"secret,omitempty"`
	// the username used to authenticate against the Git server, together with the token stored in `secret` (default `camel-k`).
	// Some Git servers validate it, for example Bitbucket Cloud access tokens require `x-token-auth`.
	Username string `json:"username,omitempty"`
	// the git branch to check out
	Branch string `json:"branch,omitempty"`
	// the git tag to check out
	Tag string `json:"tag,omitempty"`
	// the git commit (full SHA) to check out
	Commit string `json:"commit,omitempty"`
	// the path you want to use for your project. If provided, it must be an existing directory on the Git repository.
	Path string `json:"path,omitempty"`
}

// +kubebuilder:object:root=true

// IntegrationBuildList contains a list of IntegrationBuilds.
type IntegrationBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []IntegrationBuild `json:"items"`
}
