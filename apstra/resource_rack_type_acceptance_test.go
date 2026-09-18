//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/Juniper/apstra-go-sdk/enum"
	"github.com/Juniper/apstra-go-sdk/speed"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/internal/pointer"
	designtestobjects "github.com/Juniper/terraform-provider-apstra/internal/test_utils/design_test_objects"
	"github.com/Juniper/terraform-provider-apstra/internal/test_utils/random"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceRackTypeHCL = `resource %q %q {
  name                       = %q // required attribute
  description                = %s // optional attribute
  fabric_connectivity_design = %q // required attribute
  leaf_switches              = %s // required map attribute
  access_switches            = %s // optional map attribute
  generic_systems            = %s // optional map attribute
}
`
	datasourceRackTypeHCL = `data %q %q {
  id   = %s // optional attribute
  name = %s // optional attribute
}
`
	datasourceRackTypesHCL = `data %q %q {
  depends_on = [%s.%s]
}
`
)

type resourceRackType struct {
	name                     string
	description              string
	fabricConnectivityDesign enum.FabricConnectivityDesign
	leafSwitches             map[string]resourceRackTypeLeafSwitch
	accessSwitches           map[string]resourceRackTypeAccessSwitch
	genericSystems           map[string]resourceRackTypeGenericSystem
}

func (rt resourceRackType) render(rType, rName string) string {
	leafSwitches := new(strings.Builder)
	accessSwitches := new(strings.Builder)
	genericSystems := new(strings.Builder)

	leafSwitches.WriteString("{\n")
	for k, v := range rt.leafSwitches {
		leafSwitches.WriteString(v.render(k))
	}
	leafSwitches.WriteString("  }")

	accessSwitches.WriteString("{\n")
	for k, v := range rt.accessSwitches {
		accessSwitches.WriteString(v.render(k))
	}
	if len(rt.accessSwitches) > 0 {
		accessSwitches.WriteString("  }")
	} else {
		accessSwitches.Reset()
		accessSwitches.WriteString("null")
	}

	genericSystems.WriteString("{\n")
	for k, v := range rt.genericSystems {
		genericSystems.WriteString(v.render(k))
	}
	if len(rt.genericSystems) > 0 {
		genericSystems.WriteString("  }")
	} else {
		genericSystems.Reset()
		genericSystems.WriteString("null")
	}

	resourceBlock := fmt.Sprintf(resourceRackTypeHCL, rType, rName,
		rt.name,
		stringOrNull(rt.description),
		rt.fabricConnectivityDesign,
		leafSwitches.String(),
		accessSwitches.String(),
		genericSystems.String(),
	)
	datasourceByIDBlock := fmt.Sprintf(datasourceRackTypeHCL, rType, rName+"_by_id", "resource."+rType+"."+rName+".id", "null")
	datasourceByNameBlock := fmt.Sprintf(datasourceRackTypeHCL, rType, rName+"_by_name", "null", "resource."+rType+"."+rName+".name")
	datasourceIDsBlock := fmt.Sprintf(datasourceRackTypesHCL, rType+"s", "all", rType, rName)

	return resourceBlock + "\n" + datasourceByIDBlock + "\n" + datasourceByNameBlock + "\n" + datasourceIDsBlock
}

