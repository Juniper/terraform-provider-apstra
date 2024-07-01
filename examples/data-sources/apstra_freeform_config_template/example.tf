resource "apstra_freeform_config_template" "test" {
  blueprint_id = "freeform_blueprint-5ba09d07"
  name         = "test_conf_template.jinja"
  tags = ["a", "b", "c"]
  text         = "this is a test for a config template."
}

data "apstra_freeform_config_template" "test" {
  blueprint_id = "freeform_blueprint-5ba09d07"
  id = apstra_freeform_config_template.test.id
}

output "test_out" {value = data.apstra_freeform_config_template.test}
