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

func AllConfigletSectionNames() []string {
	platformToFriendlySectionName := ConfigletValidSectionsMap() // map of os -> friendly section names
	friendlySectionNames := make(map[string]struct{})            // unique-ify the friendly section names here

	for _, v := range platformToFriendlySectionName {
		for _, friendlyName := range v {
			friendlySectionNames[friendlyName] = struct{}{} // map keys are friendly names across all platforms
		}
	}

	// turn the map into a list
	var i int
	result := make([]string, len(friendlySectionNames))
	for k := range friendlySectionNames {
		result[i] = k
		i++
	}

	return result
}

func ConfigletSectionNamesByOS(os apstra.PlatformOS) []string {
	var r []string
	for _, v := range os.ValidSections() {
		r = append(r, StringersToFriendlyString(v, os))
	}
	return r
}

func ConfigletValidSectionsMap() map[string][]string {
	var m = make(map[string][]string)
	for _, i := range apstra.AllPlatformOSes() {
		m[i.String()] = ConfigletSectionNamesByOS(i)
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