func (rt resourceRackType) testChecks(t testing.TB, rType, rName string) []testChecks {
	rChecks := newTestChecks(rType + "." + rName)
	dByIDChecks := newTestChecks("data." + rType + "." + rName + "_by_id")
	dByNameChecks := newTestChecks("data." + rType + "." + rName + "_by_name")

	rChecks.append(t, "TestCheckResourceAttrSet", "id")
	dByIDChecks.append(t, "TestCheckResourceAttrSet", "id")
	dByNameChecks.append(t, "TestCheckResourceAttrSet", "id")
	rChecks.append(t, "TestCheckResourceAttr", "name", rt.name)
	dByIDChecks.append(t, "TestCheckResourceAttr", "name", rt.name)
	dByNameChecks.append(t, "TestCheckResourceAttr", "name", rt.name)
	rChecks.append(t, "TestCheckResourceAttr", "description", rt.description)
	dByIDChecks.append(t, "TestCheckResourceAttr", "description", rt.description)
	dByNameChecks.append(t, "TestCheckResourceAttr", "description", rt.description)
	rChecks.append(t, "TestCheckResourceAttr", "fabric_connectivity_design", rt.fabricConnectivityDesign.String())
	dByIDChecks.append(t, "TestCheckResourceAttr", "fabric_connectivity_design", rt.fabricConnectivityDesign.String())
	dByNameChecks.append(t, "TestCheckResourceAttr", "fabric_connectivity_design", rt.fabricConnectivityDesign.String())

	rChecks.append(t, "TestCheckResourceAttr", "leaf_switches.%", strconv.Itoa(len(rt.leafSwitches)))
	for k, v := range rt.leafSwitches {
		rChecks = v.testChecks(t, "leaf_switches."+k, rChecks, true)
		dByIDChecks = v.testChecks(t, "leaf_switches."+k, dByIDChecks, false)
		dByNameChecks = v.testChecks(t, "leaf_switches."+k, dByNameChecks, false)
	}

	rChecks.append(t, "TestCheckResourceAttr", "access_switches.%", strconv.Itoa(len(rt.accessSwitches)))
	for k, v := range rt.accessSwitches {
		rChecks = v.testChecks(t, "access_switches."+k, rChecks, true)
		dByIDChecks = v.testChecks(t, "access_switches."+k, dByIDChecks, false)
		dByNameChecks = v.testChecks(t, "access_switches."+k, dByNameChecks, false)
	}

	rChecks.append(t, "TestCheckResourceAttr", "generic_systems.%", strconv.Itoa(len(rt.genericSystems)))
	for k, v := range rt.genericSystems {
		rChecks = v.testChecks(t, "generic_systems."+k, rChecks, true)
		dByIDChecks = v.testChecks(t, "generic_systems."+k, dByIDChecks, false)
		dByNameChecks = v.testChecks(t, "generic_systems."+k, dByNameChecks, false)
	}

	return []testChecks{rChecks, dByIDChecks, dByNameChecks}
}

const resourceRackTypeLeafSwitchHCL = `    %s = {
      logical_device_id   = %q // required attribute
      mlag_info           = %s // optional attribute
      redundancy_protocol = %s // optional attribute
      spine_link_count    = %s // optional attribute
      spine_link_speed    = %s // optional attribute
      tag_ids             = %s // optional attribute
    },
`

type resourceRackTypeLeafSwitch struct {
	logicalDeviceID    string
	mlagInfo           *resourceRackTypeLeafSwitchMLAGInfo
	redundancyProtocol *enum.LeafRedundancyProtocol
	spineLinkCount     *int
	spineLinkSpeed     speed.Speed
	tagIDs             []string
}

func (ls resourceRackTypeLeafSwitch) render(key string) string {
	mlagInfo := "null"
	if ls.mlagInfo != nil {
		mlagInfo = ls.mlagInfo.render()
	}

	return fmt.Sprintf(resourceRackTypeLeafSwitchHCL, key,
		ls.logicalDeviceID,
		mlagInfo,
		stringerOrNull(ls.redundancyProtocol),
		intPtrOrNull(ls.spineLinkCount),
		stringOrNull(ls.spineLinkSpeed),
		stringSliceOrNull(ls.tagIDs),
	)
}

