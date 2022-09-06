# terraform-provider-apstra

This is the beginning of a Terraform provider for Juniper Apstra. It relies on
a Go client library at https://bitbucket.org/apstrktr/goapstra

## Getting Started

1. Get access to both this repo and the client library repo, both of which are
   currently private. You're reading this, so that's a good start.
1. Configure `go` so that it doesn't try to use a public mirror for these repos
   by running the following shell command:
   ```shell
   go env -w GOPRIVATE=bitbucket.org/apstrktr
   ```
1. Add an `ssh` key to your github account.
1. Configure `git` to use `ssh` for these repositories with this shell command:
   ```shell
   git config --global url.git@github.com:.insteadOf https://github.com/
   ```
1. Install [`terraform`](https://www.terraform.io/downloads) and [`Go`](https://go.dev/dl/) on your machine.
1. Clone this repo to your local system:
   ```shell
   git clone git@bitbucket.org/apstrktr/terraform-provider-apstra.git
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
       "example.com/apstrktr/apstra" = "/Users/cmarget/golang/bin"
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
### Data Source: apstra_asn_pool

`apstra_asn_pool` provides details of a specific ASN resource pool by ID.

#### Example Usage
The following example shows outputting a report of free space across all ASN
resource pools:

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

---
### Data Source: apstra_ip4_pool

`apstra_ip4_pool` provides details of a specific IPv4 resource pool by ID.

#### Example Usage
The following example shows outputting a report of free space across all IPv4
resource pools.

```hcl
data "apstra_ip4_pool_ids" "all" {}

data "apstra_ip4_pool" "all" {
   for_each = toset(data.apstra_ip4_pool_ids.all.ids)
   id = each.value
}

output "ipv4_pool_report" {
  value = {for k, v in data.apstra_ip4_pool.all : k => {
    name = v.name
    free = v.total - v.used
  }}
}
```
Result:
```hcl
ipv4_pool_report = {
  "091a4c18-0911-4f78-8db3-c1033b88e08f" = {
    "free" = 1008
    "name" = "leaf-loopback"
  }
  "472b87f2-6174-448c-b3d1-b57c2b48ab58" = {
    "free" = 1022
    "name" = "spine-loopback"
  }
  "Private-10_0_0_0-8" = {
    "free" = 16777216
    "name" = "Private-10.0.0.0/8"
  }
  "Private-172_16_0_0-12" = {
    "free" = 1048576
    "name" = "Private-172.16.0.0/12"
  }
  "Private-192_168_0_0-16" = {
    "free" = 65536
    "name" = "Private-192.168.0.0/16"
  }
  "cc023b60-9941-40a5-a07c-2b15920e544f" = {
    "free" = 992
    "name" = "spine-leaf"
  }
}
```

#### Argument Reference
The following arguments are required:
* `id` (string) ID of the desired IPv4 resource pool.

#### Attributes Reference
In addition to the attributes above, the following attributes are exported:
* `name` (string) The name of the IPv4 resource pool.
* `status` (string) Status of the IPv4 resource pool (string reported by Apstra).
* `tags` (list[string]) Tags applied to the IPv4 resource pool.
* `total` (number) Total number of addresses in the IPv4 resource pool.
* `used` (number) Count of used addresses in the IPv4 resource pool.
* `used_percentage` (number) Percent of used addresses in the IPv4 resource pool.
* `created_at` (string) Creation time.
* `last_modified_at` (string) Last modification time.
* `subnets` (list[object]) Individual IPv4 allocations within the pool, consisting of:
   * `status` (string) Status of the IPv4 resource pool (string reported by Apstra).
   * `network` (string) Network specification in CIDR syntax ("10.0.0.0/8")
   * `total` (number) Total number of addresses in this IPv4 range.
   * `used` (number) Count of used addresses in this IPv4 range
   * `used_percentage` (number) Percent of used addresses in this IPv4 range

---
### Data Source: apstra_ip4_pool_id
`apstra_ip4_pool_id` returns the pool ID of the IPv4 resource pool matching the
supplied criteria. It is incumbent on the user to ensure the criteria matches
exactly one IPv4 pool. Matching zero pools or more than one pool will produce an
error.

#### Example Usage
The following example shows how a module can use the search capability to deploy
a blueprint using a tag-based search for IPv4 pools:

```hcl
variable "site" { default = "ashburn"}
variable "env" {default = "dev"}
variable "template" {}
variable "spine_asn_pool" {}
variable "leaf_asn_pool" {}

data "apstra_ip4_pool_id" "spine" {
   tags = [site, env, "spine"]
}

data "apstra_ip4_pool_id" "leaf" {
   tags = [site, env, "leaf"]
}

data "apstra_ip4_pool_id" "link" {
  tags = [site, env, "link"]
}

resource "apstra_blueprint" "my_blueprint" {
   name               = "my blueprint"
   template_id        = var.bp_template
   spine_asn_pool_ids = spine.asn_pool
   leaf_asn_pool_ids  = leaf.asn_pool
   spine_ip_pool_ids  = [data.apstra_ip4_pool_id.spine.id]
   leaf_ip_pool_ids   = [data.apstra_ip4_pool_id.leaf.id]
   link_ip_pool_ids   = [data.apstra_asn_pool_id.link.id]
}
```

#### Argument Reference
The arguments of this data source act as filters for querying the available
IPv4 resource pools.

The following arguments are optional. At least one must be supplied, and the
complete set of arguments must select exactly one IPv4 resource pool:
* `name` - (Optional) Name of the IPv4 resource pool.
* `tags` - list[string] List of tags applied to IPv4 resource pools. For a pool
  to match, every tag in this list must appear on the IPv4 resource pool. The
  pool may have other tags which do not appear in this list.

#### Attributes Reference
In addition to the attributes above, the following attributes are exported:
* `id` - Apstra ID number of the matching IPv4 resource pool.

---
### Data Source: apstra_ip4_pool_ids
`apstra_ip4_pool_ids` returns the pool IDs all IPv4 resource pools

#### Example Usage
The following example shows outputting all IPv4 pool IDs.

```hcl
data "apstra_ip4_pool_ids" "all" {}

output ip4_pool_ids {
   value = data.apstra_ip4_pool_ids.all.ids
}
```

#### Argument Reference
No arguments.

#### Attributes Reference
* `ids` - list[string] Apstra ID numbers of each IPv4 resource pool.

---
### Data Source: apstra_logical_device

`apstra_logical_device` provides details about a specific logical device (a
template used by apstra when creating rack types -- which are also templates).

This resource looks up details of a logical device using either its name (which
is *not* guaranteed unique), or its ID (but not both). If your environment has
multiple logical devices with the same name, "by name" lookups will produce an
error.

#### Example Usage
The following example shows how a module might accept a logical device name as
an input variable and use it to retrieve the agent profile ID when provisioning
a rack type.

```hcl
variable "logical_device_name" {} 

data "apstra_logical_device" "selected" {
   name = var.logical_device_name
}

resource "apstra_rack_type" "my_rack" {
  todo = "all of this"
}
```

#### Argument Reference
The arguments of this data source act as filters for querying the available
agent profiles.

The following arguments are optional:
* `id` (string) ID of the logical device. Required when `name` is omitted.
* `name` (string) Name of the logical device. Required when `id` is omitted.

#### Attributes Reference
In addition to the attributes above, the following attributes are exported:
* `panels` (list[object]) Detail connectivity features of the logical device.
  * `rows` (string) Detail physical dimension of the panel in the vertical
  direction.
  * `columns` (string) Detail physical dimension of the panel in the horizontal
  direction.
  * `port_groups` (list[object]) Ordered logical groupings of interfaces by
  speed or purpose within a panel.
    * `port_count` (number) Number of ports in the group.
    * `port_speed_gbps` (number) Port speed in Gbps
    * `port_roles` (list[string]) One or more of: `access`, `generic`,
      `l3_server`, `leaf`, `peer`, `server`, `spine`, `superspine` and
      `unused`.
      
---
### Data Source: apstra_tag

`apstra_tag` provides details about a specific tag.

This resource looks up details of a tag using exactly one of its id and its
key(Web Ui: *name*).

#### Example Usage
The following example shows how a module might accept a tag key as an input
variable,then use it to retrieve the appropriate tag when templating devices
within a rack type.

```hcl
variable "tag_key" {} 

data "apstra_tag" "selected" {
   key = var.tag_key
}

resource "apstra_rack_type" "my_rack" {
  todo = "all of this"
}
```

#### Argument Reference
The arguments of this data source act as filters for querying the available
tags.

The following arguments are optional:
* `id` (string) ID of the logical device. Required when `key` is omitted.
* `key` (string) Key (Web Ui: *name*) of the tag. Required when `id` is omitted.

#### Attributes Reference
In addition to the attributes above, the following attributes are exported:
* `value` (string) Value (Web UI: *description*) field of the tag.

---
