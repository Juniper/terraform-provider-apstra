package constants

const (
	ErrApiGetWithTypeAndId   = "API error getting %s %q"
	ErrApiPatchWithTypeAndId = "API error patching %s %q"
	ErrProviderBug           = "Provider Bug. Please report this issue to the provider maintainers."
	ErrInvalidConfig         = "invalid configuration"
	ErrStringParse           = "failed to parse string value"

	L3MtuMin = 1280
	L3MtuMax = 9216
)
