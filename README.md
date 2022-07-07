# terraform-provider-apstra

This is the beginning of a Terraform provider for Juniper Apstra. It relies on
a Go client library at https://github.com/chrismarget-j/goapstra

## Getting Started

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
   git clone git@github.com:chrismarget-j/terraform-provider-apstra.git
   ```
1. Build the provider:
   ```shell
   cd terraform-provider-apstra
   go install
   ```
   The provider should now be located at `~/golang/bin/terraform-provider-apstra`
1. Configure terraform to use the local copy of apstra provider, rather than
   attempting to fetch it from a terraform registry. My `~/.terraformrc` looks
   like:
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
## Data Sources

### Data Source: apstra_agent_profile

`apstra_agent_profile` provides details about a specific agent profile.

This resource looks up details of agent profiles using either its name (Apstra
ensures these are unique), or its ID (but not both).

#### Example Usage
The following example shows how a module might accept an agent profile's name as
an input variable and use it to retrieve the agent profile ID when provisioning
a managed device (a switch).

```hcl
variable "agent_profile_name" {} 
variable "switch_mgmt_ip" {}

data "apstra_agent_profile" "selected" {
   name = var.agent_profile_name
}

resource "apstra_managed_device" "switch" {
   agent_profile_id = data.apstra_agent_profile.selected.id
   management_ip    = var.switch_mgmt_ip
}
```

#### Argument Reference
The arguments of this data source act as filters for querying the available
agent profiles.

The following arguments are optional:
* `id` (string) ID of the agent profile. Required when `name` is omitted.
* `name` (string) Name of the agent profile. Required when `id` is omitted.

#### Attributes Reference
In addition to the attributes above, the following attributes are exported:
* `platform` (string) Indicates the platform supported by the agent profile.
* `has_username` (bool) Indicates whether a username has been configured.
* `has_password` (bool) Indicates whether a password has been configured.
* `packages` (map[package]version) Admin-provided software packages stored on the Apstra server
  applied to devices using the profile.
* `open_options` (map[key]value) Configured parameters for offbox agents.

---
### Data Source: apstra_agent_profiles

`apstra_agent_profiles` provides a list of all agent profile IDs.

#### Example Usage
The following example shows outputting all agent profile IDs.

```hcl
data "apstra_agent_profiles" "all" {}

output "agent_profiles" {
   value = data.apstra_agent_profiles.all
}
```

#### Argument Reference
No arguments.

#### Attributes Reference
* `ids` (list[string]) Apstra ID numbers of each agent profile

---
### Data Source: apstra_asn_pool

`apstra_asn_pool` provides details of a specific ASN pool by ID.

#### Example Usage
The following example shows outputting a report of free space across all ASN
pools:

```hcl
data "apstra_asn_pool_ids" "all" {}

data "apstra_asn_pool" "all" {
   for_each = toset(data.apstra_asn_pool_ids.all.ids)
   id = each.value
}

output "asn_report" {
  value = {for k, v in data.apstra_asn_pool.all : k => {
    name = v.name
    free = v.total - v.used
  }}
}
```
Result:
```hcl
asn_report = {
  "3ddb7a6a-4c84-458f-8632-705764c0f6ca" = {
    "free" = 100
    "name" = "spine"
  }
  "Private-4200000000-4294967294" = {
    "free" = 94967293
    "name" = "Private-4200000000-4294967294"
  }
  "Private-64512-65534" = {
    "free" = 1020
    "name" = "Private-64512-65534"
  }
  "dd0d3b45-2020-4382-9c01-c43e7d474546" = {
    "free" = 10002
    "name" = "leaf"
  }
}
```

#### Argument Reference
The following arguments are required:
* `id` (string) ID of the desired ASN resource pool.

#### Attributes Reference
In addition to the attributes above, the following attributes are exported:
* `name` (string) The name of the ASN resource pool.
* `status` (string) Status of the ASN resource pool (string reported by Apstra).
* `tags` (list[string]) Tags applied to the ASN resource pool.
* `total` (number) Total number of ASNs in the ASN resource pool.
* `used` (number) Count of used ASNs in the ASN resource pool.
* `used_percentage` (number) Percent of used ASNs in the ASN resource pool.
* `created_at` (string) Creation time.
* `last_modified_at` (string) Last modification time.
* `ranges` (list[object]) Individual ASN ranges within the pool, consisting of:
  * `status` (string) Status of the ASN resource pool (string reported by Apstra).
  * `first` (number) Lowest numbered AS in this ASN range.
  * `last` (number) Highest numbered AS in this ASN range.
  * `total` (number) Total number of ASNs in this ASN range.
  * `used` (number) Count of used ASNs in this ASN range
  * `used_percentage` (number) Percent of used ASNs in this ASN range

---
### Data Source: apstra_asn_pool_id
`apstra_asn_pool_id` returns the pool ID of the ASN resource pool matching the
supplied criteria. It is incumbent on the user to ensure the criteria matches
exactly one ASN pool. Matching zero pools or more than one pool will produce an
error.

#### Example Usage
The following example shows how a module can use the search capability to deploy
a blueprint using a tag-based search for ASN pools:

```hcl
variable "site" { default = "ashburn"}
variable "env" {default = "dev"}
variable "template" {}
variable "spine_ip_pool" {}
variable "leaf_ip_pool" {}
variable "link_ip_pool" {}

data "apstra_asn_pool_id" "spine" {
   tags = [site, env, "spine"]
}

data "apstra_asn_pool_id" "leaf" {
   tags = [site, env, "leaf"]
}

resource "apstra_blueprint" "my_blueprint" {
   name               = "my blueprint"
   template_id        = var.bp_template
   spine_asn_pool_ids = [data.apstra_asn_pool_id.spine.id]
   leaf_asn_pool_ids  = [data.apstra_asn_pool_id.leaf.id]
   spine_ip_pool_ids  = var.spine_ip_pool
   leaf_ip_pool_ids   = var.leaf_ip_pool
   link_ip_pool_ids   = var.link_ip_pool
}
```

#### Argument Reference
The arguments of this data source act as filters for querying the available
ASN resource pools.

The following arguments are optional. At least one must be supplied, and the
complete set of arguments must select exactly one ASN resource pool:
* `name` - (Optional) Name of the ASN resource pool.
* `tags` - list[string] List of tags applied to ASN resource pools. For a pool
to match, every tag in this list must appear on the ASN resource pool. The pool
may have other tags which do not appear in this list.

#### Attributes Reference
In addition to the attributes above, the following attributes are exported:
* `id` - Apstra ID number of the matching ASN resource pool.

---
### Data Source: apstra_asn_pool_ids
`apstra_asn_pool_ids` returns the pool IDs all ASN resource pools

#### Example Usage
The following example shows outputting all ASN pool IDs.

```hcl
data "apstra_asn_pool_ids" "all" {}

output asn_pool_ids {
   value = data.apstra_asn_pool_ids.all.ids
}
```

#### Argument Reference
No arguments.

#### Attributes Reference
* `ids` - list[string] Apstra ID numbers of each ASN resource pool.
