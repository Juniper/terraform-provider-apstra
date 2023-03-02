package design

const (
	errProviderBug           = "Provider Bug. Please report this issue to the provider maintainers."
	errApiGetWithTypeAndId   = "API error getting %s %q"
	errApiPatchWithTypeAndId = "API error patching %s %q"
	errInvalidConfig         = "invalid configuration"

	AsnAllocationSingle = "single"
	AsnAllocationUnique = "unique"

	OverlayControlProtocolEvpn   = "evpn"
	OverlayControlProtocolStatic = "static"

	vlanMin = 1
	vlanMax = 4094

	poIdMin = 0
	poIdMax = 4096
)
