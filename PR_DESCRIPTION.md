# Add Support for SVI IP Management in Virtual Networks

## Description

This PR adds support for managing SVI IPs (Secondary Virtual Interface IPs) in the `apstra_datacenter_virtual_network` resource. This feature allows users to explicitly assign specific IP addresses to leaf switches in a virtual network, which is critical for multi-datacenter deployments to prevent overlapping IP assignments.

## Problem Addressed

Currently, when creating identical virtual networks in separate datacenter blueprints (a common scenario for multi-datacenter deployments), Apstra automatically assigns SVI IPs from available pools. This can lead to the same IP being assigned to leaf switches in different datacenters, causing operational issues.

While the Go SDK already fully supports specifying SVI IPs, the Terraform provider does not expose this functionality. This PR fills that gap.

## Changes

1. Added a new `svi_ips` field to the `apstra_datacenter_virtual_network` resource schema
2. Implemented SviIp struct with serialization/deserialization support
3. Updated resource handlers to process SVI IP settings for virtual networks
4. Added documentation and examples for the new feature
5. Added tests for the SVI IP functionality

## Examples

```hcl
resource "apstra_datacenter_virtual_network" "example" {
  name                       = "example-vn"
  blueprint_id               = apstra_datacenter_blueprint.example.id
  type                       = "vxlan"
  routing_zone_id            = apstra_datacenter_routing_zone.example.id
  ipv4_connectivity_enabled  = true
  ipv4_subnet                = "10.0.0.0/24"
  
  # Explicitly assign SVI IPs to leaf switches
  svi_ips {
    system_id     = "leaf1-system-id"
    ipv4_address  = "10.0.0.2/24"
    ipv4_mode     = "enabled"
  }
  
  svi_ips {
    system_id     = "leaf2-system-id"
    ipv4_address  = "10.0.0.3/24"
    ipv4_mode     = "enabled"
  }

  bindings = {
    "leaf1-system-id" = {
      access_ids = []
      vlan_id    = 100
    }
    "leaf2-system-id" = {
      access_ids = []
      vlan_id    = 100
    }
  }
}
```

## Testing

- Added unit tests for SVI IP struct serialization and deserialization
- Tested with Apstra server to ensure proper API interaction
- Verified backward compatibility with existing virtual network resources

## Related Issues

This feature addresses the need for precise control over SVI IP assignments across multiple datacenter blueprints, which is necessary for preventing IP address conflicts when managing multi-datacenter deployments with Terraform.