terraform {
  required_providers {
    apstra = {
      source = "example.com/chrismarget-j/apstra"
    }
  }
}

// this stanza is completely optional. Without it, the provider will look for
// hostname/username/password in the environment, use https, and insist on cert
// validation.
provider "apstra" {
  scheme                                                   = "https"              # optional, alternate env var APSTRA_SCHEME, default 'https'
  host                                                     = "apstra.example.com" # optional, alternate env var APSTRA_HOST
  port                                                     = 443                  # optional, alternate env var APSTRA_PORT, default 443
  username                                                 = "admin"              # optional, alternate env var APSTRA_USER
  password                                                 = "admin"              # optional, alternate env var APSTRA_PASS
  i_dont_care_about_tls_verification_and_i_should_feel_bad = true
}
