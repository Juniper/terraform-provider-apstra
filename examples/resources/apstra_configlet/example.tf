# This example creates a configlet responsible for DNS server
# addresses on Junos and EOS devices.
locals {
  cfg_data = {
    name = "DNS according to Terraform"
    generators = [
      {
        config_style  = "junos"
        section       = "top_level_hierarchical"
        template_text = <<-EOT
        name-server {
          4.2.2.1;
          4.2.2.2;
        }
      EOT
      },
      {
        config_style           = "eos"
        section                = "system"
        template_text          = "ip name-server 4.2.2.1 4.2.2.2"
        negation_template_text = "no ip name-server 4.2.2.1 4.2.2.2"
      }
    ]
  }
}
resource "apstra_configlet" "example" {
  data = local.cfg_data
}
