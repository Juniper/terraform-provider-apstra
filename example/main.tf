terraform {
  required_providers {
    apstra = {
      source  = "example.com/chrismarget-j/apstra"
    }
  }
}

provider "apstra" {
  scheme = "https"
  host     = "66.129.234.203"
  port     = 31000
  username = "admin"
  password = "admin"
  tls_no_verify = true
}

data "apstra_asn_pools" "pools" {}

output "pools" {
  value = data.apstra_asn_pools.pools
}
