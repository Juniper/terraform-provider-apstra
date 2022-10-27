terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  url = "https://3.22.233.96:20609"
  tls_validation_disabled = true
}
