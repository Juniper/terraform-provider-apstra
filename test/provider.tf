terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  url = "https://admin:AmazingFinch5%40@3.15.181.110:35159"
  tls_validation_disabled = true
}
