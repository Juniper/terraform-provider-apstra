# terraform-provider-apstra

This is a [Terraform](https://www.terraform.io)
[provider](https://developer.hashicorp.com/terraform/language/providers?page=providers)
for Juniper Apstra. It relies on a Go client library at https://github.com/Juniper/apstra-go-sdk

## Getting Started

### Install Terraform

Instructions for popular operating systems can be found [here](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli).

### Create a Terraform configuration

The terraform configuration must:
- be named with a `.tf` file extension.
- reference this provider by its global address.
  *registry.terraform.io/Juniper/apstra* or just: *Juniper/apstra*.
- include a provider configuration block which tells the provider where to
find the Apstra service.

```hcl
terraform {
  required_providers {
    apstra = {
      source = "Juniper/apstra"
    }
  }
}

provider "apstra" {
  url = "<apstra-server-url>"
}
```

### Terraform Init

Run the following at a command prompt while in the same directory as the
configuration file to fetch the Apstra provider plugin.
```shell
terraform init
```

### Supply Apstra credentials
Apstra credentials can be supplied through environment variables:
```shell
export APSTRA_USER=<username>
export APSTRA_PASS=<password>
```

Alternatively, credentials can be embedded in the URL using HTTP basic
authentication format (we don't actually *do* basic authentication, but the
format is: `https://user:password@host`). Any special characters in the username
and password must be URL-encoded when using this approach.

### Start configuring resources

Full documentation for provider, resources and data sources can be found
[here](https://registry.terraform.io/providers/Juniper/apstra/latest/docs).