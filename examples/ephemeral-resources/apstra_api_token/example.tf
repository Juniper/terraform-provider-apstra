# This example runs uses the apstra_api_token ephemeral resource
# to collect an API key which is then used in an arbitrary shell
# command.
locals {
  url = "https://apstra.example.com"
}

provider "apstra" {
  url                     = local.url
  blueprint_mutex_enabled = false
}

// fetch an API token
ephemeral "apstra_api_token" "example" {}

resource "null_resource" "example" {
  provisioner "local-exec" {
    command = "curl -X OPTIONS $URL -H 'accept: application/json' -H \"Authtoken: $APIKEY\" > /tmp/local-exec.out"
    environment = {
      APIKEY = ephemeral.apstra_api_token.example.value
      URL    = format("%s/api/blueprints", local.url)
    }
  }
}
