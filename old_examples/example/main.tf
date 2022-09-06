terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}


// this stanza is completely optional. Without it, the provider will look for
// hostname/username/password in the environment, use https, and insist on cert
// validation.
provider "apstra" {
  # scheme = "https" # optional, alternate env var APSTRA_SCHEME, default 'https'
  # host = ""        # optional, alternate env var APSTRA_HOST
  # port = 443       # optional, alternate env var APSTRA_PORT, default 443
  # username = ""    # optional, alternate env var APSTRA_USER
  # password = ""    # optional, alternate env var APSTRA_PASS
  i_dont_care_about_tls_verification_and_i_should_feel_bad = true
}

locals {
  env  = "prod"
  site = "boston"

  switch_info = {
    "525400F069AD" = { "ip" = "172.20.16.11", "platform" = "junos", "role" = "leaf1" },
    "525400807DD0" = { "ip" = "172.20.16.12", "platform" = "junos", "role" = "spine2" },
    "525400A84170" = { "ip" = "172.20.16.13", "platform" = "junos", "role" = "leaf3" },
    "525400B05278" = { "ip" = "172.20.16.14", "platform" = "junos", "role" = "spine1" },
    "525400463527" = { "ip" = "172.20.16.15", "platform" = "junos", "role" = "leaf2" }
  }

  switch_role = {
    spine = { asn = [100, 199], ip = "10.0.0.0/24", }
    leaf  = { asn = [200, 299], ip = "10.1.0.0/24", }
  }

  link_addressing = "10.2.0.0/24"

  default_tags = [local.env, local.site]
}

resource "apstra_blueprint" "my_blueprint" {
  name               = "your blueprint"
  template_id        = "lab_evpn_mlag"
  leaf_asn_pool_ids  = ["Private-64512-65534"]
  leaf_ip_pool_ids   = ["Private-192_168_0_0-16"]
  link_ip_pool_ids   = ["Private-192_168_0_0-16"]
  spine_asn_pool_ids = ["Private-64512-65534"]
  spine_ip_pool_ids  = ["Private-192_168_0_0-16"]
  switches = {
    spine1                = { device_key = "525400F069AD" }
    spine2                = { device_key = "525400807DD0" }
    evpn_single_001_leaf1 = { device_key = "525400A84170" }
    evpn_esi_001_leaf1    = { device_key = "525400B05278" }
    evpn_esi_001_leaf2    = { device_key = "525400463527" }

  }
}

resource "apstra_rack_type" "foo" {
  name                       = "terraform"
  description                = "very important rack"
  fabric_connectivity_design = "l3clos"
  leaf_switches              = {
    leaf = {
      spine_link_count    = 1
      spine_link_speed    = "10G"
      logical_device_id   = "AOS-32x10-3"
      redundancy_protocol = "esi"
    }
  }
    generic_systems = {
#      dual_homed = {
#        count=1
#        logical_device_id = "AOS-2x10-1"
#        links = [
#          {
#            name = "link_to_both"
#            links_per_switch = 1
#            speed = "10g"
#            target_switch_name = "leaf"
#            lag_mode = "lacp_active"
#          }
#        ]
#      },
      single_homed_a = {
        count=1
        logical_device_id = "AOS-2x10-1"
        links = [
          {
            name = "link_to_a"
            links_per_switch = 1
            speed = "10G"
            target_switch_name = "leaf"
            switch_peer = "first"
          }
        ]
      },
    }
}

#data "apstra_logical_device" "by_name" {
#  name = "AOS-48x10+6x40-1"
#}
#
#output "logical_device_by_name" {
#  value = data.apstra_logical_device.by_name
#}

#data "apstra_tag" "by_id" {
#  id = "bare_metal"
#  key = "foo"
#}
#
#output "apstra_tag_by_id" {
#  value = data.apstra_tag.by_id
#}

#resource "apstra_rack_type" "my_rack" {
#  name = ""
#  description = ""
#  fabric_connectivity_design = ""
#  tags = []
#  leaf_switches = []
#  access_switches = []
#  generic_system_groups = []
#}
