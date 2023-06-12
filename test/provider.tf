terraform {
  required_providers {
    apstra = {
      source = "Juniper/apstra"
    }
  }
}

provider "apstra" {
  tls_validation_disabled = true
  blueprint_mutex_disabled = true
}

