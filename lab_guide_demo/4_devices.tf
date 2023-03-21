# https://cloudlabs.apstra.com/labguide/Cloudlabs/4.1.2/lab1-junos/lab1-junos-5_devmgnt.html

## Okay, this one's a little tricky. Right now we can't merely "acknowledge"
## switches, but have to take full responsibility for their onboarding:
## - create agent object
## - run agent install job
## - discover and validate system serial number (device key)
## - "acknowledge" system
##
## In order to do that, we need to make note of the switch IPs and serial numbers
## and then un-install and delete the system agent.
##
## Onboarding a system agent requires switch credentials. For that, we lean on an
## Agent Profile.
##
## 1 Update the table below with the correct management ip and serial numbers.
## 2 Un-install and then delete the device agent on each switch in the web UI.
## 3 Add username and password to the Agent Profile named "profile_vqfx":
##   - username: "root"
##   - password: "root123"
## 4 Un-comment the "apstra_managed_device" resource below.
##
## Note the leaf names and management IP addresses are not quite in sequential
## order.
#
#locals {
#  switches = {
#    spine1                  = { management_ip = "172.20.144.11", device_key = "5254003DA3AA" }
#    spine2                  = { management_ip = "172.20.144.12", device_key = "525400B88E65" }
#    apstra_esi_001_leaf1    = { management_ip = "172.20.144.13", device_key = "525400A88F4A" }
#    apstra_esi_001_leaf2    = { management_ip = "172.20.144.15", device_key = "525400F9987D" }
#    apstra_single_001_leaf1 = { management_ip = "172.20.144.14", device_key = "5254004BF852" }
#  }
#}
#
## Look up the details of the Agent Profile to which we've added a username and password.
#data "apstra_agent_profile" "lab_guide" {
#  name = "profile_vqfx"
#}
#
## Onboard each switch. This will be a comparatively long "terraform apply".
#resource "apstra_managed_device" "lab_guide" {
#  for_each         = local.switches
#  agent_profile_id = data.apstra_agent_profile.lab_guide.id
#  off_box          = true
#  management_ip    = each.value.management_ip
#  device_key       = each.value.device_key
#}
