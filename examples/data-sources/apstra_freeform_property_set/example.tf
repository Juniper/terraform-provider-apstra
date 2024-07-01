# This example retrieves one property set from a blueprint

# first we create a property set so we can use a data source to retrieve it.
resource "apstra_freeform_property_set" "prop_set_foo" {
  blueprint_id = "freeform_blueprint-5ba09d07"
  name         = "prop_set_foo"
  values       = jsonencode({
    foo   = "bar"
    clown = 2
  })
}

# here we retrieve the property_set.

data "apstra_freeform_property_set" "foo" {
  blueprint_id = "freeform_blueprint-5ba09d07"
  name         = apstra_freeform_property_set.prop_set_foo.name
}

#here we build an output block to display it.
output "foo" {value = data.apstra_freeform_property_set.foo}

#Output looks like this
#foo = {
#  "blueprint_id" = "freeform_blueprint-5ba09d07"
#  "id" = tostring(null)
#  "name" = "prop_set_foo"
#  "system_id" = tostring(null)
#  "values" = "{\"clown\": 2, \"foo\": \"bar\"}"
#}

