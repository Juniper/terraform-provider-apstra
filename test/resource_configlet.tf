data "apstra_configlet" "d" {
  name = "US-MOTD"
}


resource "apstra_configlet" "r" {
  name        = "TEST-MOTD-AGAIN"
  ref_archs   = ["two_stage_l3clos"]
  generators  = [
  {
    config_style           = "nxos"
    filename               = null
    negation_template_text = <<-EOT
                              no banner login
                              EOT
    section                = "system"
    template_text          = <<-EOT
                    banner login #
                    --------------------------------------------------------
                    THIS IS A TEST SCRIPT
                    --------------------------------------------------------
                    #
                     EOT
  },
 {
    config_style           = "eos"
    filename               = null
    negation_template_text = <<-EOT
                      no banner motd
                    EOT
    section                = "system"
    template_text          = <<-EOT
                    banner motd
                    --------------------------------------------------------
                    This is a TEST SCRIPT
                    --------------------------------------------------------
                    EOF
                EOT
 },
  ]
}
