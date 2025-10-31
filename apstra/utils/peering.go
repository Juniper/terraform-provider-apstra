package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/internal/rosetta"
)

func PeerToTypes() []string {
	result := []string{
		rosetta.StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToLoopback),
		rosetta.StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToInterfaceOrIpEndpoint),
		rosetta.StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToInterfaceOrSharedIpEndpoint),
	}
	sort.Strings(result)
	return result
}
