# terraform-provider-apstra style guide

---
## Documentation

### Schema

The `MarkdownDescription` element of schema attributes uses markdown as appropriate. In these markdown strings, references to attribute elements (`id`, `name`, `routing_zone_id`, etc...) should use `fixed witdh` and references to Apstra objects (Blueprint, Virtual Network, etc...) should be capitalized.

The `MarkdownDescription` element does not need to enumerate permitted values for cases where only some input values are permitted: (values `evpn` and `static_vxlan` permitted). The validation error should handle that.

### Comments

Comments documenting declarations follow the [godoc conventions](https://github.com/golang/go/wiki/CodeReviewComments#comment-sentences)

### Usage Examples

Usage examples should include both functional HCL and comments explaining what's going on.

Where possible, examples should be directly runnable. This might mean including some prerequisite resources or data lookups. Feedback from users indicates that these prerequisites should be clearly marked because they're not always reading carefully: "I tried running the resource, got something completely different!" (they copy/pasted a prerequisite, not the resource they were after).

When it's not possible/practical to include a directly usable example (too many complicated prerequisites), examples may include placeholder values:

```blueprint_id = "blueprint-xyz"```

These placeholders should NOT be bogus placeholder *references*:

```blueprint_id = apstra_datacenter_blueprint.chicago_dc.id```

Examples don't need to exercise every configuration option, but should demonstrate or call out non-obvious configurations or usages via comments.

When rendered on Hashicorp's registry, the fixed-width usage examples for resources and data sources have an available width of about 76 columns. Wider from that and the users are side-scrolling to see everything. Stick to 76 columns where possible.

### Rosetta

The API and web UI are not a 1:1 match in terms of terminology used. For example, what the web UI and documentation call "Routing Zone" is referred to as "security-zone" in the API.

Broadly speaking, the objective is to make the nomenclature match for a user with 3 windows open on their desktop:

 - An editor with terraform HCL
 - The Apstra web UI
 - The Apstra user manual

We win some latitude with cases like `leaf_loopback_ips` (vs. the GUI's "Loopback IPs - Leafs") because the connection between these two is unmistakable and the input field where `leaf_loopback_ips` is used permits only a handful of values and produces an error enumerating the options when an invalid value is supplied.

Other cases like the `apstra_datacenter_blueprint_deployment` resource are less clear cut. This resource represent a button labeled "Commit" in the web UI, but it needs to be noun-ified to become a terraform resource. We could go with "Commitment", but that sounds funny to me.

The `rosetta` mechanism in the `utils` package is useful for making these translations and should be used even for string types which don't require translation (we are not consistent about this) because it's hard to track down these occurrences when we inevitably need to add translation for a new type.

### Tagged struct elements
Where the user-facing schema strings manifest as `tfsdk` Go struct tags, the tagged elements should match the config schema. This:

```go
type thing struct {
	BlueprintId types.String `tfsdk:"blueprint_id"`
	Name        types.String `tfsdk:"name"`
}
```

Not this:

```go
type thing struct {
    Id    types.String `tfsdk:"blueprint_id"`
    Label types.String `tfsdk:"name"`
}
```

The inevitable translation between API terminology and terraform terminology should be as close to the package boundary as possible: When reading data from or writing data to an SDK object the names might not match. Use consistent names within each repo even if the repos don't agree with each other.

### Diagnostics

Diagnostic Error and Warning messages are automatically prefixed with `Error: ` and `Warning: `. The `summary` field shouldn't start with "error" and should aspire to be concise.

When a diagnostic error is the result of a returned error about which we have little context, the `summary` field is the only place to add context (object IDs, operation type, etc...) and the `detail` field should include only `err.Error()`

In the case that we *do* have context around the problem, don't be shy about including it in the `detail` field. Graph query strings, object IDs, regex strings which failed to find matches, etc... are all appropriate here.

There's no guidance about capitalization, punctuation or leaky abstractions.

Keep in mind that `String()` methods on various framework objects use `%q` for formatting, sometimes in a nested fashion. For this reason, single quotes are preferred when quoting strings within diagnostic messages.

### Errors

Errors should follow [the Go conventions](https://github.com/golang/go/wiki/CodeReviewComments#error-strings) regarding punctuation and capitalization.

Where we can add useful context to an error (generally when an error is received across a package boundary), wrap it with a message for context and a single hyphen:

```go
return fmt.Errorf("while doing xyz to object '%s' - %w", objectId, err)
```

### Framework Describers

Implementations of the `Describer` interface generally return the same value (not markdown) for both their `Description()` and `MarkdownDescription()` methods. These are not surfaced to end users nearly as directly as the other topics in this section, so we have no conventions for style or formatting.

## Code/Style

### Whitespace

For improved readability, insert a single blank line between code blocks which do not have a direct and sequential dependence on one another.

### Function Signatures

Strive for consistent names when the same sort of arguments passed to different functions throughout the codebase:

- A `context.Context` is always passed as `ctx`
- The framework's request object and response pointer are always passed as `req` and `resp`

We generally return "by reference" types (pointers, slices, maps) rather than objects so that error cases can be handled with `nil`, rather than instantiating a dummy object to return. This is a conscious choice! It has implications for garbage collection because data winds up in the heap rather than being confined to the stack frame. It's fine. The provider spends the majority of its time waiting for the API. Garbage collection and performance in general are not prioritized over readability.

String types are the exception to the above rule. Returning strings directly, is fine. Use empty strings in error cases.

Most methods on structs are implemented as pointer receivers to avoid creating an unnecessary copy of the object (a nearly pointless optimization, I'm sure).

Some methods *must* use pointer receiver style because they update fields within the target of the pointer.

Other methods use a value receiver because it's super-convenient to have access to a method on an anonymous instantiation of a struct of the given type.

Mixed use of a value receivers and pointer receivers is [frowned-upon by the Go FAQ](https://go.dev/doc/faq#methods_on_values_or_pointers) for reasons relating to method set consistency. In our case (many invocations of methods on anonymous struct instances), convenience and readability win, so we choose to mix method receivers where necessary.

### Variable Naming

Variables should be instantiated as close to the place where they're used, and in the narrowest scope (smallest nested code block) possible. Despite the preceding, it is not necessary to redeclare reusable variables (e.g. `err`) within each code block.

Avoid shadowing prior declarations of a variable when practical to do so. Multiple assignment situations using the walrus operator frequently lead to inadvertent shadowing. There's a tradeoff for readability here. Use good judgement.

Variable names comply with go conventions, except in the case of [initialisms](https://github.com/golang/go/wiki/CodeReviewComments#initialisms) where Chris plays the benevolent dictator card. We use `TlsNoVerify` not `TLSNoVerify`.

Variable names should be descriptive, and longish names are okay. In the case where a long name appears multiple times on a line (say, a map lookup and replace operation), consider introduce an intermediate variable rather than shortening the variable name.

Short variable names become acceptable when:

- The scope of the variable is small. A single character variable used only on two consecutive lines is fine. The smaller the scope, the shorter the name.
- The variable is an instance of an oft-repeated pattern (iterator variable `i`, transient diagnostic stack `d`)

### Package Export

Structs, methods, functions, constants, etc...: Make the private when you can, public when you must. It's much easier recognise the need to go back and expose a private method than it is to do the opposite.

### Third party libraries

Any new inclusion of a third party library requires a team discussion. This includes indirect inclusions which come as a result of an bumping to a new upstream version of a current dependency.

Any license change on a third party library requires a team discussion.

## Git/GitHub operations

### Pull requests
The title of each pull request becomes a bullet item in the auto-generated release notes and is ultimately end-user facing.

Each PR should reference an issue which led to the work being done. Working on a new feature? Open an issue which requests it.

Generally speaking, PR titles should be verbs that cite the issue which led to their creation.
- Add support for thing
- Fix bug 123
- Complete task 456

### Issues
Issues titles created by contributors or created with the intention of leading to changes in the codebase should generally start with an tag-style identifier: bug, feature, proposal, task...

Issues created by users will of course be of the form "how do i?" or "xyz not working" :)

There's no such thing as too many issues. Spot something weird while working on something else? Log an issue so we can come back to it. If the current workstream isn't disrupted by the weird thing, a new issue is preferred over implementing a fix in an unrelated area of the codebase.

### Commit messages
Anything goes, try to make 'em useful to next-week-you.

### Branch names
Branch names should take the form `bug/123`, `task/456` or `feat/789` identifying the issue which led to the work.

### Code review
We use the github review/conversation method for review and requesting changes. We need to get better at this. To that end, conversations should generally be "resolved" by the initiator to reduce misunderstandings.

### CI workflows
We should strive for new contributions to pass the various automated checks without intervention. There [are cases](https://github.com/Juniper/terraform-provider-apstra/pull/128#pullrequestreview-1448419079) where disabling a linter check is the better choice. These should always be called out for discussion.


