package installconfig

import (
	"fmt"
	"sort"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/bailey84j/terraform_installer/pkg/asset"
	awsconfig "github.com/bailey84j/terraform_installer/pkg/asset/installconfig/aws"

	//azureconfig "github.com/bailey84j/terraform_installer/pkg/asset/installconfig/azure"

	"github.com/bailey84j/terraform_installer/pkg/types"
	"github.com/bailey84j/terraform_installer/pkg/types/aws"
	//	"github.com/bailey84j/terraform_installer/pkg/types/azure"
)

// Platform is an asset that queries the user for the platform on which to install
// the cluster.
type platform struct {
	types.Platform
}

var _ asset.Asset = (*platform)(nil)

// Dependencies returns no dependencies.
func (a *platform) Dependencies() []asset.Asset {
	return []asset.Asset{}
}

// Generate queries for input from the user.
func (a *platform) Generate(asset.Parents) error {
	platform, err := a.queryUserForPlatform()
	if err != nil {
		return err
	}

	switch platform {
	/*case alibabacloud.Name:
	a.AlibabaCloud, err = alibabacloudconfig.Platform()
	if err != nil {
		return err
	}*/
	case aws.Name:
		logrus.Debugf("Trace Me - In Platform Switch - AWS...")
		a.AWS, err = awsconfig.Platform()
		if err != nil {
			return err
		}
	/*case azure.Name:
	a.Azure, err = azureconfig.Platform()
	if err != nil {
		return err
	}*/
	/*case baremetal.Name:
		a.BareMetal, err = baremetalconfig.Platform()
		if err != nil {
			return err
		}
	case gcp.Name:
		a.GCP, err = gcpconfig.Platform()
		if err != nil {
			return err
		}
	case ibmcloud.Name:
		a.IBMCloud, err = ibmcloudconfig.Platform()
		if err != nil {
			return err
		}
	case libvirt.Name:
		a.Libvirt, err = libvirtconfig.Platform()
		if err != nil {
			return err
		}
	case none.Name:
		a.None = &none.Platform{}
	case openstack.Name:
		a.OpenStack, err = openstackconfig.Platform()
		if err != nil {
			return err
		}
	case ovirt.Name:
		a.Ovirt, err = ovirtconfig.Platform()
		if err != nil {
			return err
		}
	case vsphere.Name:
		a.VSphere, err = vsphereconfig.Platform()
		if err != nil {
			return err
		}
	case powervs.Name:
		a.PowerVS, err = powervsconfig.Platform()
		if err != nil {
			return err
		}
	case nutanix.Name:
		a.Nutanix, err = nutanixconfig.Platform()
		if err != nil {
			return err
		}*/
	default:
		return fmt.Errorf("unknown platform type %q", platform)
	}

	return nil
}

// Name returns the human-friendly name of the asset.
func (a *platform) Name() string {
	return "Platform"
}

func (a *platform) queryUserForPlatform() (platform string, err error) {
	err = survey.Ask([]*survey.Question{
		{
			Prompt: &survey.Select{
				Message: "Platform",
				Options: types.PlatformNames,
				Help:    "The platform on which the cluster will run.  For a full list of platforms, including those not supported by this wizard, see https://github.com/bailey84j/terraform_installer",
			},
			Validate: survey.ComposeValidators(survey.Required, func(ans interface{}) error {
				choice := ans.(core.OptionAnswer).Value
				i := sort.SearchStrings(types.PlatformNames, choice)
				if i == len(types.PlatformNames) || types.PlatformNames[i] != choice {
					return errors.Errorf("invalid platform %q", choice)
				}
				return nil
			}),
		},
	}, &platform)
	return
}

func (a *platform) CurrentName() string {
	return a.Platform.Name()
}
