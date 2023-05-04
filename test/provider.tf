terraform {
  required_providers {
    apstra = {
      source = "example.com/juniper/apstra"
    }
  }
}

provider "apstra" {
  tls_validation_disabled = true
  blueprint_mutex_disabled = true
 # url = "https://admin:HappyDuck6%5E@35.92.136.236:32909"
  url = "https://admin:admin@10.6.1.44/"
}
