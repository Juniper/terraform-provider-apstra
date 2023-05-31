package blueprint

const (
	errApiGetWithTypeAndId   = "API error getting %s %q"
	errApiPatchWithTypeAndId = "API error patching %s %q"
	errProviderBug           = "Provider Bug. Please report this issue to the provider maintainers."

	ErrDCBlueprintCreate = "Failed to create client for Datacenter Blueprint %s"

	vrfIdMin = 1
	vrfIdMax = 4999
)
