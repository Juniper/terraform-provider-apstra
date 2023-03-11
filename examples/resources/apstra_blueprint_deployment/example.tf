# This example commits a blueprint, including:
# - rack type
# - template
# - blueprint instantiation from template
#
# It then assigns resource pools and devices (by serial number - these are the
# serial numbers found in a specific cloudlabs "Customer" topology) to roles
# in the fabric

# Instantiate a blueprint from a template
resource "apstra_datacenter_blueprint" "r" {
  name        = "terraform commit example"
  template_id = "L2_Virtual_EVPN"
}

# ASN pools, IPv4 pools and switch devices will be allocated using looping
# resources. These three `local` maps are what we'll loop over.
locals {
  asn_pools = {
    spine_asns = ["Private-64512-65534"]
    leaf_asns  = ["Private-4200000000-4294967294"]
  }
  ipv4_pools = {
    spine_loopback_ips  = ["Private-10_0_0_0-8"]
    leaf_loopback_ips   = ["Private-10_0_0_0-8"]
    spine_leaf_link_ips = ["Private-10_0_0_0-8"]
  }
  switches = {
    spine2               = "Juniper_vQFX__AOS-7x10-Spine"
    spine1               = "Juniper_vQFX__AOS-7x10-Spine"
    l2_virtual_001_leaf1 = "Juniper_vQFX__AOS-7x10-Leaf"
    l2_virtual_002_leaf1 = "Juniper_vQFX__AOS-7x10-Leaf"
    l2_virtual_003_leaf1 = "Juniper_vQFX__AOS-7x10-Leaf"
    l2_virtual_004_leaf1 = "Juniper_vQFX__AOS-7x10-Leaf"
  }
}

# Assign interface maps to fabric roles to eliminate build errors so we can deploy
resource "apstra_datacenter_blueprint_device_allocation" "r" {
  for_each         = local.switches
  blueprint_id     = apstra_datacenter_blueprint.r.id
  node_name        = each.key
  interface_map_id = each.value
}

# Assign ASN pools to fabric roles to eliminate build errors so we can deploy
resource "apstra_datacenter_blueprint_resource_pool_allocation" "asn" {
  for_each     = local.asn_pools
  blueprint_id = apstra_datacenter_blueprint.r.id
  role         = each.key
  pool_ids     = each.value
}

# Assign IPv4 pools to fabric roles to eliminate build errors so we can deploy
resource "apstra_datacenter_blueprint_resource_pool_allocation" "ipv4" {
  for_each     = local.ipv4_pools
  blueprint_id = apstra_datacenter_blueprint.r.id
  role         = each.key
  pool_ids     = each.value
}

# The only required field for deployment is blueprint_id, but we're ensuring
# sensible run order and setting a custom commit message.
resource "apstra_blueprint_deployment" "deploy" {
  blueprint_id = apstra_datacenter_blueprint.r.id

  #ensure that deployment doesn't run before build errors are resolved
  depends_on = [
    apstra_datacenter_blueprint_device_allocation.r,
    apstra_datacenter_blueprint_resource_pool_allocation.asn,
    apstra_datacenter_blueprint_resource_pool_allocation.ipv4,
  ]

  # Version is replaced using `text/template` method. Only predefined values
  # may be replaced with this syntax. USER is replaced using values from the
  # environment. Any environment variable may be specified this way.
  comment      = "Deployment by Terraform {{.TerraformVersion }}, Apstra provider {{.ProviderVersion}}, User $USER."
}