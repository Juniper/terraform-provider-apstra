terraform {
  required_providers {
    apstra = {
      source = "example.com/apstrktr/apstra"
    }
  }
}

provider "apstra" {
  url = "https://admin@18.220.223.232:25559"
  tls_validation_disabled = true
}
