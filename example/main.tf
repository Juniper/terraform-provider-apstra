terraform {
  required_providers {
    apstra = {
      source  = "example.com/chrismarget-j/apstra"
    }
  }
}

provider "apstra" {
  i_dont_care_about_tls_verification_and_i_should_feel_bad = true
}

#resource "apstra_asn_pool" "my_pool" {
#  display_name = "terraform did this"
#  tags = ["baz", "bang", "whoosh"]
#}
#
#resource "apstra_asn_pool_range" "my_pool" {
#  count = 10
#  pool_id = apstra_asn_pool.my_pool.id
#  first = (count.index * 100) + 1
#  last = (count.index * 100) + 100
#}
#
#data "apstra_asn_pool" "my_pool" {
#  id = apstra_asn_pool.my_pool.id
#}

#data "apstra_asn_pools" "all_pools" {}

#data "apstra_asn_pool_id" "default_4_byte_pool" {
#  display_name = "terraform did this"
#  tags = ["baz"]
#}

#output "my_pool" {
#  value = data.apstra_asn_pool.my_pool.id
#}
#
#output "all_pools" {
#  value = data.apstra_asn_pools.all_pools.ids
#}

#output "lookup_pool" {
#  value = data.apstra_asn_pool_id.default_4_byte_pool.id
#}

#data "apstra_agent_profiles" "agent_profile_ids" {}
#
#output "agent_profile_ids" {
#  value = data.apstra_agent_profiles.agent_profile_ids.ids
#}

data "apstra_agent_profile" "foo" {
  # specify one of 'name' or 'id'
#  name = "manual"
  id = "99dbe9da-44e4-4de5-9f50-ebb26cd4934d"
}

output "agent_profile_manual" {
  value = data.apstra_agent_profile.foo
}

resource "apstra_agent_profile" "ap" {
  name = "terraform"
  platform = "junos"
  packages = {
    "bar" = "1.1"
    "foo" = "2.2"
  }
  open_options = {
    "op1" = "val1"
    "op2" = "val2"
  }
}