package utils

import "github.com/Juniper/apstra-go-sdk/apstra"

func ConfigletSupportsPlatforms(configlet *apstra.Configlet, platforms []apstra.PlatformOS) bool {
	supportedPlatforms := make(map[apstra.PlatformOS]struct{})
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

func SectionNamesByOS(os apstra.PlatformOS) []string {
	var r []string
	for _, v := range os.ValidSections() {
		r = append(r, StringersToFriendlyString(v, os))
	}
	return r
}

func ValidSectionsMap() map[string][]string {
	var m = make(map[string][]string)
	for _, i := range apstra.AllPlatformOSes() {
		m[i.String()] = SectionNamesByOS(i)
	}
	return m
}

func AllPlatformOSNames() []string {
	platforms := apstra.AllPlatformOSes()
	result := make([]string, len(platforms))
	for i := range platforms {
		result[i] = platforms[i].String()
	}
	return result
}
