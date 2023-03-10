package utils

import "bitbucket.org/apstrktr/goapstra"

func ConfigletSupportsPlatforms(configlet *goapstra.Configlet, platforms []goapstra.PlatformOS) bool {
	supportedPlatforms := make(map[goapstra.PlatformOS]struct{})
	for _, generator := range configlet.Data.Generators {
		supportedPlatforms[generator.ConfigStyle] = struct{}{}
	}

	for _, platform := range platforms {
		if _, ok := supportedPlatforms[platform]; !ok {
			return false
		}
	}
	return true
}

func AllConfigletSectionNames() []string {
	sections := goapstra.AllConfigletSections()
	result := make([]string, len(sections))
	for i := range sections {
		result[i] = sections[i].String()
	}
	return result
}

func AllPlatformOSNames() []string {
	platforms := goapstra.AllPlatformOSes()
	result := make([]string, len(platforms))
	for i := range platforms {
		result[i] = platforms[i].String()
	}
	return result
}
