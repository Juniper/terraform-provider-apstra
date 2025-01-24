package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
)

func ConfigletSupportsPlatforms(configletdata *apstra.ConfigletData, platforms []enum.ConfigletStyle) bool {
	supportedPlatforms := make(map[enum.ConfigletStyle]struct{})
	for _, generator := range configletdata.Generators {
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

func ConfigletSectionNamesByStyle(style enum.ConfigletStyle) []string {
	var result []string
	for _, v := range apstra.ValidConfigletSections(style) {
		result = append(result, StringersToFriendlyString(v, style))
	}
	return result
}

func ConfigletValidSectionsMap() map[string][]string {
	result := make(map[string][]string)
	for _, i := range enum.ConfigletStyles.Members() {
		result[i.String()] = ConfigletSectionNamesByStyle(i)
	}
	return result
}

func ValidSectionsAsTable() string {
	m := ConfigletValidSectionsMap()

	// collect map keys into a sorted slice so that the table renders in consistent order
	keys := make([]string, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("  | **Config Style**  | **Valid Sections** |\n")
	sb.WriteString("  |---|---|\n")

	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("  |%s|%s|\n", k, strings.Join(m[k], ", ")))
	}

	return sb.String()
}

func AllPlatformOSNames() []string {
	configletStyles := enum.ConfigletStyles.Members()
	result := make([]string, len(configletStyles))
	for i, configletStyle := range configletStyles {
		result[i] = configletStyle.String()
	}
	sort.Strings(result)
	return result
}
