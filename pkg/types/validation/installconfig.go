package validation

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/bailey84j/terraform_installer/pkg/ipnet"
	"github.com/bailey84j/terraform_installer/pkg/types"
	"github.com/bailey84j/terraform_installer/pkg/types/aws"
	awsvalidation "github.com/bailey84j/terraform_installer/pkg/types/aws/validation"
	"github.com/bailey84j/terraform_installer/pkg/validate"
)

// list of known plugins that require hostPrefix to be set
//var pluginsUsingHostPrefix = sets.NewString(string(operv1.NetworkTypeOpenShiftSDN), string(operv1.NetworkTypeOVNKubernetes))

// ValidateInstallConfig checks that the specified install config is valid.
func ValidateInstallConfig(c *types.InstallConfig) field.ErrorList {
	allErrs := field.ErrorList{}
	if c.TypeMeta.APIVersion == "" {
		return field.ErrorList{field.Required(field.NewPath("apiVersion"), "install-config version required")}
	}
	switch v := c.APIVersion; v {
	case types.InstallConfigVersion:
		// Current version
	default:
		return field.ErrorList{field.Invalid(field.NewPath("apiVersion"), c.TypeMeta.APIVersion, fmt.Sprintf("install-config version must be %q", types.InstallConfigVersion))}
	}

	if c.SSHKey != "" {
		if err := validate.SSHPublicKey(c.SSHKey); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("sshKey"), c.SSHKey, err.Error()))
		}

	}
	/*
		if c.AdditionalTrustBundle != "" {
			if err := validate.CABundle(c.AdditionalTrustBundle); err != nil {
				allErrs = append(allErrs, field.Invalid(field.NewPath("additionalTrustBundle"), c.AdditionalTrustBundle, err.Error()))
			}
		}
		if c.AdditionalTrustBundlePolicy != "" {
			if err := validateAdditionalCABundlePolicy(c); err != nil {
				allErrs = append(allErrs, field.Invalid(field.NewPath("additionalTrustBundlePolicy"), c.AdditionalTrustBundlePolicy, err.Error()))
			}
		}
	*/
	nameErr := validate.ClusterName(c.ObjectMeta.Name)
	/*if c.Platform.GCP != nil || c.Platform.Azure != nil {
		nameErr = validate.ClusterName1035(c.ObjectMeta.Name)
	}
	if c.Platform.VSphere != nil || c.Platform.BareMetal != nil || c.Platform.OpenStack != nil || c.Platform.Nutanix != nil {
		nameErr = validate.OnPremClusterName(c.ObjectMeta.Name)
	}*/
	if nameErr != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("metadata", "name"), c.ObjectMeta.Name, nameErr.Error()))
	}
	baseDomainErr := validate.DomainName(c.BaseDomain, true)
	if baseDomainErr != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("baseDomain"), c.BaseDomain, baseDomainErr.Error()))
	}
	if nameErr == nil && baseDomainErr == nil {
		clusterDomain := c.ClusterDomain()
		if err := validate.DomainName(clusterDomain, true); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("baseDomain"), clusterDomain, err.Error()))
		}
	}
	/*
		if c.Networking != nil {
			allErrs = append(allErrs, validateNetworking(c.Networking, c.IsSingleNodeOpenShift(), field.NewPath("networking"))...)
			allErrs = append(allErrs, validateNetworkingIPVersion(c.Networking, &c.Platform)...)
			allErrs = append(allErrs, validateNetworkingForPlatform(c.Networking, &c.Platform, field.NewPath("networking"))...)
			allErrs = append(allErrs, validateVIPsForPlatform(c.Networking, &c.Platform, field.NewPath("platform"))...)
		} else {
			allErrs = append(allErrs, field.Required(field.NewPath("networking"), "networking is required"))
		}
		allErrs = append(allErrs, validatePlatform(&c.Platform, field.NewPath("platform"), c.Networking, c)...)
		if c.ControlPlane != nil {
			allErrs = append(allErrs, validateControlPlane(&c.Platform, c.ControlPlane, field.NewPath("controlPlane"))...)
		} else {
			allErrs = append(allErrs, field.Required(field.NewPath("controlPlane"), "controlPlane is required"))
		}
		allErrs = append(allErrs, validateCompute(&c.Platform, c.ControlPlane, c.Compute, field.NewPath("compute"))...)
		if err := validate.ImagePullSecret(c.PullSecret); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("pullSecret"), c.PullSecret, err.Error()))
		}
		if c.Proxy != nil {
			allErrs = append(allErrs, validateProxy(c.Proxy, c, field.NewPath("proxy"))...)
		}
		allErrs = append(allErrs, validateImageContentSources(c.ImageContentSources, field.NewPath("imageContentSources"))...)
		if _, ok := validPublishingStrategies[c.Publish]; !ok {
			allErrs = append(allErrs, field.NotSupported(field.NewPath("publish"), c.Publish, validPublishingStrategyValues))
		}
		allErrs = append(allErrs, validateCloudCredentialsMode(c.CredentialsMode, field.NewPath("credentialsMode"), c.Platform)...)
		if c.Capabilities != nil {
			allErrs = append(allErrs, validateCapabilities(c.Capabilities, field.NewPath("capabilities"))...)
		}
	*/

	if c.Publish == types.InternalPublishingStrategy {
		switch platformName := c.Platform.Name(); platformName {
		case aws.Name:
		default:
			allErrs = append(allErrs, field.Invalid(field.NewPath("publish"), c.Publish, fmt.Sprintf("Internal publish strategy is not supported on %q platform", platformName)))
		}
	}

	allErrs = append(allErrs)

	return allErrs
}

