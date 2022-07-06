terraform {
  required_providers {
    apstra = {
      source = "example.com/chrismarget-j/apstra"
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

// create an ASN pool with no ASN ranges defined
#resource "apstra_asn_pool" "my_pool" {
#  name = "my_pool"
#  tags = ["foo", "bar"] # optional, does not appear in web UI
#}

// create ASN ranges within the ASN pool
#resource "apstra_asn_pool_range" "one_hundred_ASNs" {
#  pool_id = apstra_asn_pool.my_pool.id
#  first   = 1
#  last    = 100
#}

// look up the details of an ASN pool by ID number
#data "apstra_asn_pool" "my_pool" {
#  id = apstra_asn_pool.my_pool.id
#}
// data.apstra_asn_pool output looksl like:
/*
{
  "created_at" = "1970-01-01 00:00:00 +0000 UTC"
  "name" = "my_pool"
  "id" = "1f712e32-8187-4c8d-a720-73ef1bae5c34"
  "last_modified_at" = "2022-06-25 02:57:16.332729 +0000 UTC"
  "ranges" = tolist([
    {
      "first" = 1
      "last" = 100
      "status" = "pool_element_available"
      "total" = 500
      "used" = 0
      "used_percentage" = 0
    },
    {
      "first" = 101
      "last" = 200
      "status" = "pool_element_available"
      "total" = 500
      "used" = 0
      "used_percentage" = 0
    },
  ])
  "status" = "not_in_use"
  "tags" = tolist([
    "bar",
    "foo",
  ])
  "total" = 200
  "used" = 0
  "used_percentage" = 0
}
*/

// look up ID numbers of all ASN pools
#data "apstra_asn_pools" "all_pools" {}
// data.apstra_asn_pools output looks like:
/*
{
  "ids" = toset([
    "1ef214d6-3810-4ab9-a673-4cd45e535d03",
    "1f712e32-8187-4c8d-a720-73ef1bae5c34",
    "926b59bb-291a-4ce0-bd93-7e9f20ce0dc2",
    "Private-4200000000-4294967294",
    "Private-64512-65534",
  ])
}
*/

// Create an agent profile. note that we cannot reasonably manage the username
// or password in the profile via terraform, because we cannot check the state.
// A feature enhancement which returns a partial credential hash or timestamp
// would likely make it possible to drive these credentials via terraform.
// That may be a good thing: Filling TF config and state with secrets is a
// bummer. For now, add the credentials (or the whole agent profile) manually
// via web UI.
#resource "apstra_agent_profile" "my_agent_profile" {
#  name     = "my agent profile"
#  platform = "junos"
#  #  packages = { # optional
#  #    "foo" = "1.1"
#  #    "bar" = "2.2"
#  #  }
#  #  open_options = { # optional
#  #    "op1" = "val1"
#  #    "op2" = "val2"
#  #  }
#}

// look up an agent profile
data "apstra_agent_profile" "aap" {
  # must populate either name or its id (not both)
  name = "profile_vqfx"
  #  id = apstra_agent_profile.my_agent_profile.id
}
// data.apstra_agent_profile output looks like:
/*
{
  "has_password" = false
  "has_username" = false
  "id" = "b72dead6-072a-4ed5-a765-a7c79d4dea9c"
  "name" = "my agent profile"
  "open_options" = tomap({
    "op1" = "val1"
    "op2" = "val2"
  })
  "packages" = tomap({
    "bar" = "2.2"
    "foo" = "1.1"
  })
  "platform" = "junos"
}
*/

// List all agent profile IDs. Output looks like:
/*
{
  "ids" = toset([
    "77c27232-8dc0-4e1c-a939-c6c9c1d827fc",
    "99dbe9da-44e4-4de5-9f50-ebb26cd4934d",
    "b72dead6-072a-4ed5-a765-a7c79d4dea9c",
  ])
}
*/
#data "apstra_agent_profiles" "all_agent_profiles" {}

locals {
  switch_info = {
    "525400559292" = { "ip" = "172.20.23.11", "location" = "spine1" },
    "525400601877" = { "ip" = "172.20.23.12", "location" = "spine2" },
    "5254007642A5" = { "ip" = "172.20.23.13", "location" = "leaf1" },
    "5254000A1177" = { "ip" = "172.20.23.14", "location" = "leaf3" },
    "52540035B92A" = { "ip" = "172.20.23.15", "location" = "leaf2" }
  }
}

// create a managed device

#resource "apstra_managed_device" "switch" {
#  for_each         = local.switch_info
#  management_ip    = each.value.ip
#  agent_label      = each.value.location              # optional, does not appear in web UI
#  agent_profile_id = data.apstra_agent_profile.aap.id # required, sets platform type and credentials
#  device_key       = each.key
#}

// create an 'datacenter/two_stage_l3clos' blueprint from an existing template.
resource "apstra_blueprint" "my_blueprint" {
  name               = "my blueprint"
  template_id        = "lab_evpn_mlag"
  spine_asn_pool_ids = ["Private-4200000000-4294967294"]
  leaf_asn_pool_ids  = ["Private-64512-65534"]
  spine_ip_pool_ids  = ["TESTNET-203_0_113_0-24"]
  leaf_ip_pool_ids   = ["Private-172_16_0_0-12"]
  link_ip_pool_ids   = ["Private-10_0_0_0-8", "Private-172_16_0_0-12", "Private-10_0_0_0-8"]
}