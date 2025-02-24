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

package azure

import (
	"strconv"

	"github.com/apache/camel-k/v2/pkg/util/kubernetes"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	traitv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1/trait"
	"github.com/apache/camel-k/v2/pkg/trait"
	"github.com/apache/camel-k/v2/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// The Azure Key Vault trait can be used to use secrets from Azure Key Vault service
//
// The Azure Key Vault trait is disabled by default.
//
// For more information about how to use secrets from Azure Key Vault component take a look at the components docs: xref:components::azure-key-vault-component.adoc[Azure Key Vault component]
//
// A sample execution of this trait, would require
// the following trait options:
// -t azure-key-vault.enabled=true -t azure-key-vault.tenant-id="tenant-id" -t azure-key-vault.client-id="client-id" -t azure-key-vault.client-secret="client-secret" -t azure-key-vault.vault-name="vault-name"
//
// To enable the automatic context reload on secrets updates you should define
// the following trait options:
// -t azure-key-vault.enabled=true -t azure-key-vault.tenant-id="tenant-id" -t azure-key-vault.client-id="client-id" -t azure-key-vault.client-secret="client-secret" -t azure-key-vault.vault-name="vault-name" -t azure-key-vault.context-reload-enabled="true" -t azure-key-vault.refresh-enabled="true" -t azure-key-vault.refresh-period="30000" -t azure-key-vault.secrets="test*" -t azure-key-vault.eventhub-connection-string="connection-string" -t azure-key-vault.blob-account-name="account-name"  -t azure-key-vault.blob-container-name="container-name"  -t azure-key-vault.blob-access-key="account-name" -t azure-key-vault.azure-identity-enabled="true"
//
// WARNING: The trait is **deprecated** and will removed in future release versions: configure directly the Camel properties as required by the component instead.
//
// +camel-k:trait=azure-key-vault.
// +camel-k:deprecated=2.5.0.
type Trait struct {
	traitv1.Trait `property:",squash"`
	// Enables automatic configuration of the trait.
	Auto *bool `property:"auto" json:"auto,omitempty"`
	// The Azure Tenant Id for accessing Key Vault
	TenantID string `property:"tenant-id" json:"tenantId,omitempty"`
	// The Azure Client Id for accessing Key Vault
	ClientID string `property:"client-id" json:"clientId,omitempty"`
	// The Azure Client Secret for accessing Key Vault. This could be a plain text or a configmap/secret.
	// The content of the azure key vault client secret is expected to be a text containing a valid Client Secret.
	// Syntax: [configmap|secret]:name[/key], where name represents the resource name, key optionally represents the resource key to be filtered (default key value = azure-key-vault-client-secret).
	ClientSecret string `property:"client-secret" json:"clientSecret,omitempty"`
	// The Azure Vault Name for accessing Key Vault
	VaultName string `property:"vault-name" json:"vaultName,omitempty"`
	// Define if we want to use the Camel Context Reload feature or not
	ContextReloadEnabled *bool `property:"context-reload-enabled" json:"contextReloadEnabled,omitempty"`
	// Define if we want to use the Refresh Feature for secrets
	RefreshEnabled *bool `property:"refresh-enabled" json:"refreshEnabled,omitempty"`
	// Whether the Azure Identity Authentication should be used or not
	AzureIdentityEnabled *bool `property:"azure-identity-enabled" json:"azureIdentityEnabled,omitempty"`
	// If Refresh is enabled, this defines the interval to check the refresh event
	RefreshPeriod string `property:"refresh-period" json:"refreshPeriod,omitempty"`
	// If Refresh is enabled, the regular expression representing the secrets we want to track
	Secrets string `property:"secrets" json:"secrets,omitempty"`
	// If Refresh is enabled, the connection String to point to the Eventhub service used to track updates
	EventhubConnectionString string `property:"eventhub-connection-string" json:"eventhubConnectionString,omitempty"`
	// If Refresh is enabled, the account name for Azure Storage Blob service used to save checkpoint while consuming from Eventhub
	BlobAccountName string `property:"blob-account-name" json:"blobAccountName,omitempty"`
	// If Refresh is enabled, the access key for Azure Storage Blob service used to save checkpoint while consuming from Eventhub. This could be a plain text or a configmap/secret.
	// The content of the azure key vault blob access key is expected to be a text containing a valid Access Key for Azure Storage Blob.
	// Syntax: [configmap|secret]:name[/key], where name represents the resource name, key optionally represents the resource key to be filtered (default key value = azure-storage-blob-access-key).
	BlobAccessKey string `property:"blob-access-key" json:"blobAccessKey,omitempty"`
	// If Refresh is enabled, the container name for Azure Storage Blob service used to save checkpoint while consuming from Eventhub
	BlobContainerName string `property:"blob-container-name" json:"blobContainerName,omitempty"`
}

type azureKeyVaultTrait struct {
	trait.BaseTrait
	Trait `property:",squash"`
}

func NewAzureKeyVaultTrait() trait.Trait {
	return &azureKeyVaultTrait{
		BaseTrait: trait.NewBaseTrait("azure-key-vault", trait.TraitOrderBeforeControllerCreation),
	}
}

func (t *azureKeyVaultTrait) Configure(environment *trait.Environment) (bool, *trait.TraitCondition, error) {
	if environment.Integration == nil || !ptr.Deref(t.Enabled, false) {
		return false, nil, nil
	}

	if !environment.IntegrationInPhase(v1.IntegrationPhaseInitialization) && !environment.IntegrationInRunningPhases() {
		return false, nil, nil
	}

	if t.ContextReloadEnabled == nil {
		t.ContextReloadEnabled = ptr.To(false)
	}

	if t.RefreshEnabled == nil {
		t.RefreshEnabled = ptr.To(false)
	}

	if t.AzureIdentityEnabled == nil {
		t.AzureIdentityEnabled = ptr.To(false)
	}

	condition := trait.NewIntegrationCondition(
		"AzureKeyVault",
		v1.IntegrationConditionTraitInfo,
		corev1.ConditionTrue,
		trait.TraitConfigurationReason,
		"AzureKeyVault trait is deprecated and may be removed in future version: "+
			"configure directly the Camel properties as required by the component instead",
	)

	return true, condition, nil
}

func (t *azureKeyVaultTrait) Apply(environment *trait.Environment) error {

	if environment.IntegrationInPhase(v1.IntegrationPhaseInitialization) {
		util.StringSliceUniqueAdd(&environment.Integration.Status.Capabilities, v1.CapabilityAzureKeyVault)
	}

	if !environment.IntegrationInRunningPhases() {
		return nil
	}

	hits := v1.PlainConfigSecretRegexp.FindAllStringSubmatch(t.ClientSecret, -1)
	if len(hits) >= 1 {
		var res, _ = v1.DecodeValueSource(t.ClientSecret, "azure-key-vault-client-secret")
		if secretValue, err := kubernetes.ResolveValueSource(environment.Ctx, environment.Client, environment.Platform.Namespace, &res); err != nil {
			return err
		} else if secretValue != "" {
			environment.ApplicationProperties["camel.vault.azure.clientSecret"] = string([]byte(secretValue))
		}
	} else {
		environment.ApplicationProperties["camel.vault.azure.clientSecret"] = t.ClientSecret
	}
	hits = v1.PlainConfigSecretRegexp.FindAllStringSubmatch(t.BlobAccessKey, -1)
	if len(hits) >= 1 {
		var res, _ = v1.DecodeValueSource(t.BlobAccessKey, "azure-storage-blob-access-key")
		if secretValue, err := kubernetes.ResolveValueSource(environment.Ctx, environment.Client, environment.Platform.Namespace, &res); err != nil {
			return err
		} else if secretValue != "" {
			environment.ApplicationProperties["camel.vault.azure.blobAccessKey"] = string([]byte(secretValue))
		}
	} else {
		environment.ApplicationProperties["camel.vault.azure.blobAccessKey"] = t.BlobAccessKey
	}
	environment.ApplicationProperties["camel.vault.azure.tenantId"] = t.TenantID
	environment.ApplicationProperties["camel.vault.azure.clientId"] = t.ClientID
	environment.ApplicationProperties["camel.vault.azure.vaultName"] = t.VaultName
	environment.ApplicationProperties["camel.vault.azure.refreshEnabled"] = strconv.FormatBool(*t.RefreshEnabled)
	environment.ApplicationProperties["camel.vault.azure.azureIdentityEnabled"] = strconv.FormatBool(*t.AzureIdentityEnabled)
	environment.ApplicationProperties["camel.main.context-reload-enabled"] = strconv.FormatBool(*t.ContextReloadEnabled)
	environment.ApplicationProperties["camel.vault.azure.refreshPeriod"] = t.RefreshPeriod
	if t.Secrets != "" {
		environment.ApplicationProperties["camel.vault.azure.secrets"] = t.Secrets
	}
	environment.ApplicationProperties["camel.vault.azure.eventhubConnectionString"] = t.EventhubConnectionString
	environment.ApplicationProperties["camel.vault.azure.blobAccountName"] = t.BlobAccountName
	environment.ApplicationProperties["camel.vault.azure.blobContainerName"] = t.BlobContainerName

	return nil
}