// ipAddressType indicates the address types provided for a given field
type ipAddressType struct {
	IPv4 bool
	IPv6 bool
	//Primary corev1.IPFamily
}

// ipAddressTypeByField is a map of field path to ipAddressType
type ipAddressTypeByField map[string]ipAddressType

// ipNetByField is a map of field path to the IPNets
type ipNetByField map[string][]ipnet.IPNet

func ipnetworksToStrings(networks []ipnet.IPNet) []string {
	var diag []string
	for _, sn := range networks {
		diag = append(diag, sn.String())
	}
	return diag
}

func validatePlatform(platform *types.Platform, fldPath *field.Path, network *types.Networking, c *types.InstallConfig) field.ErrorList {
	allErrs := field.ErrorList{}
	activePlatform := platform.Name()
	platforms := make([]string, len(types.PlatformNames))
	copy(platforms, types.PlatformNames)
	platforms = append(platforms, types.HiddenPlatformNames...)
	sort.Strings(platforms)
	i := sort.SearchStrings(platforms, activePlatform)
	if i == len(platforms) || platforms[i] != activePlatform {
		allErrs = append(allErrs, field.Invalid(fldPath, activePlatform, fmt.Sprintf("must specify one of the platforms (%s)", strings.Join(platforms, ", "))))
	}
	validate := func(n string, value interface{}, validation func(*field.Path) field.ErrorList) {
		if n != activePlatform {
			allErrs = append(allErrs, field.Invalid(fldPath, activePlatform, fmt.Sprintf("must only specify a single type of platform; cannot use both %q and %q", activePlatform, n)))
		}
		allErrs = append(allErrs, validation(fldPath.Child(n))...)
	}
	/*if platform.AlibabaCloud != nil {
		validate(alibabacloud.Name, platform.AlibabaCloud, func(f *field.Path) field.ErrorList {
			return alibabacloudvalidation.ValidatePlatform(platform.AlibabaCloud, network, f)
		})
	}*/
	if platform.AWS != nil {
		validate(aws.Name, platform.AWS, func(f *field.Path) field.ErrorList { return awsvalidation.ValidatePlatform(platform.AWS, f) })
	}
	/*if platform.Azure != nil {
		validate(azure.Name, platform.Azure, func(f *field.Path) field.ErrorList {
			return azurevalidation.ValidatePlatform(platform.Azure, c.Publish, f)
		})
	}
	if platform.GCP != nil {
		validate(gcp.Name, platform.GCP, func(f *field.Path) field.ErrorList { return gcpvalidation.ValidatePlatform(platform.GCP, f, c) })
	}
	if platform.IBMCloud != nil {
		validate(ibmcloud.Name, platform.IBMCloud, func(f *field.Path) field.ErrorList { return ibmcloudvalidation.ValidatePlatform(platform.IBMCloud, f) })
	}
	if platform.Libvirt != nil {
		validate(libvirt.Name, platform.Libvirt, func(f *field.Path) field.ErrorList { return libvirtvalidation.ValidatePlatform(platform.Libvirt, f) })
	}
	if platform.OpenStack != nil {
		validate(openstack.Name, platform.OpenStack, func(f *field.Path) field.ErrorList {
			return openstackvalidation.ValidatePlatform(platform.OpenStack, network, f, c)
		})
	}
	if platform.PowerVS != nil {
		validate(powervs.Name, platform.PowerVS, func(f *field.Path) field.ErrorList { return powervsvalidation.ValidatePlatform(platform.PowerVS, f) })
	}
	if platform.VSphere != nil {
		validate(vsphere.Name, platform.VSphere, func(f *field.Path) field.ErrorList {
			return vspherevalidation.ValidatePlatform(platform.VSphere, f)
		})
	}
	if platform.BareMetal != nil {
		validate(baremetal.Name, platform.BareMetal, func(f *field.Path) field.ErrorList {
			return baremetalvalidation.ValidatePlatform(platform.BareMetal, network, f, c)
		})
	}
	if platform.Ovirt != nil {
		validate(ovirt.Name, platform.Ovirt, func(f *field.Path) field.ErrorList {
			return ovirtvalidation.ValidatePlatform(platform.Ovirt, f)
		})
	}
	if platform.Nutanix != nil {
		validate(nutanix.Name, platform.Nutanix, func(f *field.Path) field.ErrorList {
			return nutanixvalidation.ValidatePlatform(platform.Nutanix, f)
		})
	}*/
	return allErrs
}

