## Change to default behavior of `apstra_agent_profile` resource

### Background
Behavior of the [`apstra_agent_profile` resource](https://registry.terraform.io/providers/Juniper/apstra/latest/docs/resources/agent_profile)
changed with version v0.54 of the Terraform Provider for Apstra.

In earlier versions, the default value for the `platform` attribute was `junos`.

Beginning with v0.54, there is no longer any default value for this attribute. Omitting the `platform`
attribute from the Terraform configuration will result in an Agent Profile not tied to any platform.

This change applies to Agent Profiles created with earlier versions of Apstra: If they don't have a `platform`
attribute configured, the Provider will attempt to clear the `platform` attribute after upgrading to v0.54.

### Error
The `Cannot clear Agent Profile 'platform' attribute` error indicates that an attempt to clear the `platform`
attribute from the Agent Profile has failed. If the provider has just been upgraded to v0.54 or later, it is
possible that the configuration was written to depend on the `junos` platform default of an earlier release,
and that clearing the `platform` is not the intended outcome.

The error occurs when the `platform` attribute cannot be cleared from the Agent Profile because all of the
following are true:
- The Agent Profile is used by a Managed Device Agent.
- Tha Managed Device Agent does not have its own `platform` specified.
- The Managed Device associated with the Managed Device Agent has been [acknowledged](https://www.juniper.net/documentation/us/en/software/apstra4.2/apstra-user-guide/topics/task/device-add.html).

### Workaround

Add the `platform = "junos"` [configuration](https://registry.terraform.io/providers/Juniper/apstra/latest/docs/resources/agent_profile#platform)
element to the `apstra_agent_profile` resource to restore the value applied by the default behavior of the
earlier release.

### Workaround (alternate)

Set the `platform` attribute to `Junos` [in the Web UI](https://www.juniper.net/documentation/us/en/software/apstra4.2/apstra-user-guide/topics/task/device-edit.html)
for all Managed Device Agents which reference the Agent Profile.
