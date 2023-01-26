terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  url = "https://admin:TrustworthyKite5_@18.116.164.244:29609"
  tls_validation_disabled = true
}
