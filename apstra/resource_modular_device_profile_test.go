package tfapstra

import (
	"fmt"
	"strings"
	"testing"

	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	resourceModularDeviceProfileSlotConfigHCL = "\n    %d = \"%s\""
	resourceModularDeviceProfileHCL           = `
resource "apstra_modular_device_profile" "test" {
  name = "%s"
  chassis_profile_id = "%s"
  line_card_profile_ids = {%s
  }
}
`
)

func TestAccResourceModularDeviceProfile(t *testing.T) {
	testutils.TestCfgFileToEnv(t)

	name1 := acctest.RandString(5)
	chassisProfile1 := "Juniper_PTX10008"
	lineCardProfiles1 := map[int]string{
		1: "Juniper_PTX10K_LC1201_36CD",
		3: "Juniper_PTX10K_LC1201_36CD",
	}
	lineCardProfiles1SB := new(strings.Builder)
	for k, v := range lineCardProfiles1 {
		_, _ = lineCardProfiles1SB.WriteString(fmt.Sprintf(resourceModularDeviceProfileSlotConfigHCL, k, v))
	}

	name2 := acctest.RandString(5)
	chassisProfile2 := "Juniper_PTX10016"
	lineCardProfiles2 := map[int]string{
		13: "Juniper_PTX10K_LC1202_36MR",
		14: "Juniper_PTX10K_LC1202_36MR",
		15: "Juniper_PTX10K_LC1202_36MR",
	}
	lineCardProfiles2SB := new(strings.Builder)
	for k, v := range lineCardProfiles2 {
		_, _ = lineCardProfiles2SB.WriteString(fmt.Sprintf(resourceModularDeviceProfileSlotConfigHCL, k, v))
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(
					resourceModularDeviceProfileHCL, name1,
					chassisProfile1, lineCardProfiles1SB.String(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_modular_device_profile.test", "id"),
					// Verify Name has the expected value
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "name", name1),
					// Verify Chassis Profile has the expected value
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "chassis_profile_id", chassisProfile1),
					// Verify Line Card Map
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "line_card_profile_ids.%", "2"),
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "line_card_profile_ids.1", lineCardProfiles1[1]),
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "line_card_profile_ids.3", lineCardProfiles1[3]),
				),
			},
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(
					resourceModularDeviceProfileHCL, name2,
					chassisProfile2, lineCardProfiles2SB.String(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify ID has any value set
					resource.TestCheckResourceAttrSet("apstra_modular_device_profile.test", "id"),
					// Verify Name has the expected value
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "name", name2),
					// Verify Chassis Profile has the expected value
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "chassis_profile_id", chassisProfile2),
					// Verify Line Card Map
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "line_card_profile_ids.%", "3"),
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "line_card_profile_ids.13", lineCardProfiles2[13]),
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "line_card_profile_ids.14", lineCardProfiles2[14]),
					resource.TestCheckResourceAttr("apstra_modular_device_profile.test", "line_card_profile_ids.15", lineCardProfiles2[15]),
				),
			},
		},
	})
}
