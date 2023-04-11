package utils

import (
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"testing"
)

func TestRosetta(t *testing.T) {
	type tc struct {
		string    string
		stringers []fmt.Stringer
	}

	testCases := []tc{
		{string: "unique", stringers: []fmt.Stringer{apstra.AsnAllocationSchemeDistinct}},
		{string: "single", stringers: []fmt.Stringer{apstra.AsnAllocationSchemeSingle}},

		{string: "delete_based_interface", stringers: []fmt.Stringer{apstra.ConfigletSectionDeleteBasedInterface, apstra.PlatformOSCumulus}},
		{string: "file", stringers: []fmt.Stringer{apstra.ConfigletSectionFile}},

		{string: "top_level_hierarchical", stringers: []fmt.Stringer{apstra.ConfigletSectionSystem, apstra.PlatformOSJunos}},
		{string: "top_level_set_delete", stringers: []fmt.Stringer{apstra.ConfigletSectionSetBasedSystem, apstra.PlatformOSJunos}},
		{string: "interface_level_hierarchical", stringers: []fmt.Stringer{apstra.ConfigletSectionInterface, apstra.PlatformOSJunos}},
		{string: "interface_level_set", stringers: []fmt.Stringer{apstra.ConfigletSectionSetBasedInterface, apstra.PlatformOSJunos}},
		{string: "interface_level_delete", stringers: []fmt.Stringer{apstra.ConfigletSectionDeleteBasedInterface, apstra.PlatformOSJunos}},

		{string: "static", stringers: []fmt.Stringer{apstra.OverlayControlProtocolNone}},
		{string: "evpn", stringers: []fmt.Stringer{apstra.OverlayControlProtocolEvpn}},

		{string: "datacenter", stringers: []fmt.Stringer{apstra.RefDesignDatacenter}},
		{string: "freeform", stringers: []fmt.Stringer{apstra.RefDesignFreeform}},
	}

	for i, tc := range testCases {
		// test creating friendly string from iota/stringer type
		result := StringersToFriendlyString(tc.stringers...)
		if result != tc.string {
			t.Fatalf("testcase [%d], expected %q, got %q", i, tc.string, result)
		}

		// test creating iota/stringer type from friendly string
		var target StringerWithFromString
		switch tc.stringers[0].(type) {
		case apstra.ConfigletSection:
			x := apstra.ConfigletSection(-1)
			target = &x
		case apstra.AsnAllocationScheme:
			x := apstra.AsnAllocationScheme(-1)
			target = &x
		case apstra.OverlayControlProtocol:
			x := apstra.OverlayControlProtocol(-1)
			target = &x
		case apstra.RefDesign:
			x := apstra.RefDesign(-1)
			target = &x
		}

		if target == nil {
			t.Fatalf("missing case above - target is nil")
		}

		// stringsWithContext is the []string sent to the rosetta function to populate target
		stringsWithContext := []string{tc.string}
		for _, s := range tc.stringers[1:] {
			stringsWithContext = append(stringsWithContext, s.String())
		}

		// populate the target
		err := ApiStringerFromFriendlyString(target, stringsWithContext...)
		if err != nil {
			t.Fatalf("[%d] produced error: %s", i, err.Error())
		}

		// invoke the un-translated String() method to compare against the original input
		if target.String() != tc.stringers[0].String() {
			t.Fatalf("[%d] got %s expected %s", i, tc.stringers[0], target.String())
		}
	}
}
