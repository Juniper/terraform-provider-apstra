terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  url = "https://admin:EuphoricTiglon9!@13.38.32.96:32609"
  tls_validation_disabled = true
  blueprint_mutex_disabled = true
}