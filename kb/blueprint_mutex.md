# Blueprint Mutex

### The problem

Pending changes to an Apstra Blueprint are "staged" in a non-operational clone
of the live Blueprint. When changes are promoted from the staging Blueprint,
they're all sent together. Every detail configured in the staging Blueprint is
promoted at the same time.

This all-or-nothing strategy requires that apstra administrators coordinate
their efforts: Each administrator should endeavor to start with a clean-slate
(an unmodified staging blueprint), make their changes, and complete their
deployment in an expeditious manner so that other administrators aren't left to
contend with half-baked changes, and so nobody inadvertently promotes somebody
else's configuration experiment.

The same holds for non-human Apstra clients like Terraform, but these clients
are more likely to kick off parallel execution workstreams than human users.
So how do we keep two invocations of Terraform run by the "jenkins" Apstra
user from polluting each other's staging blueprints?

### What's A Mutex?

A mutex is any object which signifies "mutual exclusion", or a need for
exclusive access to some resource. You can think of it as a "do not disturb"
sign hanging on a hotel room doorknob. It's trivially bypassed by anyone with a
room key, but the expectation is that anyone with access to the room will honor
the sign placer's desire for exclusive use of the room.

The important features of our mutex are:

1. Everyone can see the mutex.
1. Nobody wonders: Is the mutex mine?
1. The mutex identifies a specific blueprint.
1. The mutex doesn't affect the blueprint.

### Blueprint Mutexes are Apstra Global Catalog Tags

Tags in the global catalog are well positioned to satisfy our requirements:

##### Everyone Can See The Mutex

Automation processes aren't constrained to running on a single system, so an
in-memory or on-filesystem mutex doesn't fit the bill. The mutex needs to live
on a network service accessible to all systems which might attempt writes to a
single blueprint. Perhaps the Apstra API is appropriate?

##### Nobody Wonders: Is The Mutex Mine?

Multiple Tags in the Apstra API cannot share a single name. When a client
attempts to create a tag using an existing name, the Apstra API returns an error
indicating *this tag already exists*. There's no risk of two clients
simultaneously attempting to create the same mutex/tag, and both believing that
they have succeeded, because *one will have received an error*.

##### The Mutex Identifies A Specific Blueprint

The mutex/tags use a [well-known scheme which uniquely identifies a specific blueprint](https://github.com/Juniper/apstra-go-sdk/blob/92821ce72546334b90e4f24342ceca12f33577a7/apstra/two_stage_l3_clos_mutex.go#L12C67-L12C67).

We hope that other automation systems adopt this strategy so that the Terraform
provider won't find itself in conflict with other automation systems using the
Apstra API.

##### The Mutex Doesn't Affect The Blueprint

Because these tags live in the Global Catalog, creating and deleting them does
not revise any blueprint.

### Working with Mutex/Tags in Terraform

##### Terraform provider configuration

The [Terraform provider for Apstra](https://registry.terraform.io/providers/Juniper/apstra/latest/docs)
has two tag-related configuration attributes:

- `blueprint_mutex_enabled` (Boolean)
  - `true` When true, the provider creates a blueprint-specific mutex / tag
  before modifying any Blueprint. If it is unable to create the tag because it
  already exists (the blueprint is locked), the provider will wait until the tag
  is removed (the blueprint is unlocked). There is no timeout. This is setting
  is probably appropriate for a production network environment.
  - `false` When false, the provider neither creates mutex / tags, nor checks if
  one exists before making changes. This setting is reasonable to use in a
  development environment, or anywhere that there is no concern about concurrent
  access by multiple instances of Terraform (or similar automation software).
  - `null` When `null` or omitted (the default), the provider behaves the same
  as the `false` case, but also prints a warning that exclusive access is not
  guaranteed and the user should take steps to understand the risks and then
  explicitly opt in or out of locking.
- `blueprint_mutex_message` (String, Optional) the mutex / tag's `Description`
field is not prescribed, by the locking scheme and [can be used to indicate](https://github.com/Juniper/terraform-provider-apstra/blob/ce37a9c19e62c7709fa7232f0f37d600b17f8e69/apstra/provider.go#L32)
what system or process created the mutex, or other information which might be
useful in the event that it must be manually cleared. Environment variables will
be expanded in the message, so it can include usernames, PIDs, etc...
 
##### Mutex Locking and Unlocking Rules

When locking is enabled, every resource which modifies a resource will ensure
that the blueprint is locked before executing any `Create`, `Update`, or
`Delete` operations. Only the first resource to assert a specific mutex in any
`terraform apply` run *actually causes* the mutex to be created. Subsequent
resources within a single run of `terraform apply` re-use the earlier mutex. It
is not created and then destroyed by each resource in sequence, because doing so
would create an opening for a *different* Terraform to pollute our staging
blueprint.

Mutxes are automatically cleared in exactly two circumstances:
- When the blueprint is deleted by destruction of the
`apstra_datacenter_blueprint` resource
- When the blueprint is committed by the `apstra_blueprint_deploymnet` resource

Practically, these rules mean that automatically asserting and clearing locks is
only possible when the Terraform resources are arranged with appropriate
Terraform lifecycle decorators to ensure the `apstra_blueprint_deploymnet`
resource completes after every other blueprint has completed its changes.

### Manually Clearing Mutexes

When a mutex has been left behind after Terraform exits, either because of a bug,
loss of connectivity to the Apstra API, or a premature exit due to invalid or
impossible resource configuration, no system which honors the mutex will be able
to proceed until the mutex is manually cleared.

In the specific case of a `terraform apply`, this will look like an interminable
`apply` run which never seems to make any progress. In reality, it's regularly
polling the API, waiting for the offending mutex to disappear so that it can
go to work.

It is reasonable and safe to manually clear a mutex any time one has been left
behind so long as it's clear that the mutex's creator has exited without
clearing it.

To manually clear a mutex:
1. Open the Apstra Web Us "Delete" button.
1. Navigate to [Design] -> [Tags]
1. Identify the offending mutex / tag and delete it using its "Delete" button.