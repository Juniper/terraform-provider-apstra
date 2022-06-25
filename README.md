# terraform-provider-apstra

This is the beginning of a Terraform provider for Juniper Apstra. It relies on
a Go client library at https://github.com/chrismarget-j/goapstra

### Getting Started

1. Get access to both this repo and the client library repo, both of which are
   currently private. You're reading this, so that's a good start.
1. Configure `go` so that it doesn't try to use a public mirror for these repos
   by running the following shell command:
   ```shell
   go env -w GOPRIVATE=github.com/chrismarget-j
   ```
1. Add an `ssh` key to your github account.
1. Configure `git` to use `ssh` for these repositories with this shell command:
   ```shell
   git config --global url.git@github.com:.insteadOf https://github.com/
   ```
1. Install [`terraform`](https://www.terraform.io/downloads) and [`Go`](https://go.dev/dl/) on your machine.
1. Clone this repo to your local system:
   ```shell
   git@github.com:chrismarget-j/terraform-provider-apstra.git
   ```
1. Build the provider:
   ```shell
   cd terraform-provider-apstra
   go install
   ```
   The provider should now be located at `~/golang/bin/terraform-provider-apstra`
1. Configure terraform to look for the apstra provider locally. My `~/.terraformrc` looks like:
   ```
   provider_installation {
   
     dev_overrides {
       "example.com/chrismarget-j/apstra" = "/Users/cmarget/golang/bin"
     }
   
     # For all other providers, install them directly from their origin provider
     # registries as normal. If you omit this, Terraform will _only_ use
     # the dev_overrides block, and so no other providers will be available.
     direct {}
   }
   ```
1. Optional: The provider allows you to set Apstra credentials in the
   `terraform` configuration, but can also get those details from the environment.
    Don't put passwords in config files. Use environment variables:
   ```shell
   export APSTRA_SCHEME=https               # https is default, can omit
   export APSTRA_USER=<username>
   export APSTRA_PASS=<password>
   export APSTRA_HOST=<hostname or IP>
   export APSTRA_PORT=443                   # 443 is default, can omit
   export APSTRA_API_TLS_LOGFILE=<filename> # useful for decrypting transactions with wireshark
   ```
1. Change into the `example` directory, and apply a configuration!
   ```shell
   cd example
   terraform plan
   ```
   
### Resources
ToDo: check out `example/main.tf`

### Data Sources
ToDo: check out `example/main.tf`
