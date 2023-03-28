terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  tls_validation_disabled = true
  blueprint_mutex_disabled = true
}