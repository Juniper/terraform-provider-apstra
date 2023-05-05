data "apstra_property_set" "d" {
  name = "TF Property Set"
}
output "d_apstra_property_set" {
  value = data.apstra_property_set.d
}
locals {
	payload = jsonencode({
		value_str  = "str"
		value_int  = 42
		value_json = {
			inner_value_str = "innerstr"
			inner_value_int = 4242
		}
	})
}
resource "apstra_property_set" "r" {
	name = "TF Property Set 1234567"
	data = "{}"
}
