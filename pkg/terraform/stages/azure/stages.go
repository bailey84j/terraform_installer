package azure

import (
	"github.com/bailey84j/terraform_installer/pkg/terraform"
	"github.com/bailey84j/terraform_installer/pkg/terraform/providers"
	"github.com/bailey84j/terraform_installer/pkg/terraform/stages"
	typesazure "github.com/bailey84j/terraform_installer/pkg/types/azure"
)

// PlatformStages are the stages to run to provision the infrastructure in Azure.
var PlatformStages = []terraform.Stage{
	stages.NewStage(
		typesazure.Name,
		"vnet",
		[]providers.Provider{providers.AzureRM},
	),
	stages.NewStage(
		typesazure.Name,
		"bootstrap",
		[]providers.Provider{providers.AzureRM},
		stages.WithNormalBootstrapDestroy(),
	),
	stages.NewStage(
		typesazure.Name,
		"cluster",
		[]providers.Provider{providers.AzureRM},
	),
}

// StackPlatformStages are the stages to run to provision the infrastructure in Azure Stack.
var StackPlatformStages = []terraform.Stage{
	stages.NewStage(
		typesazure.StackTerraformName,
		"vnet",
		[]providers.Provider{providers.AzureStack},
	),
	stages.NewStage(
		typesazure.StackTerraformName,
		"bootstrap",
		[]providers.Provider{providers.AzureStack},
		stages.WithNormalBootstrapDestroy(),
	),
	stages.NewStage(
		typesazure.StackTerraformName,
		"cluster",
		[]providers.Provider{providers.AzureStack},
	),
}
