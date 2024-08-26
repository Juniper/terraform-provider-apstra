package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
)

func PeerToTypes() []string {
	result := []string{
		StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToLoopback),
		StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToInterfaceOrIpEndpoint),
		StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToInterfaceOrSharedIpEndpoint),
	}
	sort.Strings(result)
	return result
}