func (ls resourceRackTypeLeafSwitch) testChecks(t testing.TB, name string, checks testChecks, nestedIDsAvailable bool) testChecks {
	if nestedIDsAvailable {
		checks.append(t, "TestCheckResourceAttr", name+".logical_device_id", ls.logicalDeviceID)
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".logical_device_id")
	}
	if ls.mlagInfo != nil {
		checks = ls.mlagInfo.testChecks(t, name+".mlag_info", checks)
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".mlag_info")
	}
	if ls.redundancyProtocol != nil {
		checks.append(t, "TestCheckResourceAttr", name+".redundancy_protocol", ls.redundancyProtocol.String())
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".redundancy_protocol")
	}
	if ls.spineLinkCount != nil {
		checks.append(t, "TestCheckResourceAttr", name+".spine_link_count", strconv.Itoa(*ls.spineLinkCount))
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".spine_link_count")
	}
	if ls.spineLinkSpeed != "" {
		checks.append(t, "TestCheckResourceAttr", name+".spine_link_speed", string(ls.spineLinkSpeed))
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".spine_link_speed")
	}
	if nestedIDsAvailable {
		checks.append(t, "TestCheckResourceAttr", name+".tag_ids.#", strconv.Itoa(len(ls.tagIDs)))
		for _, tagID := range ls.tagIDs {
			checks.append(t, "TestCheckTypeSetElemAttr", name+".tag_ids.*", tagID)
		}
	} else {
		checks.append(t, "TestCheckResourceAttr", name+".tags.#", strconv.Itoa(len(ls.tagIDs)))
	}
	return checks
}

const resourceRackTypeLeafSwitchMLAGInfoHCL = `{
        mlag_keepalive_vlan          = %d // required attribute
        peer_link_count              = %d // required attribute
        peer_link_speed              = %q // required attribute
        peer_link_port_channel_id    = %d // required attribute
        l3_peer_link_count           = %s // optional attribute
        l3_peer_link_speed           = %s // optional attribute
        l3_peer_link_port_channel_id = %s // optional attribute
      }`

type resourceRackTypeLeafSwitchMLAGInfo struct {
	mlagKeepaliveVLAN       int
	peerLinkCount           int
	peerLinkSpeed           speed.Speed
	peerLinkPortChannelID   int
	l3PeerLinkCount         *int
	l3PeerLinkSpeed         speed.Speed
	l3PeerLinkPortChannelID *int
}

func (mlag resourceRackTypeLeafSwitchMLAGInfo) render() string {
	return fmt.Sprintf(resourceRackTypeLeafSwitchMLAGInfoHCL,
		mlag.mlagKeepaliveVLAN,
		mlag.peerLinkCount,
		mlag.peerLinkSpeed,
		mlag.peerLinkPortChannelID,
		intPtrOrNull(mlag.l3PeerLinkCount),
		stringOrNull(mlag.l3PeerLinkSpeed),
		intPtrOrNull(mlag.l3PeerLinkPortChannelID),
	)
}

func (mlag resourceRackTypeLeafSwitchMLAGInfo) testChecks(t testing.TB, name string, checks testChecks) testChecks {
	checks.append(t, "TestCheckResourceAttr", name+".mlag_keepalive_vlan", strconv.Itoa(mlag.mlagKeepaliveVLAN))
	checks.append(t, "TestCheckResourceAttr", name+".peer_link_count", strconv.Itoa(mlag.peerLinkCount))
	checks.append(t, "TestCheckResourceAttr", name+".peer_link_speed", string(mlag.peerLinkSpeed))
	checks.append(t, "TestCheckResourceAttr", name+".peer_link_port_channel_id", strconv.Itoa(mlag.peerLinkPortChannelID))
	if mlag.l3PeerLinkCount != nil {
		checks.append(t, "TestCheckResourceAttr", name+".l3_peer_link_count", strconv.Itoa(*mlag.l3PeerLinkCount))
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".l3_peer_link_count")
	}
	if mlag.l3PeerLinkSpeed != "" {
		checks.append(t, "TestCheckResourceAttr", name+".l3_peer_link_speed", string(mlag.l3PeerLinkSpeed))
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".l3_peer_link_speed")
	}
	if mlag.l3PeerLinkPortChannelID != nil {
		checks.append(t, "TestCheckResourceAttr", name+".l3_peer_link_port_channel_id", strconv.Itoa(*mlag.l3PeerLinkPortChannelID))
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".l3_peer_link_port_channel_id")
	}
	return checks
}

