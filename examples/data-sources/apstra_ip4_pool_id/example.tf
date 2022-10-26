# The following example shows how a module can use the search capability to deploy a blueprint using a tag-based search for IPv4 pools:

variable "site" { default = "ashburn"}
variable "env" {default = "dev"}
variable "template" {}
variable "spine_asn_pool" {}
variable "leaf_asn_pool" {}

data "apstra_ip4_pool_id" "spine" {
   tags = [site, env, "spine"]
}

data "apstra_ip4_pool_id" "leaf" {
   tags = [site, env, "leaf"]
}

data "apstra_ip4_pool_id" "link" {
  tags = [site, env, "link"]
}

resource "apstra_blueprint" "my_blueprint" {
   name               = "my blueprint"
   template_id        = var.bp_template
   spine_asn_pool_ids = spine.asn_pool
   leaf_asn_pool_ids  = leaf.asn_pool
   spine_ip_pool_ids  = [data.apstra_ip4_pool_id.spine.id]
   leaf_ip_pool_ids   = [data.apstra_ip4_pool_id.leaf.id]
   link_ip_pool_ids   = [data.apstra_asn_pool_id.link.id]
}
