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

package camel

import (
	"fmt"
	"io"
	"strings"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/util/jitpack"
	"github.com/apache/camel-k/v2/pkg/util/maven"
	"github.com/rs/xid"
)

// NormalizeDependency converts different forms of camel dependencies
// -- `camel-xxx`, `camel-quarkus-xxx`, and `camel-quarkus:xxx` --
// into the unified form `camel:xxx`.
func NormalizeDependency(dependency string) string {
	newDep := dependency
	switch {
	case strings.HasPrefix(newDep, "camel-quarkus-"):
		newDep = "camel:" + strings.TrimPrefix(dependency, "camel-quarkus-")
	case strings.HasPrefix(newDep, "camel-quarkus:"):
		newDep = "camel:" + strings.TrimPrefix(dependency, "camel-quarkus:")
	case strings.HasPrefix(newDep, "camel-k-"):
		newDep = "camel-k:" + strings.TrimPrefix(dependency, "camel-k-")
	case strings.HasPrefix(newDep, "camel-k:"):
		// do nothing
	case strings.HasPrefix(newDep, "camel-"):
		newDep = "camel:" + strings.TrimPrefix(dependency, "camel-")
	}
	return newDep
}

// ValidateDependency validates a dependency against Camel catalog.
// It only shows warning and does not throw error in case the Catalog is just not complete,
// and we don't want to let it stop the process.
func ValidateDependency(catalog *RuntimeCatalog, dependency string, out io.Writer) {
	if err := ValidateDependencyE(catalog, dependency); err != nil {
		fmt.Fprintf(out, "Warning: %s\n", err.Error())
	}

	switch {
	case strings.HasPrefix(dependency, "mvn:org.apache.camel:"):
		component := strings.Split(dependency, ":")[2]
		fmt.Fprintf(out, "Warning: do not use %s. Use %s instead\n",
			dependency, NormalizeDependency(component))
	case strings.HasPrefix(dependency, "mvn:org.apache.camel.quarkus:"):
		component := strings.Split(dependency, ":")[2]
		fmt.Fprintf(out, "Warning: do not use %s. Use %s instead\n",
			dependency, NormalizeDependency(component))
	}
}

// ValidateDependencyE validates a dependency against Camel catalog and throws error
// in case it does not exist in the catalog.
func ValidateDependencyE(catalog *RuntimeCatalog, dependency string) error {
	var artifact string
	switch {
	case strings.HasPrefix(dependency, "camel:"):
		artifact = strings.TrimPrefix(dependency, "camel:")
	case strings.HasPrefix(dependency, "camel-quarkus:"):
		artifact = strings.TrimPrefix(dependency, "camel-quarkus:")
	case strings.HasPrefix(dependency, "camel-"):
		artifact = dependency
	}

	if artifact == "" {
		return nil
	}

	if ok := catalog.IsValidArtifact(artifact); !ok {
		return fmt.Errorf("dependency %s not found in Camel catalog", dependency)
	}

	return nil
}

// ValidateDependenciesE validates dependencies against Camel catalog and throws error
// in case it does not exist in the catalog.
func ValidateDependenciesE(catalog *RuntimeCatalog, dependencies []string) error {
	for _, dependency := range dependencies {
		if err := ValidateDependencyE(catalog, dependency); err != nil {
			return err
		}
	}

	return nil
}

// ManageIntegrationDependencies sets up all the required dependencies for the given Maven project.
func ManageIntegrationDependencies(project *maven.Project, dependencies []string, catalog *RuntimeCatalog) error {
	// Add dependencies from build
	if err := addDependencies(project, dependencies, catalog); err != nil {
		return err
	}
	// Add dependencies from catalog
	addDependenciesFromCatalog(project, catalog)
	// Post process dependencies
	postProcessDependencies(project, catalog)

	return nil
}

func addDependencies(project *maven.Project, dependencies []string, catalog *RuntimeCatalog) error {
	for _, d := range dependencies {
		switch {
		case strings.HasPrefix(d, "bom:"):
			if err := addBOM(project, d); err != nil {
				return err
			}
		case strings.HasPrefix(d, "camel:"):
			addCamelComponent(project, catalog, d)
		case strings.HasPrefix(d, "camel-k:"):
			addCamelKComponent(project, d)
		case strings.HasPrefix(d, "camel-quarkus:"):
			addCamelQuarkusComponent(project, d)
		case strings.HasPrefix(d, "mvn:"):
			addMavenDependency(project, d)
		default:
			if err := addJitPack(project, d); err != nil {
				return err
			}
		}
	}

	return nil
}