const resourceRackTypeAccessSwitchHCL = `    %s = {
      count             = %d // required attribute
      logical_device_id = %q // required attribute
      esi_lag_info      = %s // optional attribute
      links             = %s // required map attribute
      tag_ids           = %s // optional attribute
    },
`

type resourceRackTypeAccessSwitch struct {
	count           int
	logicalDeviceID string
	esiLagInfo      *resourceRackTypeAccessSwitchESILAGInfo
	links           map[string]resourceRackTypeLink
	tagIDs          []string
}

func (as resourceRackTypeAccessSwitch) render(key string) string {
	esiLagInfo := "null"
	if as.esiLagInfo != nil {
		esiLagInfo = as.esiLagInfo.render()
	}

	links := new(strings.Builder)
	links.WriteString("{\n")
	for k, link := range as.links {
		links.WriteString(link.render(k))
	}
	links.WriteString("    }")

	return fmt.Sprintf(resourceRackTypeAccessSwitchHCL, key,
		as.count,
		as.logicalDeviceID,
		esiLagInfo,
		links.String(),
		stringSliceOrNull(as.tagIDs),
	)
}

func (as resourceRackTypeAccessSwitch) testChecks(t testing.TB, name string, checks testChecks, nestedIDsAvailable bool) testChecks {
	checks.append(t, "TestCheckResourceAttr", name+".count", strconv.Itoa(as.count))
	if nestedIDsAvailable {
		checks.append(t, "TestCheckResourceAttr", name+".logical_device_id", as.logicalDeviceID)
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".logical_device_id")
	}
	if as.esiLagInfo != nil {
		checks = as.esiLagInfo.testChecks(t, name+".esi_lag_info", checks)
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".esi_lag_info")
	}
	checks.append(t, "TestCheckResourceAttr", name+".links.%", strconv.Itoa(len(as.links)))
	for k, v := range as.links {
		checks = v.testChecks(t, name+".links."+k, checks, nestedIDsAvailable)
	}
	if nestedIDsAvailable {
		checks.append(t, "TestCheckResourceAttr", name+".tag_ids.#", strconv.Itoa(len(as.tagIDs)))
		for _, tagID := range as.tagIDs {
			checks.append(t, "TestCheckTypeSetElemAttr", name+".tag_ids.*", tagID)
		}
	} else {
		checks.append(t, "TestCheckResourceAttr", name+".tags.#", strconv.Itoa(len(as.tagIDs)))
	}
	return checks
}

const resourceRackTypeAccessSwitchESILAGInfoHCL = `{
        l3_peer_link_count = %d // required attribute
        l3_peer_link_speed = %q // required attribute
      }`

type resourceRackTypeAccessSwitchESILAGInfo struct {
	l3PeerLinkCount int
	l3PeerLinkSpeed speed.Speed
}

func (eli resourceRackTypeAccessSwitchESILAGInfo) render() string {
	return fmt.Sprintf(resourceRackTypeAccessSwitchESILAGInfoHCL, eli.l3PeerLinkCount, eli.l3PeerLinkSpeed)
}

func (eli resourceRackTypeAccessSwitchESILAGInfo) testChecks(t testing.TB, name string, checks testChecks) testChecks {
	checks.append(t, "TestCheckResourceAttr", name+".l3_peer_link_count", strconv.Itoa(eli.l3PeerLinkCount))
	checks.append(t, "TestCheckResourceAttr", name+".l3_peer_link_speed", string(eli.l3PeerLinkSpeed))
	return checks
}

const resourceRackTypeGenericSystemHCL = `%s = {
      count               = %d // required attribute
      logical_device_id   = %q // required attribute
      port_channel_id_min = %s // optional attribute
      port_channel_id_max = %s // optional attribute
      links               = %s // required map attribute
      tag_ids             = %s // optional attribute
    },
`

