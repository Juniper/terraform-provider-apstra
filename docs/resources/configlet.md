---
page_title: "apstra_configlet Resource - terraform-provider-apstra"
subcategory: "Design"
description: |-
  This resource creates a specific Configlet.
---

# apstra_configlet (Resource)

This resource creates a specific Configlet.


## Example Usage

```terraform
# This example creates a configlet responsible for DNS server
# addresses on Junos and EOS devices.
resource "apstra_configlet" "example" {
  name = "DNS according to Terraform"
  generators = [
    {
      config_style  = "junos"
      section       = "top_level_hierarchical"
      template_text = <<-EOT
        name-server {
          4.2.2.1;
          4.2.2.2;
        }
      EOT
    },
    {
      config_style           = "eos"
      section                = "system"
      template_text          = "ip name-server 4.2.2.1 4.2.2.2"
      negation_template_text = "no ip name-server 4.2.2.1 4.2.2.2"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `generators` (Attributes List) Generators organized by Network OS (see [below for nested schema](#nestedatt--generators))
- `name` (String) Configlet name displayed in the Apstra web UI

### Read-Only

- `id` (String) Apstra ID number of Configlet

<a id="nestedatt--generators"></a>
### Nested Schema for `generators`

Required:

- `config_style` (String) Specifies the switch platform, must be one of 'cumulus', 'eos', 'junos', 'nxos', 'sonic'.
- `section` (String) Specifies where in the target device the configlet should be  applied. Varies by network OS:

  | **Config Style**  | **Valid Sections** |
  |---|---|
  |cumulus|file, frr, interface, ospf, system|
  |eos|interface, ospf, system, system_top|
  |junos|interface_level_hierarchical, interface_level_delete, interface_level_set, top_level_hierarchical, top_level_set_delete|
  |nxos|system, interface, system_top, ospf|
  |sonic|file, frr, ospf, system|
- `template_text` (String) Template Text

Optional:

- `filename` (String) FileName
- `negation_template_text` (String) Negation Template Text



