# This example retrireves the RAW JSON representing available Tags
#  from the Apstra service
data "apstra_raw_json" "example" {
  url = "/api/design/tags"
}
