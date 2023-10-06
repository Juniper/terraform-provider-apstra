# This example creates a device profile for a PTX 10008 with
# two cards: One in slot 3 and one in slot 4.
resource "apstra_modular_device_profile" "example" {
  name               = "PTX 10K with Two Cards"
  chassis_profile_id = "Juniper_PTX10008"
  line_card_profile_ids = {
    3 = "Juniper_PTX10K_LC1201_36CD"
    4 = "Juniper_PTX10K_LC1201_36CD"
  }
}
