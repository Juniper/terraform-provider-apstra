provider "apstra" {
  url                     = "https://apstra.example.com" # required
  tls_validation_disabled = true                         # optional
  blueprint_mutex_enabled = false                        # optional
  api_timeout             = 0                            # optional; 0 == infinite
}
