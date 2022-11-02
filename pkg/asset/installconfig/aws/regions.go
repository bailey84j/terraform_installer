package aws

import (
	"github.com/bailey84j/terraform_installer/pkg/types"
)

// knownPublicRegions is the subset of public AWS regions where RHEL CoreOS images are published.
// This subset does not include supported regions which are found in other partitions, such as us-gov-east-1.
// Returns: a map of region identifier to region description.
func knownPublicRegions(architecture types.Architecture) map[string]string {
	//	required := rhcos.AMIRegions(architecture)

	regions := map[string]string{
		"us-east-1": "Description of us-east-1",
	}

	return regions
}

// IsKnownPublicRegion returns true if a specified region is Known to the installer.
// A known region is the subset of public AWS regions where RHEL CoreOS images are published.
func IsKnownPublicRegion(region string, architecture types.Architecture) bool {
	if _, ok := knownPublicRegions(architecture)[region]; ok {
		return true
	}
	return false
}