type resourceRackTypeGenericSystem struct {
	count            int
	logicalDeviceID  string
	portChannelIDMin *int
	portChannelIDMax *int
	links            map[string]resourceRackTypeLink
	tagIDs           []string
}

func (gs resourceRackTypeGenericSystem) render(key string) string {
	links := new(strings.Builder)
	links.WriteString("{\n")
	for k, link := range gs.links {
		links.WriteString(link.render(k))
	}
	links.WriteString("    }")

	return fmt.Sprintf(resourceRackTypeGenericSystemHCL, key,
		gs.count,
		gs.logicalDeviceID,
		intPtrOrNull(gs.portChannelIDMin),
		intPtrOrNull(gs.portChannelIDMax),
		links,
		stringSliceOrNull(gs.tagIDs),
	)
}

func (gs resourceRackTypeGenericSystem) testChecks(t testing.TB, name string, checks testChecks, nestedIDsAvailable bool) testChecks {
	checks.append(t, "TestCheckResourceAttr", name+".count", strconv.Itoa(gs.count))
	if nestedIDsAvailable {
		checks.append(t, "TestCheckResourceAttr", name+".logical_device_id", gs.logicalDeviceID)
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".logical_device_id")
	}
	if gs.portChannelIDMin != nil {
		checks.append(t, "TestCheckResourceAttr", name+".port_channel_id_min", strconv.Itoa(*gs.portChannelIDMin))
	} else {
		checks.append(t, "TestCheckResourceAttr", name+".port_channel_id_min", "0")
	}
	if gs.portChannelIDMax != nil {
		checks.append(t, "TestCheckResourceAttr", name+".port_channel_id_max", strconv.Itoa(*gs.portChannelIDMax))
	} else {
		checks.append(t, "TestCheckResourceAttr", name+".port_channel_id_max", "0")
	}
	checks.append(t, "TestCheckResourceAttr", name+".links.%", strconv.Itoa(len(gs.links)))
	for k, v := range gs.links {
		checks = v.testChecks(t, name+".links."+k, checks, nestedIDsAvailable)
	}
	if nestedIDsAvailable {
		checks.append(t, "TestCheckResourceAttr", name+".tag_ids.#", strconv.Itoa(len(gs.tagIDs)))
		for _, tagID := range gs.tagIDs {
			checks.append(t, "TestCheckTypeSetElemAttr", name+".tag_ids.*", tagID)
		}
	} else {
		checks.append(t, "TestCheckResourceAttr", name+".tags.#", strconv.Itoa(len(gs.tagIDs)))
	}
	return checks
}

const resourceRackTypeLinkHCL = `        %s = {
          target_switch_name = %q // required attribute
          lag_mode           = %s // optional attribute
          links_per_switch   = %s // optional attribute
          speed              = %q // required attribute
          switch_peer        = %s // optional attribute
          tag_ids            = %s // optional attribute
        },
`

type resourceRackTypeLink struct {
	targetSwitchName string
	lagMode          *enum.LAGMode
	linksPerSwitch   *int
	speed            speed.Speed
	switchPeer       enum.LinkSwitchPeer
	tagIDs           []string
}

func (l resourceRackTypeLink) render(key string) string {
	return fmt.Sprintf(resourceRackTypeLinkHCL, key,
		l.targetSwitchName,
		stringerOrNull(l.lagMode),
		intPtrOrNull(l.linksPerSwitch),
		l.speed,
		stringOrNull(l.switchPeer.String()),
		stringSliceOrNull(l.tagIDs),
	)
}

