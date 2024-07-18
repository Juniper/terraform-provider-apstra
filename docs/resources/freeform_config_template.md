---
page_title: "apstra_freeform_config_template Resource - terraform-provider-apstra"
subcategory: "Reference Design: Freeform"
description: |-
  This resource creates a Config Template in a Freeform Blueprint.
---

# apstra_freeform_config_template (Resource)

This resource creates a Config Template in a Freeform Blueprint.


## Example Usage

```terraform
# This example creates a  Config Template from a local jinja file.
locals {
  template_filename = "interfaces.jinja"
}

resource "apstra_freeform_config_template" "test" {
  blueprint_id = "043c5787-66e8-41c7-8925-c7e52fbe6e32"
  name         = local.template_filename
  tags         = ["a", "b", "c"]
  text         = file("${path.module}/${local.template_filename}")
}

/*
# Contents of the interfaces.jinja file in this directory follows here:
{% set this_router=hostname %}
interfaces {
{% for interface_name, iface in interfaces.items() %}
    replace: {{ interface_name }} {
        unit 0 {
            description "{{iface['description']}}";
    {% if iface['ipv4_address'] and iface['ipv4_prefixlen'] %}
            family inet {
                address {{iface['ipv4_address']}}/{{iface['ipv4_prefixlen']}};
            }
    {% endif %}
        }
    }
{% endfor %}
    replace: lo0 {
        unit 0 {
            family inet {
                address {{ property_sets.data[this_router | replace('-', '_')]['loopback'] }}/32;
            }
        }
    }
}
*/
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `blueprint_id` (String) Apstra Blueprint ID.
- `name` (String) Config Template name as shown in the Web UI. Must end with `.jinja`.
- `text` (String) Configuration Jinja2 template text

### Optional

- `tags` (Set of String) Set of Tag labels

### Read-Only

- `id` (String) ID of the Config Template.


