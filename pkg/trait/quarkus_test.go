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

package trait

import (
	"testing"

	traitv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1/trait"

	"github.com/stretchr/testify/assert"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/builder"
	"github.com/apache/camel-k/v2/pkg/util/camel"
	"github.com/apache/camel-k/v2/pkg/util/test"
)

func TestConfigureQuarkusTraitBuildSubmitted(t *testing.T) {
	quarkusTrait, environment := createNominalQuarkusTest()
	environment.IntegrationKit.Status.Phase = v1.IntegrationKitPhaseBuildSubmitted

	configured, condition, err := quarkusTrait.Configure(environment)

	assert.True(t, configured)
	assert.Nil(t, err)
	assert.Nil(t, condition)

	err = quarkusTrait.Apply(environment)
	assert.Nil(t, err)

	build := getBuilderTask(environment.Pipeline)
	assert.NotNil(t, t, build)
	assert.Len(t, build.Steps, len(builder.Quarkus.CommonSteps))

	packageTask := getPackageTask(environment.Pipeline)
	assert.NotNil(t, t, packageTask)
	assert.Len(t, packageTask.Steps, 4)
}

func TestApplyQuarkusTraitDefaultKitLayout(t *testing.T) {
	quarkusTrait, environment := createNominalQuarkusTest()
	environment.Integration.Status.Phase = v1.IntegrationPhaseBuildingKit

	configured, condition, err := quarkusTrait.Configure(environment)
	assert.True(t, configured)
	assert.Nil(t, err)
	assert.Nil(t, condition)

	err = quarkusTrait.Apply(environment)
	assert.Nil(t, err)
	assert.Len(t, environment.IntegrationKits, 1)
	assert.Equal(t, environment.IntegrationKits[0].Labels[v1.IntegrationKitLayoutLabel], v1.IntegrationKitLayoutFastJar)
}

func TestApplyQuarkusTraitAnnotationKitConfiguration(t *testing.T) {
	quarkusTrait, environment := createNominalQuarkusTest()
	environment.Integration.Status.Phase = v1.IntegrationPhaseBuildingKit

	v1.SetAnnotation(&environment.Integration.ObjectMeta, v1.TraitAnnotationPrefix+"quarkus.foo", "camel-k")

	configured, condition, err := quarkusTrait.Configure(environment)
	assert.True(t, configured)
	assert.Nil(t, err)
	assert.Nil(t, condition)

	err = quarkusTrait.Apply(environment)
	assert.Nil(t, err)
	assert.Len(t, environment.IntegrationKits, 1)
	assert.Equal(t, v1.IntegrationKitLayoutFastJar, environment.IntegrationKits[0].Labels[v1.IntegrationKitLayoutLabel])
	assert.Equal(t, "camel-k", environment.IntegrationKits[0].Annotations[v1.TraitAnnotationPrefix+"quarkus.foo"])

}

func TestQuarkusTraitBuildModeOrder(t *testing.T) {
	quarkusTrait, environment := createNominalQuarkusTest()
	quarkusTrait.Modes = []traitv1.QuarkusMode{traitv1.NativeQuarkusMode, traitv1.JvmQuarkusMode}
	environment.Integration.Status.Phase = v1.IntegrationPhaseBuildingKit
	environment.Integration.Spec.Sources = []v1.SourceSpec{
		{
			Language: v1.LanguageYaml,
		},
	}

	err := quarkusTrait.Apply(environment)
	assert.Nil(t, err)
	assert.Len(t, environment.IntegrationKits, 2)
	// assure jvm mode is executed before native mode
	assert.Equal(t, environment.IntegrationKits[0].Labels[v1.IntegrationKitLayoutLabel], v1.IntegrationKitLayoutFastJar)
	assert.Equal(t, environment.IntegrationKits[1].Labels[v1.IntegrationKitLayoutLabel], v1.IntegrationKitLayoutNativeSources)
}

func createNominalQuarkusTest() (*quarkusTrait, *Environment) {
	trait, _ := newQuarkusTrait().(*quarkusTrait)
	client, _ := test.NewFakeClient()

	environment := &Environment{
		Catalog:      NewCatalog(client),
		CamelCatalog: &camel.RuntimeCatalog{},
		Integration: &v1.Integration{
			Spec: v1.IntegrationSpec{
				Sources: []v1.SourceSpec{
					{
						Language: v1.LanguageJavaSource,
					},
				},
			},
		},
		IntegrationKit: &v1.IntegrationKit{},
		Pipeline: []v1.Task{
			{
				Builder: &v1.BuilderTask{},
			},
			{
				Package: &v1.BuilderTask{},
			},
		},
		Platform: &v1.IntegrationPlatform{},
	}

	return trait, environment
}

func TestGetLanguageSettingsWithoutLoaders(t *testing.T) {
	environment := &Environment{
		CamelCatalog: &camel.RuntimeCatalog{
			CamelCatalogSpec: v1.CamelCatalogSpec{
				Loaders: map[string]v1.CamelLoader{},
			},
		},
	}
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageJavaSource))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageGroovy))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageJavaScript))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageKotlin))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageJavaShell))
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageKamelet))
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageXML))
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageYaml))
}