func addBOM(project *maven.Project, dependency string) error {
	gav := strings.TrimPrefix(dependency, "bom:")
	d, err := maven.ParseGAV(gav)
	if err != nil {
		return err
	}
	project.DependencyManagement.Dependencies = append(project.DependencyManagement.Dependencies,
		maven.Dependency{
			GroupID:    d.GroupID,
			ArtifactID: d.ArtifactID,
			Version:    d.Version,
			Type:       "pom",
			Scope:      "import",
		})

	return nil
}

func addCamelComponent(project *maven.Project, catalog *RuntimeCatalog, dependency string) {
	artifactID := strings.TrimPrefix(dependency, "camel:")
	if catalog != nil && catalog.Runtime.Provider.IsQuarkusBased() {
		if !strings.HasPrefix(artifactID, "camel-") {
			artifactID = "camel-quarkus-" + artifactID
		}
		project.AddDependencyGAV("org.apache.camel.quarkus", artifactID, "")
	} else {
		if !strings.HasPrefix(artifactID, "camel-") {
			artifactID = "camel-" + artifactID
		}
		project.AddDependencyGAV("org.apache.camel", artifactID, "")
	}
}

func addCamelKComponent(project *maven.Project, dependency string) {
	artifactID := strings.TrimPrefix(dependency, "camel-k:")
	if !strings.HasPrefix(artifactID, "camel-k-") {
		artifactID = "camel-k-" + artifactID
	}
	project.AddDependencyGAV("org.apache.camel.k", artifactID, "")
}

func addCamelQuarkusComponent(project *maven.Project, dependency string) {
	artifactID := strings.TrimPrefix(dependency, "camel-quarkus:")
	if !strings.HasPrefix(artifactID, "camel-quarkus-") {
		artifactID = "camel-quarkus-" + artifactID
	}
	project.AddDependencyGAV("org.apache.camel.quarkus", artifactID, "")
}

func addMavenDependency(project *maven.Project, dependency string) {
	gav := strings.TrimPrefix(dependency, "mvn:")
	project.AddEncodedDependencyGAV(gav)
}

func addJitPack(project *maven.Project, dependency string) error {
	dep := jitpack.ToDependency(dependency)
	if dep == nil {
		return fmt.Errorf("unknown dependency type: %s", dependency)
	}

	project.AddDependency(*dep)

	addRepo := true
	for _, repo := range project.Repositories {
		if repo.URL == jitpack.RepoURL {
			addRepo = false
			break
		}
	}
	if addRepo {
		project.Repositories = append(project.Repositories, v1.Repository{
			ID:  "jitpack.io-" + xid.New().String(),
			URL: jitpack.RepoURL,
			Releases: v1.RepositoryPolicy{
				Enabled:        true,
				ChecksumPolicy: "fail",
			},
			Snapshots: v1.RepositoryPolicy{
				Enabled:        true,
				ChecksumPolicy: "fail",
			},
		})
	}

	return nil
}

func addDependenciesFromCatalog(project *maven.Project, catalog *RuntimeCatalog) {
	deps := make([]maven.Dependency, len(project.Dependencies))
	copy(deps, project.Dependencies)

	for _, d := range deps {
		if a, ok := catalog.Artifacts[d.ArtifactID]; ok {
			for _, dep := range a.Dependencies {
				md := maven.Dependency{
					GroupID:    dep.GroupID,
					ArtifactID: dep.ArtifactID,
					Type:       dep.Type,
					Classifier: dep.Classifier,
				}

				project.AddDependency(md)

				for _, e := range dep.Exclusions {
					me := maven.Exclusion{
						GroupID:    e.GroupID,
						ArtifactID: e.ArtifactID,
					}

					project.AddDependencyExclusion(md, me)
				}
			}
		}
	}
}

func postProcessDependencies(project *maven.Project, catalog *RuntimeCatalog) {
	deps := make([]maven.Dependency, len(project.Dependencies))
	copy(deps, project.Dependencies)

	for _, d := range deps {
		if a, ok := catalog.Artifacts[d.ArtifactID]; ok {
			md := maven.Dependency{
				GroupID:    a.GroupID,
				ArtifactID: a.ArtifactID,
			}

			for _, e := range a.Exclusions {
				me := maven.Exclusion{
					GroupID:    e.GroupID,
					ArtifactID: e.ArtifactID,
				}

				project.AddDependencyExclusion(md, me)
			}
		}
	}
}

// SanitizeIntegrationDependencies --.
func SanitizeIntegrationDependencies(dependencies []maven.Dependency) error {
	for i := range dependencies {
		dep := dependencies[i]

		// It may be externalized into runtime provider specific steps
		switch dep.GroupID {
		case "org.apache.camel":
			fallthrough
		case "org.apache.camel.k":
			fallthrough
		case "org.apache.camel.quarkus":
			//
			// Remove the version so we force using the one configured by the bom
			//
			dependencies[i].Version = ""
		}
	}

	return nil
}
