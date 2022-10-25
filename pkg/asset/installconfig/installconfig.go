package installconfig

const (
	installConfigFilename = "install-config.yaml"
)

// InstallConfig generates the install-config.yaml file.
type InstallConfig struct {
	Config       *types.InstallConfig   `json:"config"`
	File         *asset.File            `json:"file"`
	AWS          *aws.Metadata          `json:"aws,omitempty"`
	Azure        *icazure.Metadata      `json:"azure,omitempty"`
	IBMCloud     *icibmcloud.Metadata   `json:"ibmcloud,omitempty"`
	AlibabaCloud *alibabacloud.Metadata `json:"alibabacloud,omitempty"`
	PowerVS      *icpowervs.Metadata    `json:"powervs,omitempty"`
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
	networking := &networking{}
	pullSecret := &pullSecret{}
	platform := &platform{}
	parents.Get(
		sshPublicKey,
		baseDomain,
		clusterName,
		networking,
		pullSecret,
		platform,
	)

	a.Config = &types.InstallConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: types.InstallConfigVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName.ClusterName,
		},
		SSHKey:     sshPublicKey.Key,
		BaseDomain: baseDomain.BaseDomain,
		PullSecret: pullSecret.PullSecret,
		Networking: &types.Networking{
			MachineNetwork: networking.machineNetwork,
		},
	}

	a.Config.AlibabaCloud = platform.AlibabaCloud
	a.Config.AWS = platform.AWS
	a.Config.Libvirt = platform.Libvirt
	a.Config.None = platform.None
	a.Config.OpenStack = platform.OpenStack
	a.Config.VSphere = platform.VSphere
	a.Config.Azure = platform.Azure
	a.Config.GCP = platform.GCP
	a.Config.IBMCloud = platform.IBMCloud
	a.Config.BareMetal = platform.BareMetal
	a.Config.Ovirt = platform.Ovirt
	a.Config.PowerVS = platform.PowerVS
	a.Config.Nutanix = platform.Nutanix

	return a.finish("")
}
