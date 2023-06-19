terraform {
  required_providers {
    apstra = {
      source = "Juniper/apstra"
    }
  }
}

provider "apstra" {
  url = "https://admin:VirtuousMoose4%5E@13.37.222.29:21809/"
  tls_validation_disabled = true
  blueprint_mutex_disabled = true
}
