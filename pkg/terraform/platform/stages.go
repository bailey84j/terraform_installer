package platform

import (
	"fmt"

	"github.com/bailey84j/terraform_installer/pkg/terraform"
	"github.com/bailey84j/terraform_installer/pkg/terraform/stages/aws"
	"github.com/bailey84j/terraform_installer/pkg/terraform/stages/azure"
	awstypes "github.com/bailey84j/terraform_installer/pkg/types/aws"
	azuretypes "github.com/bailey84j/terraform_installer/pkg/types/azure"
)

// StagesForPlatform returns the terraform stages to run to provision the infrastructure for the specified platform.
func StagesForPlatform(platform string) []terraform.Stage {
	switch platform {
	/*case alibabacloudtypes.Name:
	return alibabacloud.PlatformStages
	*/
	case awstypes.Name:
		return aws.PlatformStages
	case azuretypes.Name:
		return azure.PlatformStages
	case azuretypes.StackTerraformName:
		return azure.StackPlatformStages
	/*case baremetaltypes.Name:
		return baremetal.PlatformStages
	case gcptypes.Name:
		return gcp.PlatformStages
	case ibmcloudtypes.Name:
		return ibmcloud.PlatformStages
	case libvirttypes.Name:
		return libvirt.PlatformStages
	case nutanixtypes.Name:
		return nutanix.PlatformStages
	case powervstypes.Name:
		return powervs.PlatformStages
	case openstacktypes.Name:
		return openstack.PlatformStages
	case ovirttypes.Name:
		return ovirt.PlatformStages
	case vspheretypes.Name:
		return vsphere.PlatformStages
	case vspheretypes.ZoningTerraformName:
		return vsphere.ZoningPlatformStages
	case nonetypes.Name:
		// terraform is not used when the platform is "none"
		return []terraform.Stage{}
	*/
	default:
		panic(fmt.Sprintf("unsupported platform %q", platform))
	}
}
