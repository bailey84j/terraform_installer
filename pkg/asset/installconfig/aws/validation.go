package aws

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/bailey84j/terraform_installer/pkg/types"
	awstypes "github.com/bailey84j/terraform_installer/pkg/types/aws"
)

type resourceRequirements struct {
	minimumVCpus  int64
	minimumMemory int64
}

var controlPlaneReq = resourceRequirements{
	minimumVCpus:  4,
	minimumMemory: 16384,
}

var computeReq = resourceRequirements{
	minimumVCpus:  2,
	minimumMemory: 8192,
}

// Validate executes platform-specific validation.
func Validate(ctx context.Context, meta *Metadata, config *types.InstallConfig) error {
	allErrs := field.ErrorList{}

	if config.Platform.AWS == nil {
		return errors.New(field.Required(field.NewPath("platform", "aws"), "AWS validation requires an AWS platform configuration").Error())
	}
	allErrs = append(allErrs, validateAMI(ctx, config)...)
	//allErrs = append(allErrs, validatePlatform(ctx, meta, field.NewPath("platform", "aws"), config.Platform.AWS, config.Networking, config.Publish)...)
	/*
		if config.ControlPlane != nil && config.ControlPlane.Platform.AWS != nil {
			allErrs = append(allErrs, validateMachinePool(ctx, meta, field.NewPath("controlPlane", "platform", "aws"), config.Platform.AWS, config.ControlPlane.Platform.AWS, controlPlaneReq)...)
		}
		for idx, compute := range config.Compute {
			fldPath := field.NewPath("compute").Index(idx)
			if compute.Platform.AWS != nil {
				allErrs = append(allErrs, validateMachinePool(ctx, meta, fldPath.Child("platform", "aws"), config.Platform.AWS, compute.Platform.AWS, computeReq)...)
			}
		}*/
	return allErrs.ToAggregate()
}

func validatePlatform(ctx context.Context, meta *Metadata, fldPath *field.Path, platform *awstypes.Platform, networking *types.Networking, publish types.PublishingStrategy) field.ErrorList {
	allErrs := field.ErrorList{}

	//allErrs = append(allErrs, validateServiceEndpoints(fldPath.Child("serviceEndpoints"), platform.Region, platform.ServiceEndpoints)...)

	// Fail fast when service endpoints are invalid to avoid long timeouts.
	if len(allErrs) > 0 {
		return allErrs
	}

	if len(platform.Subnets) > 0 {
		allErrs = append(allErrs, validateSubnets(ctx, meta, fldPath.Child("subnets"), platform.Subnets, networking, publish)...)
	}
	//if platform.DefaultMachinePlatform != nil {
	//	allErrs = append(allErrs, validateMachinePool(ctx, meta, fldPath.Child("defaultMachinePlatform"), platform, platform.DefaultMachinePlatform, controlPlaneReq)...)
	//}
	return allErrs
}

func validateAMI(ctx context.Context, config *types.InstallConfig) field.ErrorList {
	// accept AMI from the rhcos stream metadata
	//if rhcos.AMIRegions(config.ControlPlane.Architecture).Has(config.Platform.AWS.Region) {
	//	return nil
	//}

	// accept AMI specified at the platform level
	if config.Platform.AWS.AMIID != "" {
		return nil
	}

	// accept AMI specified for the default machine platform
	/*
		if config.Platform.AWS.DefaultMachinePlatform != nil {
			if config.Platform.AWS.DefaultMachinePlatform.AMIID != "" {
				return nil
			}
		}
	*/

	// accept AMI that can be copied from us-east-1 if the region is in the standard AWS partition
	if partition, partitionFound := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), config.Platform.AWS.Region); partitionFound {
		if partition.ID() == endpoints.AwsPartitionID {
			return nil
		}
	}

	// fail validation since we do not have an AMI to use
	return field.ErrorList{field.Required(field.NewPath("platform", "aws", "amiID"), "AMI must be provided")}
}

func validateSubnets(ctx context.Context, meta *Metadata, fldPath *field.Path, subnets []string, networking *types.Networking, publish types.PublishingStrategy) field.ErrorList {
	allErrs := field.ErrorList{}
	privateSubnets, err := meta.PrivateSubnets(ctx)
	if err != nil {
		return append(allErrs, field.Invalid(fldPath, subnets, err.Error()))
	}
	privateSubnetsIdx := map[string]int{}
	for idx, id := range subnets {
		if _, ok := privateSubnets[id]; ok {
			privateSubnetsIdx[id] = idx
		}
	}
	if len(privateSubnets) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, subnets, "No private subnets found"))
	}

	publicSubnets, err := meta.PublicSubnets(ctx)
	if err != nil {
		return append(allErrs, field.Invalid(fldPath, subnets, err.Error()))
	}
	publicSubnetsIdx := map[string]int{}
	for idx, id := range subnets {
		if _, ok := publicSubnets[id]; ok {
			publicSubnetsIdx[id] = idx
		}
	}

	privateZones := sets.NewString()
	publicZones := sets.NewString()
	for _, subnet := range privateSubnets {
		privateZones.Insert(subnet.Zone)
	}
	for _, subnet := range publicSubnets {
		publicZones.Insert(subnet.Zone)
	}
	if publish == types.ExternalPublishingStrategy && !publicZones.IsSuperset(privateZones) {
		errMsg := fmt.Sprintf("No public subnet provided for zones %s", privateZones.Difference(publicZones).List())
		allErrs = append(allErrs, field.Invalid(fldPath, subnets, errMsg))
	}

	return allErrs
}