func (l resourceRackTypeLink) testChecks(t testing.TB, name string, checks testChecks, nestedIDsAvailable bool) testChecks {
	checks.append(t, "TestCheckResourceAttr", name+".target_switch_name", l.targetSwitchName)
	if l.lagMode != nil {
		checks.append(t, "TestCheckResourceAttr", name+".lag_mode", l.lagMode.String())
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".lag_mode")
	}
	if l.linksPerSwitch != nil {
		checks.append(t, "TestCheckResourceAttr", name+".links_per_switch", strconv.Itoa(*l.linksPerSwitch))
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".links_per_switch")
	}
	checks.append(t, "TestCheckResourceAttr", name+".speed", string(l.speed))
	if l.switchPeer.String() != "" {
		checks.append(t, "TestCheckResourceAttr", name+".switch_peer", l.switchPeer.String())
	} else {
		checks.append(t, "TestCheckNoResourceAttr", name+".switch_peer")
	}
	if nestedIDsAvailable {
		checks.append(t, "TestCheckResourceAttr", name+".tag_ids.#", strconv.Itoa(len(l.tagIDs)))
		for _, tagID := range l.tagIDs {
			checks.append(t, "TestCheckTypeSetElemAttr", name+".tag_ids.*", tagID)
		}
	} else {
		checks.append(t, "TestCheckResourceAttr", name+".tags.#", strconv.Itoa(len(l.tagIDs)))
	}
	return checks
}

