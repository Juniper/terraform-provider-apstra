#resource "asptra_managed_Device" "spine1" {
#
#}

resource "apstra_blueprint" "bp" {
  name        = "aaa"
  template_id = "3cd561f2-250c-47a2-af3a-fdc802e22864"
  #  superspine_spine_addressing = "ipv4"
  #  leaf_asn_pool_ids = ["Private-64512-65534"]
  #  leaf_ip4_pool_ids = ["a88df6c4-deef-4b33-b95f-8d6f0df88f8e"]
  #  superspine_asn_pool_ids = ["f50bf427-30d5-40cb-9a18-60722777aeb6", "f72ddab6-7946-43c0-85db-0794f7be7891"]
  #  superspine_asn_pool_ids = ["5363446d-389d-440e-9b6a-d3fdc95fc064"]
  spine_asn_pool_ids                = ["5363446d-389d-440e-9b6a-d3fdc95fc064", "4bdc942f-4561-415d-b977-65818640e69b"]

  leaf_asn_pool_ids               = ["f50bf427-30d5-40cb-9a18-60722777aeb6"]
  access_asn_pool_ids                 = ["f72ddab6-7946-43c0-85db-0794f7be7891"]
#  access_asn_pool_ids               = ["f50bf427-30d5-40cb-9a18-60722777aeb6"]
#  leaf_asn_pool_ids                 = ["f72ddab6-7946-43c0-85db-0794f7be7891"]

  access_esi_peer_link_ip4_pool_ids = ["Private-10_0_0_0-8"]
  access_loopback_pool_ids          = ["Private-10_0_0_0-8"]
  leaf_loopback_pool_ids            = ["Private-10_0_0_0-8"]
  spine_leaf_ip4_pool_ids           = ["Private-10_0_0_0-8"]
  spine_loopback_pool_ids           = ["Private-10_0_0_0-8"]
  vtep_ip4_pool_ids                 = ["Private-10_0_0_0-8"]

  switches = {
    spine1 = {
      interface_map_id = "Juniper_vQFX__AOS-7x10-Spine"
#      interface_map_id = "Accton_5712-54X-O_Cumulus__AOS-48x10_6x40-1"
      device_key = "525400F872B7"
    }
  }

}
