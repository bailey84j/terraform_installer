// Package aws contains AWS-specific Terraform-variable logic.
package aws

type config struct {
	AMI                          string            `json:"aws_ami"`
	AMIRegion                    string            `json:"aws_ami_region"`
	CustomEndpoints              map[string]string `json:"custom_endpoints,omitempty"`
	ExtraTags                    map[string]string `json:"aws_extra_tags,omitempty"`
	BootstrapInstanceType        string            `json:"aws_bootstrap_instance_type,omitempty"`
	MasterInstanceType           string            `json:"aws_master_instance_type,omitempty"`
	MasterAvailabilityZones      []string          `json:"aws_master_availability_zones"`
	WorkerAvailabilityZones      []string          `json:"aws_worker_availability_zones"`
	IOPS                         int64             `json:"aws_master_root_volume_iops"`
	Size                         int64             `json:"aws_master_root_volume_size,omitempty"`
	Type                         string            `json:"aws_master_root_volume_type,omitempty"`
	Encrypted                    bool              `json:"aws_master_root_volume_encrypted"`
	KMSKeyID                     string            `json:"aws_master_root_volume_kms_key_id,omitempty"`
	Region                       string            `json:"aws_region,omitempty"`
	VPC                          string            `json:"aws_vpc,omitempty"`
	PrivateSubnets               []string          `json:"aws_private_subnets,omitempty"`
	PublicSubnets                *[]string         `json:"aws_public_subnets,omitempty"`
	InternalZone                 string            `json:"aws_internal_zone,omitempty"`
	PublishStrategy              string            `json:"aws_publish_strategy,omitempty"`
	IgnitionBucket               string            `json:"aws_ignition_bucket"`
	BootstrapIgnitionStub        string            `json:"aws_bootstrap_stub_ignition"`
	MasterIAMRoleName            string            `json:"aws_master_iam_role_name,omitempty"`
	WorkerIAMRoleName            string            `json:"aws_worker_iam_role_name,omitempty"`
	MasterMetadataAuthentication string            `json:"aws_master_instance_metadata_authentication,omitempty"`
}

// TFVarsSources contains the parameters to be converted into Terraform variables
type TFVarsSources struct {
	VPC string
}

// TFVars generates AWS-specific Terraform variables launching the cluster.
func TFVars(sources TFVarsSources) ([]byte, error) {
	masterConfig := sources.MasterConfigs[0]
}
