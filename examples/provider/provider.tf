# Configuring the Apstra provider is completely optional. The only feature which
# cannot be set by environment variable is `tls_validation_disabled`
provider "apstra" {
  #  scheme               = "https"              # optional, alternate env var APSTRA_SCHEME, default 'https'
  host                    = "apstra.example.com" # optional, alternate env var APSTRA_HOST
  #  port                 = 443                  # optional, alternate env var APSTRA_PORT, default 443
  username                = "admin"              # optional, alternate env var APSTRA_USER
  password                = "password"           # optional, alternate env var APSTRA_PASS
  tls_validation_disabled = true
}
