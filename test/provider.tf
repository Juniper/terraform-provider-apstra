terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  url = "https://admin:ShinyHedgehog7-@13.38.52.89:27959"
  tls_validation_disabled = true
}
