# The following example retrieves a Config Template from a Freeform Blueprint
data "apstra_freeform_config_template" "interfaces" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  name         = "interfaces.jinja"
}
