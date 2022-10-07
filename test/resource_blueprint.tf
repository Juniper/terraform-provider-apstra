resource "apstra_blueprint" "bp" {
  name        = "aaa from tf"
  template_id = "3cd561f2-250c-47a2-af3a-fdc802e22864"
#  superspine_spine_addressing = "ipv4"
#  leaf_asn_pool_ids = ["Private-64512-65534"]
#  leaf_ip4_pool_ids = ["a88df6c4-deef-4b33-b95f-8d6f0df88f8e"]
  spine_asn_pool_ids = ["Private-64512-65534"]
}
