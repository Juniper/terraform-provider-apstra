# Create a device profile for us to look kup by name.
# We'll pretend we don't already have the ID available.
resource "apstra_modular_device_profile" "my_ptx_10k" {
  name               = "PTX 10K with Two Cards"
  chassis_profile_id = "Juniper_PTX10008"
  line_card_profile_ids = {
    3 = "Juniper_PTX10K_LC1201_36CD"
    4 = "Juniper_PTX10K_LC1201_36CD"
  }
}

# Discover the device profile we just created, using its name. This is a
# common pattern when you want to use a resource that was created outside
# of Terraform, or if you want to look up a resource by name instead of ID
# because, for example, you're maintaining identical objects in two
# different Apstra backends.
data "apstra_device_profile" "my_ptx_10k" {
  name = apstra_modular_device_profile.my_ptx_10k.name
}

output "my_ptx_10k" { value = data.apstra_device_profile.my_ptx_10k }

# The output from this data source looks like this:

#    Outputs:
#
#    my_ptx_10k = {
#      "id" = "6a7b63b7-1f3b-4d32-b9b0-d6c51b20f4ef"
#      "name" = "PTX 10K with Two Cards"
#    }

