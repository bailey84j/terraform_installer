package aws

import (
	"fmt"
	"sort"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/bailey84j/terraform_installer/pkg/types/aws"
	"github.com/bailey84j/terraform_installer/pkg/version"
)

// Platform collects AWS-specific configuration.
func Platform() (*aws.Platform, error) {
	logrus.Debugf("Trace Me - In aws.Platform()")
	architecture := version.DefaultArch()
	logrus.Debugf("Trace Me - Arch - %v", architecture)
	regions := knownPublicRegions(architecture)
	longRegions := make([]string, 0, len(regions))
	shortRegions := make([]string, 0, len(regions))
	for id, location := range regions {
		longRegions = append(longRegions, fmt.Sprintf("%s (%s)", id, location))
		shortRegions = append(shortRegions, id)
	}
	logrus.Debugf("Trace Me - In aws.Platform() - D1")
	var regionTransform survey.Transformer = func(ans interface{}) interface{} {
		switch v := ans.(type) {
		case core.OptionAnswer:
			return core.OptionAnswer{Value: strings.SplitN(v.Value, " ", 2)[0], Index: v.Index}
		case string:
			return strings.SplitN(v, " ", 2)[0]
		}
		return ""
	}
	logrus.Debugf("Trace Me - In aws.Platform() - D2")

	defaultRegion := "us-east-1"
	//if !IsKnownPublicRegion(defaultRegion, architecture) {
	//	panic(fmt.Sprintf("installer bug: invalid default AWS region %q", defaultRegion))
	//}
	/*
		ssn, err := GetSession()
		if err != nil {
			return nil, err
		}

			defaultRegionPointer := ssn.Config.Region

				if defaultRegionPointer != nil && *defaultRegionPointer != "" {
					if IsKnownPublicRegion(*defaultRegionPointer, architecture) {
						defaultRegion = *defaultRegionPointer
					} else {
						logrus.Warnf("Unrecognized AWS region %q, defaulting to %s", *defaultRegionPointer, defaultRegion)
					}
				}
	*/
	sort.Strings(longRegions)
	sort.Strings(shortRegions)

	var region string
	logrus.Debugf("Trace Me - Regions %v", regions)
	logrus.Debugf("Trace Me - LongRegions %v", longRegions)
	logrus.Debugf("Trace Me - In aws.Platform() - D3")
	err := survey.Ask([]*survey.Question{
		{
			Prompt: &survey.Select{
				Message: "Region",
				Help:    "The AWS region to be used for installation.",
				Default: fmt.Sprintf("%s (%s)", defaultRegion, regions[defaultRegion]),
				Options: longRegions,
			},
			Validate: survey.ComposeValidators(survey.Required, func(ans interface{}) error {
				choice := regionTransform(ans).(core.OptionAnswer).Value
				i := sort.SearchStrings(shortRegions, choice)
				if i == len(shortRegions) || shortRegions[i] != choice {
					return errors.Errorf("invalid region %q", choice)
				}
				return nil
			}),
			Transform: regionTransform,
		},
	}, &region)
	if err != nil {
		logrus.Debugf("Trace Me - Error %s", err.Error())
		return nil, err
	}
	logrus.Debugf("Trace Me - In aws.Platform() - D4")
	logrus.Debugf("Trace Me - Region - %s", region)

	return &aws.Platform{
		Region: region,
	}, nil
}
