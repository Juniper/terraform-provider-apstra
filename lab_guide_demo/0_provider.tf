terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  # URL and credentials can be supplied using the "url" parameter in this file.
  #  url = "https://<username>:<password>@<hostname-or-ip-address>:<port>"
  #
  # ...or using the environment variable APSTRA_URL.
  #
  # If Username or Password are not embedded in the URL, the provider will look
  # for them in the APSTRA_USER and APSTRA_PASS environment variables.
  #
  tls_validation_disabled = true  # CloudLabs doesn't present a valid TLS cert
  blueprint_mutex_disabled = true # Don't attempt worry about competing clients
}
