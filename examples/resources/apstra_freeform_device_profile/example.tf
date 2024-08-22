# This example imports a device profile from the Global Catalog
# into a preexisting Freeform Blueprint.

resource "apstra_freeform_device_profile" "test" {
  blueprint_id      = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  device_profile_id = "vJunosEvolved"
}
