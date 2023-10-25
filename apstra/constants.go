package tfapstra

const (
	errApiCompatibility                       = "Apstra API version incompatibility"
	errApiData                                = "API produced unexpected result"
	errApiError                               = "API response included an error"
	errDataSourceConfigureProviderDataSummary = "Unexpected Data Source Configure Type."
	errDataSourceConfigureProviderDataDetail  = "Expected '%T', got: '%T'. Please report this issue to the provider maintainers."
	errResourceConfigureProviderDataSummary   = "Unexpected Resource Configure Type."
	errResourceConfigureProviderDataDetail    = "Expected '%T', got: '%T'. Please report this issue to the provider maintainers."
	errProviderInvalidConfig                  = "Provider configuration invalid"
	errReadingAllocation                      = "error reading '%s' resource allocation '%s' for Blueprint '%s'"
	errSettingAllocation                      = "error setting resource allocation"
	errProviderBug                            = "Provider Bug. Please report this issue to the provider maintainers."
	errInvalidConfig                          = "invalid configuration"
	errTemplateTypeInvalidElement             = "template '%s' has type '%s' which never permits '%s' to be set"
	errDataSourceReadFail                     = "Data Source Read() failure'"
	errResourceReadFail                       = "Resource Read() failure'"
)