func validateCloudCredentialsMode(mode types.CredentialsMode, fldPath *field.Path, platform types.Platform) field.ErrorList {
	if mode == "" {
		return nil
	}
	allErrs := field.ErrorList{}

	//allowedAzureModes := []types.CredentialsMode{types.PassthroughCredentialsMode, types.ManualCredentialsMode}
	//if platform.Azure != nil && platform.Azure.CloudName == azure.StackCloud {
	//		allowedAzureModes = []types.CredentialsMode{types.ManualCredentialsMode}
	//	}

	// validPlatformCredentialsModes is a map from the platform name to a slice of credentials modes that are valid
	// for the platform. If a platform name is not in the map, then the credentials mode cannot be set for that platform.
	validPlatformCredentialsModes := map[string][]types.CredentialsMode{
		//alibabacloud.Name: {types.ManualCredentialsMode},
		aws.Name: {types.MintCredentialsMode, types.PassthroughCredentialsMode, types.ManualCredentialsMode},
		//azure.Name:        allowedAzureModes,
		//gcp.Name:          {types.MintCredentialsMode, types.PassthroughCredentialsMode, types.ManualCredentialsMode},
		//ibmcloud.Name:     {types.ManualCredentialsMode},
		//powervs.Name:      {types.ManualCredentialsMode},
		//nutanix.Name:      {types.ManualCredentialsMode},
	}
	if validModes, ok := validPlatformCredentialsModes[platform.Name()]; ok {
		validModesSet := sets.NewString()
		for _, m := range validModes {
			validModesSet.Insert(string(m))
		}
		if !validModesSet.Has(string(mode)) {
			allErrs = append(allErrs, field.NotSupported(fldPath, mode, validModesSet.List()))
		}
	} else {
		allErrs = append(allErrs, field.Invalid(fldPath, mode, fmt.Sprintf("cannot be set when using the %q platform", platform.Name())))
	}
	return allErrs
}

// validateURI checks if the given url is of the right format. It also checks if the scheme of the uri
// provided is within the list of accepted schema provided as part of the input.
func validateURI(uri string, fldPath *field.Path, schemes []string) field.ErrorList {
	parsed, err := url.ParseRequestURI(uri)
	if err != nil {
		return field.ErrorList{field.Invalid(fldPath, uri, err.Error())}
	}
	for _, scheme := range schemes {
		if scheme == parsed.Scheme {
			return nil
		}
	}
	return field.ErrorList{field.NotSupported(fldPath, parsed.Scheme, schemes)}
}
