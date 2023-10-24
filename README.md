[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]

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
```console
terraform init
```

### Apstra credentials can be supplied through environment variables:
```console
export APSTRA_USER=<username>
export APSTRA_PASS=<password>
```

Alternatively, credentials can be embedded in the URL using HTTP basic
authentication format (we don't actually *do* basic authentication, but the
format is: ``https://user:pass@host``). Any special characters in the username
and password must be URL-encoded when using this approach.

### Start configuring resources

Full documentation for provider, resources and data sources can be found
[here](https://registry.terraform.io/providers/Juniper/apstra/latest/docs).

See the [open issues](https://github.com/Juniper/terraform-provider-apstra/issues) for a full list of proposed features (and known issues).

## Public Roadmap

View the [Roadmap](https://github.com/orgs/Juniper/projects/5/views/2) if you would like to see what is next on our plate.


## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
1. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
1. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
1. Push to the Branch (`git push origin feature/AmazingFeature`)
1. Open a Pull Request


            
            
            

[contributors-shield]: https://img.shields.io/github/contributors/Juniper/terraform-provider-apstra?style=for-the-badge
[contributors-url]: https://github.com/Juniper/terraform-provider-apstra/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/Juniper/terraform-provider-apstra.svg?style=for-the-badge
[forks-url]: https://github.com/Juniper/terraform-provider-apstra/network/members
[stars-shield]: https://img.shields.io/github/stars/Juniper/terraform-provider-apstra.svg?style=for-the-badge
[stars-url]: https://github.com/Juniper/terraform-provider-apstra/stargazers
[issues-shield]: https://img.shields.io/github/issues/Juniper/terraform-provider-apstra.svg?style=for-the-badge
[issues-url]: https://github.com/Juniper/terraform-provider-apstra/issues