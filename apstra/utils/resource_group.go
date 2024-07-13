package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
)

func AllResourceGroupNameStrings() []string {
	argn := apstra.AllResourceGroupNames()
	var result []string
	for _, rgn := range argn {
		if rgn == apstra.ResourceGroupNameNone {
			continue
		}
		result = append(result, StringersToFriendlyString(rgn))
	}

	sort.Strings(result)
	return result
}
