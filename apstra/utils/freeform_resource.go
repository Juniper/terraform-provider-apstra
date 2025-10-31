package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
)

func AllFFResourceTypes() []string {
	members := enum.FFResourceTypes.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = rosetta.StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}
