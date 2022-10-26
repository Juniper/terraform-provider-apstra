terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  url = "https://3.14.12.106:35309"
  tls_validation_disabled = true
}
