---
page_title: "apstra_managed_device_ack Resource - terraform-provider-apstra"
subcategory: ""
description: |-
  This resource acknowledges the System (probably a switch) discovered by a running System Agent. The acknowledgement of a System cannot be modified nor deleted. Any modification to the inputs of this resource will cause it to be removed from the Terraform state and recreated. Modifying or deleting this resource has no effect on Apstra.
---

# apstra_managed_device_ack (Resource)

This resource *acknowledges* the System (probably a switch) discovered by a running System Agent. The acknowledgement of a System cannot be modified nor deleted. Any modification to the inputs of this resource will cause it to be removed from the Terraform state and recreated. Modifying or deleting this resource has no effect on Apstra.

## Example Usage

```terraform
# This example "acknowledges" a managed device after checking that the
# discovered serial number (device_key) matches the supplied value. The
# only managed device checked is the one managed by the specified System
# Agent.

resource "apstra_managed_device_ack" "spine1" {
  agent_id = "d4689206-fe4b-4e39-9a96-3a398f319660"
  device_key = "525400BB9B5F"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `agent_id` (String) Apstra ID for the System Agent responsible for the Managed Device.
- `device_key` (String) Key which uniquely identifies a System asset, probably the serial number.

### Read-Only

- `system_id` (String) Apstra ID for the System discovered by the System Agent.