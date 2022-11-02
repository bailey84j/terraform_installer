package installconfig

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"

	"github.com/bailey84j/terraform_installer/pkg/asset"
	"github.com/bailey84j/terraform_installer/pkg/asset/installconfig/aws"
	"github.com/bailey84j/terraform_installer/pkg/types"
	"github.com/bailey84j/terraform_installer/pkg/types/conversion"
	"github.com/bailey84j/terraform_installer/pkg/types/defaults"
	"github.com/bailey84j/terraform_installer/pkg/types/validation"
	//icazure "github.com/bailey84j/terraform_installer/pkg/asset/installconfig/azure"
)

const (
	installConfigFilename = "install-config.yaml"
)

// InstallConfig generates the install-config.yaml file.
type InstallConfig struct {
	Config *types.InstallConfig `json:"config"`
	File   *asset.File          `json:"file"`
	AWS    *aws.Metadata        `json:"aws,omitempty"`
	//Azure  *icazure.Metadata    `json:"azure,omitempty"`
}

var _ asset.WritableAsset = (*InstallConfig)(nil)

// Dependencies returns all of the dependencies directly needed by an
// InstallConfig asset.
func (a *InstallConfig) Dependencies() []asset.Asset {
	return []asset.Asset{
		&sshPublicKey{},
		&baseDomain{},
		&clusterName{},
		&networking{},
		&pullSecret{},
		&platform{},
	}
}

// Generate generates the install-config.yaml file.
func (a *InstallConfig) Generate(parents asset.Parents) error {
	sshPublicKey := &sshPublicKey{}
	baseDomain := &baseDomain{}
	clusterName := &clusterName{}
	//networking := &networking{}
	pullSecret := &pullSecret{}
	platform := &platform{}

	parents.Get(
		sshPublicKey,
		baseDomain,
		clusterName,
		//networking,
		pullSecret,
		platform,
	)

	logrus.Debugf("Trace Me - clusterName.ClusterName - %v", clusterName)
	a.Config = &types.InstallConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: types.InstallConfigVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName.ClusterName,
		},
		SSHKey:     sshPublicKey.Key,
		BaseDomain: baseDomain.BaseDomain,
		Licence:    pullSecret.PullSecret,
	}

	logrus.Debugf("Trace Me - config - %+v", a.Config)
	//a.Config.AlibabaCloud = platform.AlibabaCloud
	a.Config.AWS = platform.AWS
	/*
		a.Config.Libvirt = platform.Libvirt
		a.Config.None = platform.None
		a.Config.OpenStack = platform.OpenStack
		a.Config.VSphere = platform.VSphere
	*/
	a.Config.Azure = platform.Azure
	/*
		a.Config.GCP = platform.GCP
		a.Config.IBMCloud = platform.IBMCloud
		a.Config.BareMetal = platform.BareMetal
		a.Config.Ovirt = platform.Ovirt
		a.Config.PowerVS = platform.PowerVS
		a.Config.Nutanix = platform.Nutanix
	*/
	return a.finish("")
}

func (a *InstallConfig) finish(filename string) error {
	defaults.SetInstallConfigDefaults(a.Config)

	if a.Config.AWS != nil {
		a.AWS = aws.NewMetadata(a.Config.Platform.AWS.Region, a.Config.Platform.AWS.Subnets, a.Config.AWS.ServiceEndpoints)
	}
	/*
		if a.Config.AlibabaCloud != nil {
			a.AlibabaCloud = alibabacloud.NewMetadata(a.Config.AlibabaCloud.Region, a.Config.AlibabaCloud.VSwitchIDs)
		}
		if a.Config.Azure != nil {
			a.Azure = icazure.NewMetadata(a.Config.Azure.CloudName, a.Config.Azure.ARMEndpoint)
		}
		if a.Config.IBMCloud != nil {
			a.IBMCloud = icibmcloud.NewMetadata(a.Config.BaseDomain, a.Config.IBMCloud.Region, a.Config.IBMCloud.ControlPlaneSubnets, a.Config.IBMCloud.ComputeSubnets)
		}
		if a.Config.PowerVS != nil {
			a.PowerVS = icpowervs.NewMetadata(a.Config.BaseDomain)
		}
	*/

	if err := validation.ValidateInstallConfig(a.Config).ToAggregate(); err != nil {
		if filename == "" {
			return errors.Wrap(err, "invalid install config")
		}
		return errors.Wrapf(err, "invalid %q file", filename)
	}

	if err := a.platformValidation(); err != nil {
		return err
	}

	data, err := yaml.Marshal(a.Config)
	if err != nil {
		return errors.Wrap(err, "failed to Marshal InstallConfig")
	}
	a.File = &asset.File{
		Filename: installConfigFilename,
		Data:     data,
	}
	return nil
}