func TestGetLanguageSettingsWithoutMetadata(t *testing.T) {
	environment := &Environment{
		CamelCatalog: &camel.RuntimeCatalog{
			CamelCatalogSpec: v1.CamelCatalogSpec{
				Loaders: map[string]v1.CamelLoader{
					"java":    {},
					"groovy":  {},
					"js":      {},
					"kts":     {},
					"jsh":     {},
					"kamelet": {},
					"xml":     {},
					"yaml":    {},
				},
			},
		},
	}
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageJavaSource))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageGroovy))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageJavaScript))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageKotlin))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageJavaShell))
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageKamelet))
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageXML))
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageYaml))
}

func TestGetLanguageSettingsWithLoaders(t *testing.T) {
	environment := &Environment{
		CamelCatalog: &camel.RuntimeCatalog{
			CamelCatalogSpec: v1.CamelCatalogSpec{
				Loaders: map[string]v1.CamelLoader{
					"java": {
						Metadata: map[string]string{
							"native":                         "true",
							"sources-required-at-build-time": "true",
						},
					},
					"groovy": {
						Metadata: map[string]string{
							"native":                         "false",
							"sources-required-at-build-time": "false",
						},
					},
					"js": {
						Metadata: map[string]string{
							"native":                         "true",
							"sources-required-at-build-time": "false",
						},
					},
					"kts": {
						Metadata: map[string]string{
							"native":                         "false",
							"sources-required-at-build-time": "true",
						},
					},
					"jsh": {
						Metadata: map[string]string{
							"native": "true",
						},
					},
				},
			},
		},
	}
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: true}, getLanguageSettings(environment, v1.LanguageJavaSource))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageGroovy))
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageJavaScript))
	assert.Equal(t, languageSettings{native: false, sourcesRequiredAtBuildTime: true}, getLanguageSettings(environment, v1.LanguageKotlin))
	assert.Equal(t, languageSettings{native: true, sourcesRequiredAtBuildTime: false}, getLanguageSettings(environment, v1.LanguageJavaShell))
}

func TestPropagateStatusTraits(t *testing.T) {
	quarkusTrait, environment := createNominalQuarkusTest()
	environment.IntegrationKit = nil
	environment.Integration = &v1.Integration{
		Spec:   v1.IntegrationSpec{},
		Status: v1.IntegrationStatus{},
	}

	environment.ExecutedTraits = []Trait{
		&camelTrait{
			BasePlatformTrait: NewBasePlatformTrait("camel", 200),
			CamelTrait: traitv1.CamelTrait{
				Properties:     []string{"hello=world"},
				RuntimeVersion: "1.2.3",
			},
		},
	}

	newKit, err := quarkusTrait.newIntegrationKit(environment, fastJarPackageType)
	assert.Nil(t, err)
	assert.Equal(t, []string{"hello=world"}, newKit.Spec.Traits.Camel.Properties)
	assert.Equal(t, "1.2.3", newKit.Spec.Traits.Camel.RuntimeVersion)
	// The Quarkus trait was declared empty
	assert.Equal(t, &traitv1.QuarkusTrait{}, newKit.Spec.Traits.Quarkus)
}

func TestPropagateQuarkusTrait(t *testing.T) {
	quarkusTrait, environment := createNominalQuarkusTest()
	quarkusTrait.QuarkusTrait.NativeBuilderImage = "overridden-value"
	environment.IntegrationKit = nil
	environment.Integration = &v1.Integration{
		Spec:   v1.IntegrationSpec{},
		Status: v1.IntegrationStatus{},
	}

	environment.ExecutedTraits = []Trait{}

	newKit, err := quarkusTrait.newIntegrationKit(environment, fastJarPackageType)
	assert.Nil(t, err)
	// The Quarkus trait was providing some value
	assert.Equal(t, &traitv1.QuarkusTrait{NativeBuilderImage: "overridden-value"}, newKit.Spec.Traits.Quarkus)
}

func TestQuarkusMatches(t *testing.T) {
	qt := quarkusTrait{
		BasePlatformTrait: NewBasePlatformTrait("quarkus", 600),
		QuarkusTrait: traitv1.QuarkusTrait{
			Modes: []traitv1.QuarkusMode{traitv1.JvmQuarkusMode},
		},
	}
	qt2 := quarkusTrait{
		BasePlatformTrait: NewBasePlatformTrait("quarkus", 600),
		QuarkusTrait: traitv1.QuarkusTrait{
			Modes: []traitv1.QuarkusMode{traitv1.JvmQuarkusMode},
		},
	}

	assert.True(t, qt.Matches(&qt2))
	qt2.Modes = append(qt2.Modes, traitv1.NativeQuarkusMode)
	assert.True(t, qt.Matches(&qt2))
	qt2.Modes = []traitv1.QuarkusMode{traitv1.NativeQuarkusMode}
	assert.False(t, qt.Matches(&qt2))
	qt2.Modes = nil
	assert.True(t, qt.Matches(&qt2))
	qt2.Modes = []traitv1.QuarkusMode{}
	assert.True(t, qt.Matches(&qt2))
}
