# This example defines a freeform system in a blueprint


resource "apstra_freeform_system" "test" {
  blueprint_id = "freeform_blueprint-5ba09d07"
  name         = "test_system"
  tags = ["a", "b", "c"]
  type         = "internal"
  hostname = "testsystem"
  deploy_mode = "deploy"
  device_profile_id = "PtrWb4-VSwKiYRbCodk"
}

# here we retrieve the freeform system

data "apstra_freeform_system" "test" {
  blueprint_id = "freeform_blueprint-5ba09d07"
  id = apstra_freeform_system.test.id
}

# here we build an output bock to display it

output "test_System_out" {value = data.apstra_freeform_system.test}

#Output looks like this
#test_System_out = {
#  "blueprint_id" = "freeform_blueprint-5ba09d07"
#  "deploy_mode" = tostring(null)
#  "device_profile_id" = "PtrWb4-VSwKiYRbCodk"
#  "hostname" = "systemfoo"
#  "id" = "-63CYLAiWuAq0ljzX0Q"
#  "name" = "test_system_foo"
#  "system_id" = tostring(null)
#  "tags" = toset([
#    "a",
#    "b",
#    "c",
#  ])
#  "type" = "internal"
#}

