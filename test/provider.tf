terraform {
  required_providers {
    apstra = {
      source = "Juniper/apstra"
    }
  }
}

provider "apstra" {
  url = "https://admin:admin@10.6.1.44/"
  tls_validation_disabled = true
  blueprint_mutex_disabled = true
}
