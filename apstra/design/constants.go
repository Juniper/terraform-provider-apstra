package design

const (
	AlphaNumericChars  = "[A-Za-z0-9_-]"
	AlphaNumericRegexp = "^" + AlphaNumericChars + "*$"

	errProviderBug           = "Provider Bug. Please report this issue to the provider maintainers."
	errApiGetWithTypeAndId   = "API error getting %s %q"
	errApiPatchWithTypeAndId = "API error patching %s %q"
	errApiParseWithTypeAndId = "failed parsing API response for %s %q"
	errInvalidConfig         = "invalid configuration"

	AsnAllocationSingle = "single"
	AsnAllocationUnique = "unique"

	OverlayControlProtocolEvpn   = "evpn"
	OverlayControlProtocolStatic = "static"

	VlanMin = 1
	VlanMax = 4094

	PoIdMin = 0
	PoIdMax = 4096
)
