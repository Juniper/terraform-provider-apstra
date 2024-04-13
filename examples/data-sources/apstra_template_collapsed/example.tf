# This example fetches the details of a Collapsed Template

data "apstra_template_collapsed" "example" {
# id   = "4ef45fb3-4e7c-4bbd-8378-a0722ee8ba38"  # must specify either id
  name = "example collapsed template"            # or name in data source
}