func (a *InstallConfig) platformValidation() error {
	/*
		if a.Config.Platform.AlibabaCloud != nil {
			client, err := a.AlibabaCloud.Client()
			if err != nil {
				return err
			}
			return alibabacloud.Validate(client, a.Config)
		}
		if a.Config.Platform.Azure != nil {
			client, err := a.Azure.Client()
			if err != nil {
				return err
			}
			return icazure.Validate(client, a.Config)
		}
		if a.Config.Platform.GCP != nil {
			client, err := icgcp.NewClient(context.TODO())
			if err != nil {
				return err
			}
			return icgcp.Validate(client, a.Config)
		}
		if a.Config.Platform.IBMCloud != nil {
			client, err := icibmcloud.NewClient()
			if err != nil {
				return err
			}
			return icibmcloud.Validate(client, a.Config)
		}*/
	if a.Config.Platform.AWS != nil {
		return aws.Validate(context.TODO(), a.AWS, a.Config)
	}
	/*
		if a.Config.Platform.VSphere != nil {
			return icvsphere.Validate(a.Config)
		}
		if a.Config.Platform.Ovirt != nil {
			return icovirt.Validate(a.Config)
		}
		if a.Config.Platform.OpenStack != nil {
			return icopenstack.Validate(a.Config)
		}
		if a.Config.Platform.PowerVS != nil {
			return icpowervs.Validate(a.Config)
		}
		if a.Config.Platform.Nutanix != nil {
			return icnutanix.Validate(a.Config)
		}*/
	return field.ErrorList{}.ToAggregate()
}

// Name returns the human-friendly name of the asset.
func (a *InstallConfig) Name() string {
	return "Install Config"
}

// Files returns the files generated by the asset.
func (a *InstallConfig) Files() []*asset.File {
	if a.File != nil {
		return []*asset.File{a.File}
	}
	return []*asset.File{}
}

// Load returns the installconfig from disk.
func (a *InstallConfig) Load(f asset.FileFetcher) (found bool, err error) {
	file, err := f.FetchByName(installConfigFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, asset.InstallConfigError)
	}

	config := &types.InstallConfig{}
	if err := yaml.UnmarshalStrict(file.Data, config, yaml.DisallowUnknownFields); err != nil {
		err = errors.Wrapf(err, "failed to unmarshal %s", installConfigFilename)
		if !strings.Contains(err.Error(), "unknown field") {
			return false, errors.Wrap(err, asset.InstallConfigError)
		}
		err = errors.Wrapf(err, "failed to parse first occurence of unknown field")
		logrus.Warnf(err.Error())
		logrus.Info("Attempting to unmarshal while ignoring unknown keys because strict unmarshaling failed")
		if err = yaml.UnmarshalStrict(file.Data, config); err != nil {
			err = errors.Wrapf(err, "failed to unmarshal %s", installConfigFilename)
			return false, errors.Wrap(err, asset.InstallConfigError)
		}
	}
	a.Config = config

	// Upconvert any deprecated fields
	if err := conversion.ConvertInstallConfig(a.Config); err != nil {
		return false, errors.Wrap(errors.Wrap(err, "failed to upconvert install config"), asset.InstallConfigError)
	}

	err = a.finish(installConfigFilename)
	if err != nil {
		return false, errors.Wrap(err, asset.InstallConfigError)
	}
	return true, nil
}
