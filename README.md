# terraform-provider-apstra

This is the beginning of a [Terraform](https://www.terraform.io)
[provider](https://developer.hashicorp.com/terraform/language/providers?page=providers)
for Juniper Apstra. It relies on a Go client library at https://github.com/Juniper/apstra-go-sdk

## Getting Started

You can build from source or use a precompiled binary.

#### Build from source

1. Install [`Terraform`](https://www.terraform.io/downloads) and [`Go`](https://go.dev/dl/)
   on your machine.
1. Clone this repo to your local system:
   ```shell
   git clone git@github.com:Juniper/terraform-provider-apstra.git
   ```
1. Build the provider:
   ```shell
   cd terraform-provider-apstra
   go install
   ```
   The provider binary should have appeared at `$GOROOT/bin/terraform-provider-apstra`. In my case, that's
   `~/golang/bin/terraform-provider-apstra`.

#### Use a precompiled binary

1. Head over to the repository's [releases](https://github.com/chrismarget-j/terraform-provider-apstra/releases)
page, select a version, expand the `assets` tree, and download the zip file appropriate for your system.
   
1. Unpack the zip file and copy the `terraform-provider-apstra_vX.X.X` binary to wherever you keep this sort of thing.

### Configure Terraform to use the provider binary

Terraform needs to know that it should use the provider binary you've just compiled or downloaded, rather than 
trying to fetch it from a registry server somewhere. The easiest way to do this is to use a `dev_overrides`
directive.

My `~/.terraformrc` looks like:

```hcl
provider_installation {

    dev_overrides {
        "example.com/apstrktr/apstra" = "/Users/cmarget/golang/bin"
    }

    # For all other providers, install them directly from their origin provider
    # registries as normal. If you omit this, Terraform will _only_ use
    # the dev_overrides block, and so no other providers will be available.
    direct {}
}
```

### Optional: Put credentials in the environment
The provider allows you to set Apstra credentials in the `terraform` configuration,
but can also get those details from the environment. Don't put passwords in config
files. Use environment variables:

```shell
export APSTRA_USER=<username>
export APSTRA_PASS=<password>
```
