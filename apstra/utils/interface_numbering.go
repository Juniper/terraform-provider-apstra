package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
)

func AllInterfaceNumberingIpv4Types() []string {
	members := apstra.InterfaceNumberingIpv4Types.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}

	sort.Strings(result)
	return result
}

func AllInterfaceNumberingIpv6Types() []string {
	members := apstra.InterfaceNumberingIpv6Types.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}

	sort.Strings(result)
	return result
}
