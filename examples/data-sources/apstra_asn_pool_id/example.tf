# The following example shows how a module can use the search capability to deploy a blueprint using a tag-based search for ASN pools:

variable "site" { default = "ashburn"}
variable "env" {default = "dev"}
variable "template" {}
variable "spine_ip_pool" {}
variable "leaf_ip_pool" {}
variable "link_ip_pool" {}

data "apstra_asn_pool_id" "spine" {
   tags = [site, env, "spine"]
}

data "apstra_asn_pool_id" "leaf" {
   tags = [site, env, "leaf"]
}

resource "apstra_blueprint" "my_blueprint" {
   name               = "my blueprint"
   template_id        = var.bp_template
   spine_asn_pool_ids = [data.apstra_asn_pool_id.spine.id]
   leaf_asn_pool_ids  = [data.apstra_asn_pool_id.leaf.id]
   spine_ip_pool_ids  = var.spine_ip_pool
   leaf_ip_pool_ids   = var.leaf_ip_pool
   link_ip_pool_ids   = var.link_ip_pool
}
