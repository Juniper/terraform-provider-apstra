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

	VlanMin = 1
	VlanMax = 4094

	poIdMin = 0
	poIdMax = 4096
)
