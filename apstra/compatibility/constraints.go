package compatibility

import (
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/chrismarget-j/version-constraints"
)

var (
	BPDefaultRoutingZoneAddressingOK            = versionconstraints.New(apiversions.GeApstra610)
	BlueprintIPv6ApplicationsOK                 = versionconstraints.New(apiversions.LtApstra610)
	DatacenterCTPrimitiveVNSingleOverrideVLANOK = versionconstraints.New(apiversions.GeApstra620)
	DatacenterPolicyAddressFamilyForbidden      = versionconstraints.New(apiversions.LtApstra620)
	DatacenterPolicyAddressFamilyNotRequired    = versionconstraints.New(apiversions.LtApstra620)
	DatacenterPolicyAddressFamilyOK             = versionconstraints.New(apiversions.GeApstra620)
	DatacenterPolicyAddressFamilyRequired       = versionconstraints.New(apiversions.GeApstra620)
	DCIVNUpdatedPatchSemantics                  = versionconstraints.New(apiversions.GeApstra620)
	PolicyNodesUseTagAttribute                  = versionconstraints.New(apiversions.LtApstra620)
	SwitchingZoneOK                             = versionconstraints.New(apiversions.GeApstra620)
	VnAPITagsOk                                 = versionconstraints.New(apiversions.GeApstra620)
	VnDHCPUnsafeWithoutWithoutBindings          = versionconstraints.New(apiversions.LtApstra620)
	VnEncapsulateInnerVLANOK                    = versionconstraints.New(apiversions.GeApstra620)
)