func validateServiceEndpoints(fldPath *field.Path, region string, services []awstypes.ServiceEndpoint) field.ErrorList {
	allErrs := field.ErrorList{}
	ec2Endpoint := ""
	for id, service := range services {
		err := validateEndpointAccessibility(service.URL)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Index(id).Child("url"), service.URL, err.Error()))
			continue
		}
		if service.Name == ec2.ServiceName {
			ec2Endpoint = service.URL
		}
	}

	if partition, partitionFound := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); partitionFound {
		if _, ok := partition.Regions()[region]; !ok && ec2Endpoint == "" {
			err := validateRegion(region)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("region"), region, err.Error()))
			}
		}
		return allErrs
	}

	resolver := newAWSResolver(region, services)
	var errs []error
	for _, service := range requiredServices {
		_, err := resolver.EndpointFor(service, region, endpoints.StrictMatchingOption)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to find endpoint for service %q", service))
		}
	}
	if err := utilerrors.NewAggregate(errs); err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath, services, err.Error()))
	}
	return allErrs
}

func validateRegion(region string) error {
	ses, err := GetSessionWithOptions(func(sess *session.Options) {
		sess.Config.Region = aws.String(region)
	})
	if err != nil {
		return err
	}
	ec2Session := ec2.New(ses)
	return validateEndpointAccessibility(ec2Session.Endpoint)
}

func validateEndpointAccessibility(endpointURL string) error {
	// For each provided service endpoint, verify we can resolve and connect with net.Dial.
	// Ignore e2e.local from unit tests.
	if endpointURL == "e2e.local" {
		return nil
	}
	_, err := url.Parse(endpointURL)
	if err != nil {
		return err
	}
	_, err = http.Head(endpointURL)
	return err
}

var requiredServices = []string{
	"ec2",
	"elasticloadbalancing",
	"iam",
	"route53",
	"s3",
	"sts",
	"tagging",
}

// ValidateForProvisioning validates if the install config is valid for provisioning the cluster.
func ValidateForProvisioning(client API, ic *types.InstallConfig, metadata *Metadata) error {
	if ic.Publish == types.InternalPublishingStrategy && ic.AWS.HostedZone == "" {
		return nil
	}

	var zoneName string
	var zonePath *field.Path
	var zone *route53.HostedZone

	errors := field.ErrorList{}
	allErrs := field.ErrorList{}

	if ic.AWS.HostedZone != "" {
		zoneName = ic.AWS.HostedZone
		zonePath = field.NewPath("aws", "hostedZone")
		zoneOutput, err := client.GetHostedZone(zoneName)
		if err != nil {
			return field.ErrorList{
				field.Invalid(zonePath, zoneName, "cannot find hosted zone"),
			}.ToAggregate()
		}

		if errors = validateHostedZone(zoneOutput, zonePath, zoneName, metadata); len(errors) > 0 {
			allErrs = append(allErrs, errors...)
		}

		zone = zoneOutput.HostedZone
	} else {
		zoneName = ic.BaseDomain
		zonePath = field.NewPath("baseDomain")
		baseDomainOutput, err := client.GetBaseDomain(zoneName)
		if err != nil {
			return field.ErrorList{
				field.Invalid(zonePath, zoneName, "cannot find base domain"),
			}.ToAggregate()
		}

		zone = baseDomainOutput
	}

	if errors = client.ValidateZoneRecords(zone, zoneName, zonePath, ic); len(errors) > 0 {
		allErrs = append(allErrs, errors...)
	}

	return allErrs.ToAggregate()
}

func validateHostedZone(hostedZoneOutput *route53.GetHostedZoneOutput, hostedZonePath *field.Path, hostedZoneName string, metadata *Metadata) field.ErrorList {
	allErrs := field.ErrorList{}

	// validate that the hosted zone is associated with the VPC containing the existing subnets for the cluster
	vpcID, err := metadata.VPC(context.TODO())
	if err == nil {
		if !isHostedZoneAssociatedWithVPC(hostedZoneOutput, vpcID) {
			allErrs = append(allErrs, field.Invalid(hostedZonePath, hostedZoneName, "hosted zone is not associated with the VPC"))
		}
	} else {
		allErrs = append(allErrs, field.Invalid(hostedZonePath, hostedZoneName, "no VPC found"))
	}

	return allErrs
}

func isHostedZoneAssociatedWithVPC(hostedZone *route53.GetHostedZoneOutput, vpcID string) bool {
	if vpcID == "" {
		return false
	}
	for _, vpc := range hostedZone.VPCs {
		if aws.StringValue(vpc.VPCId) == vpcID {
			return true
		}
	}
	return false
}
