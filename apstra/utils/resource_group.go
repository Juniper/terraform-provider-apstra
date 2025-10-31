package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
)

func AllResourceGroupNameStrings() []string {
	argn := apstra.AllResourceGroupNames()
	var result []string
	for _, rgn := range argn {
		if rgn == apstra.ResourceGroupNameNone {
			continue
		}
		result = append(result, rosetta.StringersToFriendlyString(rgn))
	}

	sort.Strings(result)
	return result
}
