package targets

import (
	"github.com/bailey84j/terraform_installer/pkg/asset"
	"github.com/bailey84j/terraform_installer/pkg/asset/cluster"
	"github.com/bailey84j/terraform_installer/pkg/asset/installconfig"
	"github.com/bailey84j/terraform_installer/pkg/asset/password"
)

var (
	// InstallConfig are the install-config targeted assets.
	InstallConfig = []asset.WritableAsset{
		&installconfig.InstallConfig{},
	}

	// Manifests are the manifests targeted assets.
	Manifests = []asset.WritableAsset{
		//&machines.Master{},
		//&machines.Worker{},
		//&manifests.Manifests{},
		//&manifests.Openshift{},
	}

	// ManifestTemplates are the manifest-templates targeted assets.
	ManifestTemplates = []asset.WritableAsset{
		/*&bootkube.KubeCloudConfig{},
		&bootkube.MachineConfigServerTLSSecret{},
		&bootkube.CVOOverrides{},
		&bootkube.KubeSystemConfigmapRootCA{},
		&bootkube.OpenshiftConfigSecretPullSecret{},
		&openshift.CloudCredsSecret{},
		&openshift.TFEPasswordSecret{},
		&openshift.RoleCloudCredsSecretReader{},
		&openshift.AzureCloudProviderSecret{},*/
	}

	// IgnitionConfigs are the ignition-configs targeted assets.
	IgnitionConfigs = []asset.WritableAsset{
		/*&kubeconfig.AdminClient{},
		&password.TFEPassword{},
		&machine.Master{},
		&machine.Worker{},
		&bootstrap.Bootstrap{},
		&cluster.Metadata{},*/
	}

	// SingleNodeIgnitionConfig is the bootstrap-in-place ignition-config targeted assets.
	SingleNodeIgnitionConfig = []asset.WritableAsset{
		/*&kubeconfig.AdminClient{},
		&password.TFEPassword{},
		&machine.Worker{},
		&bootstrap.SingleNodeBootstrapInPlace{},
		&cluster.Metadata{},*/
	}

	// Cluster are the cluster targeted assets.
	Cluster = []asset.WritableAsset{
		//&cluster.Metadata{},
		//&machine.MasterIgnitionCustomizations{},
		//&machine.WorkerIgnitionCustomizations{},
		&cluster.TerraformVariables{},
		//&kubeconfig.AdminClient{},
		&password.TFEPassword{},
		//&tls.JournalCertKey{},
		&cluster.Cluster{},
	}
)