func TestACCResourceRackType(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)

	ldCount := 3
	ldIDs := make([]string, ldCount)
	for i := 0; i < ldCount; i++ {
		ldIDs[i] = designtestobjects.LogicalDeviceA(ctx, t, client)
	}

	tagCount := 10
	tagIDs := make([]string, tagCount)
	for i := 0; i < tagCount; i++ {
		tagIDs[i] = designtestobjects.RandomTag(ctx, t, client)
	}

	type testStep struct {
		config resourceRackType
	}

	type testCase struct {
		steps              []testStep
		versionConstraints version.Constraints
	}

	// randIntPtrOrNil returns a pointer to a random integer between min and max (inclusive) or
	// nil. The decision to return nil is based on a persistent random choice keyed by the
	// provided string. Any caller using a given key will get a result with the same nil/non-nil
	// decision, but the actual integer value will be different for each call.
	randIntPtrOrNil := func(key string, min, max int) *int {
		if random.PersistentIntn(key, 1) == 0 {
			return nil
		}
		return pointer.To(random.PersistentIntn(key, max-min+1) + min)
	}

	testCases := map[string]testCase{
		"l3_clos_minimal_with_2_steps": {
			steps: []testStep{
				{
					config: resourceRackType{
						name:                     acctest.RandString(6),
						description:              acctest.RandString(10),
						fabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
						leafSwitches: map[string]resourceRackTypeLeafSwitch{
							acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
								logicalDeviceID: random.OneOf(ldIDs...),
								spineLinkCount:  pointer.To(rand.Intn(5) + 1), // 1-5
								spineLinkSpeed:  speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
							},
						},
					},
				},
				{
					config: resourceRackType{
						name:                     acctest.RandString(6),
						description:              acctest.RandString(10),
						fabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
						leafSwitches: map[string]resourceRackTypeLeafSwitch{
							acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
								logicalDeviceID: random.OneOf(ldIDs...),
								spineLinkCount:  pointer.To(rand.Intn(5) + 1), // 1-5
								spineLinkSpeed:  speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
							},
						},
					},
				},
			},
		},
		"l3clos_one_of_everything": {
			steps: []testStep{
				{
					config: resourceRackType{
						name:                     acctest.RandString(6),
						description:              acctest.RandString(10),
						fabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
						leafSwitches: map[string]resourceRackTypeLeafSwitch{
							random.PersistentString("l3clos_one_of_everything_l1", 10, acctest.CharSetAlpha): {
								logicalDeviceID: random.OneOf(ldIDs...),
								spineLinkCount:  pointer.To(rand.Intn(3) + 1), // 1-3
								spineLinkSpeed:  speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
								tagIDs:          random.SomeOf(tagIDs, 2, 6),
							},
							random.PersistentString("l3clos_one_of_everything_l2", 10, acctest.CharSetAlpha): {
								logicalDeviceID:    random.OneOf(ldIDs...),
								redundancyProtocol: &enum.LeafRedundancyProtocolESI,
								spineLinkCount:     pointer.To(rand.Intn(3) + 1), // 1-3
								spineLinkSpeed:     speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
								tagIDs:             random.SomeOf(tagIDs, 2, 6),
							},
						},
						accessSwitches: map[string]resourceRackTypeAccessSwitch{
							random.PersistentString("l3clos_one_of_everything_a1", 10, acctest.CharSetAlpha): {
								count:           rand.Intn(3) + 1, // 1-3
								logicalDeviceID: random.OneOf(ldIDs...),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_l1", 10, acctest.CharSetAlpha),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
							random.PersistentString("l3clos_one_of_everything_a2", 10, acctest.CharSetAlpha): {
								count:           rand.Intn(3) + 1, // 1-3
								logicalDeviceID: random.OneOf(ldIDs...),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_l2", 10, acctest.CharSetAlpha),
										switchPeer:       random.OneOf(enum.LinkSwitchPeers.Members()...),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
							random.PersistentString("l3clos_one_of_everything_a3", 10, acctest.CharSetAlpha): {
								count:           rand.Intn(3) + 1, // 1-3
								logicalDeviceID: random.OneOf(ldIDs...),
								esiLagInfo: &resourceRackTypeAccessSwitchESILAGInfo{
									l3PeerLinkCount: rand.Intn(3) + 1, // 1-3
									l3PeerLinkSpeed: speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
								},
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_l2", 10, acctest.CharSetAlpha),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
						},
						genericSystems: map[string]resourceRackTypeGenericSystem{
							acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
								count:            rand.Intn(3) + 1, // 1-3
								logicalDeviceID:  random.OneOf(ldIDs...),
								portChannelIDMin: randIntPtrOrNil("l3clos_one_of_everything_gs1_port_channel", 11, 15),
								portChannelIDMax: randIntPtrOrNil("l3clos_one_of_everything_gs1_port_channel", 16, 20),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_l1", 10, acctest.CharSetAlpha),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
							acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
								count:            rand.Intn(3) + 1, // 1-3
								logicalDeviceID:  random.OneOf(ldIDs...),
								portChannelIDMin: randIntPtrOrNil("l3clos_one_of_everything_gs2_port_channel", 21, 25),
								portChannelIDMax: randIntPtrOrNil("l3clos_one_of_everything_gs2_port_channel", 26, 30),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_l2", 10, acctest.CharSetAlpha),
										switchPeer:       random.OneOf(enum.LinkSwitchPeers.Members()...),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
							acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
								count:            rand.Intn(3) + 1, // 1-3
								logicalDeviceID:  random.OneOf(ldIDs...),
								portChannelIDMin: randIntPtrOrNil("l3clos_one_of_everything_gs3_port_channel", 31, 35),
								portChannelIDMax: randIntPtrOrNil("l3clos_one_of_everything_gs3_port_channel", 36, 40),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_a1", 10, acctest.CharSetAlpha),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
							acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
								count:            rand.Intn(3) + 1, // 1-3
								logicalDeviceID:  random.OneOf(ldIDs...),
								portChannelIDMin: randIntPtrOrNil("l3clos_one_of_everything_gs3_port_channel", 41, 45),
								portChannelIDMax: randIntPtrOrNil("l3clos_one_of_everything_gs3_port_channel", 46, 50),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_a2", 10, acctest.CharSetAlpha),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
							acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
								count:            rand.Intn(3) + 1, // 1-3
								logicalDeviceID:  random.OneOf(ldIDs...),
								portChannelIDMin: randIntPtrOrNil("l3clos_one_of_everything_gs3_port_channel", 51, 55),
								portChannelIDMax: randIntPtrOrNil("l3clos_one_of_everything_gs3_port_channel", 56, 60),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_a3", 10, acctest.CharSetAlpha),
										switchPeer:       random.OneOf(enum.LinkSwitchPeers.Members()...),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
							acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
								count:            rand.Intn(3) + 1, // 1-3
								logicalDeviceID:  random.OneOf(ldIDs...),
								portChannelIDMin: randIntPtrOrNil("l3clos_one_of_everything_gs3_port_channel", 61, 65),
								portChannelIDMax: randIntPtrOrNil("l3clos_one_of_everything_gs3_port_channel", 66, 70),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_a3", 10, acctest.CharSetAlpha),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
						},
					},
				},
				{
					config: resourceRackType{
						name:                     acctest.RandString(6),
						description:              acctest.RandString(10),
						fabricConnectivityDesign: enum.FabricConnectivityDesignL3Clos,
						leafSwitches: map[string]resourceRackTypeLeafSwitch{
							random.PersistentString("l3clos_one_of_everything_l1", 10, acctest.CharSetAlpha): {
								logicalDeviceID: random.OneOf(ldIDs...),
								spineLinkCount:  pointer.To(rand.Intn(3) + 1), // 1-3
								spineLinkSpeed:  speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
								tagIDs:          random.SomeOf(tagIDs, 2, 6),
							},
							random.PersistentString("l3clos_one_of_everything_l2", 10, acctest.CharSetAlpha): {
								logicalDeviceID:    random.OneOf(ldIDs...),
								redundancyProtocol: &enum.LeafRedundancyProtocolMLAG,
								mlagInfo: &resourceRackTypeLeafSwitchMLAGInfo{
									mlagKeepaliveVLAN:       rand.Intn(4000) + 50,
									peerLinkCount:           rand.Intn(3) + 1,
									peerLinkSpeed:           speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
									peerLinkPortChannelID:   rand.Intn(60) + 4,
									l3PeerLinkCount:         pointer.To(rand.Intn(3) + 1),
									l3PeerLinkSpeed:         speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
									l3PeerLinkPortChannelID: pointer.To(rand.Intn(60) + 4),
								},
								spineLinkCount: pointer.To(rand.Intn(3) + 1), // 1-3
								spineLinkSpeed: speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
								tagIDs:         random.SomeOf(tagIDs, 2, 6),
							},
						},
						accessSwitches: map[string]resourceRackTypeAccessSwitch{
							random.PersistentString("l3clos_one_of_everything_a1", 10, acctest.CharSetAlpha): {
								count:           rand.Intn(3) + 1, // 1-3
								logicalDeviceID: random.OneOf(ldIDs...),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_l1", 10, acctest.CharSetAlpha),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
							random.PersistentString("l3clos_one_of_everything_a2", 10, acctest.CharSetAlpha): {
								count:           rand.Intn(3) + 1, // 1-3
								logicalDeviceID: random.OneOf(ldIDs...),
								links: map[string]resourceRackTypeLink{
									acctest.RandStringFromCharSet(10, acctest.CharSetAlpha): {
										targetSwitchName: random.PersistentString("l3clos_one_of_everything_l2", 10, acctest.CharSetAlpha),
										lagMode:          &enum.LAGModeActiveLACP,
										linksPerSwitch:   pointer.To(rand.Intn(3) + 1), // 1-3
										speed:            speed.Speed(random.OneOf("1G", "10G", "25G", "100G")),
										tagIDs:           random.SomeOf(tagIDs, 2, 6),
									},
								},
								tagIDs: random.SomeOf(tagIDs, 2, 6),
							},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceRackType)
	for tName, tCase := range testCases {
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.versionConstraints.Check(version.Must(version.NewVersion(client.ApiVersion()))) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.versionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName)

				var checkLog string
				var checkFuncs []resource.TestCheckFunc

				for _, checkList := range checks {
					checkLog = checkLog + checkList.string(len(checkFuncs))
					checkFuncs = append(checkFuncs, checkList.checks...)
				}

				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, checkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checkFuncs...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
