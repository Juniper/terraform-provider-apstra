package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
)

func FcdModes() []string {
	result := []string{
		StringersToFriendlyString(apstra.FabricConnectivityDesignL3Clos),
		StringersToFriendlyString(apstra.FabricConnectivityDesignL3Collapsed),
	}
	sort.Strings(result)
	return result
}

func NeighborAsnTypes() []string {
	return []string{
		"static",
		"dynamic",
	}
}

func OverlayControlProtocols() []string {
	members := apstra.AllOverlayControlProtocols()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

func PeerToTypes() []string {
	result := []string{
		StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToLoopback),
		StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToInterfaceOrIpEndpoint),
		StringersToFriendlyString(apstra.CtPrimitiveBgpPeerToInterfaceOrSharedIpEndpoint),
	}
	sort.Strings(result)
	return result
}

func StorageSchemaPaths() []string {
	members := apstra.StorageSchemaPaths.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

func TemplateTypes() []string {
	members := apstra.AllTemplateTypes()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}
