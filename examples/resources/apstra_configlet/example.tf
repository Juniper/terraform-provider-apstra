# This example creates a configlet responsible for DNS server
# addresses on Junos and EOS devices.
resource "apstra_configlet" "example" {
  name = "DNS according to Terraform"
  generators = [
    {
      config_style  = "junos"
      section       = "system"
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
