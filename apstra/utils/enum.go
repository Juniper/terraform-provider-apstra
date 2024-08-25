package utils

import (
	"sort"

	"github.com/Juniper/apstra-go-sdk/apstra"
)

func AgentProfilePlatforms() []string {
	result := []string{
		StringersToFriendlyString(apstra.AgentPlatformNXOS),
		StringersToFriendlyString(apstra.AgentPlatformJunos),
		StringersToFriendlyString(apstra.AgentPlatformEOS),
	}
	sort.Strings(result)
	return result
}

func DcRoutingPolicyImportPolicy() []string {
	members := apstra.AllDcRoutingPolicyImportPolicies()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

func FcdModes() []string {
	result := []string{
		StringersToFriendlyString(apstra.FabricConnectivityDesignL3Clos),
		StringersToFriendlyString(apstra.FabricConnectivityDesignL3Collapsed),
	}
	sort.Strings(result)
	return result
}

func InterfaceNumberingIpv4Types() []string {
	members := apstra.InterfaceNumberingIpv4Types.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	sort.Strings(result)
	return result
}

func InterfaceNumberingIpv6Types() []string {
	members := apstra.InterfaceNumberingIpv6Types.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
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

func NodeDeployModes() []string {
	members := apstra.DeployModes.Members()
	result := make([]string, len(members))
	for i, member := range members {
		result[i] = StringersToFriendlyString(member)
	}
	return result
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

func PlatformOSNames() []string {
	platforms := apstra.AllPlatformOSes()
	result := make([]string, len(platforms))
	for i := range platforms {
		result[i] = StringersToFriendlyString(platforms[i])
	}
	sort.Strings(result)
	return result
}

func ResourceGroupNameStrings() []string {
	members := apstra.AllResourceGroupNames()
	var result []string
	for _, member := range members {
		if member == apstra.ResourceGroupNameNone {
			continue
		}
		result = append(result, StringersToFriendlyString(member))
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

func ValidPrefixFilterActions() []string {
	members := apstra.AllPrefixFilterActions()
	result := make([]string, len(members))
	var i int
	for _, member := range members {
		if member != apstra.PrefixFilterActionNone {
			result[i] = StringersToFriendlyString(member)
			i++
		}
	}
	result = result[:i]
	sort.Strings(result)
	return result
}
