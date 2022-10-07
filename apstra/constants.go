package apstra

const (
	errDataSourceConfigureProviderDataSummary = "Unexpected Data Source Configure Type."
	errDataSourceConfigureProviderDataDetail  = "Expected '%T', got: '%T'. Please report this issue to the provider maintainers."
	errResourceConfigureProviderDataSummary   = "Unexpected Resource Configure Type."
	errResourceConfigureProviderDataDetail    = "Expected '%T', got: '%T'. Please report this issue to the provider maintainers."
	errDataSourceUnconfiguredSummary          = "Data Source not configured"
	errDatasourceUnconfiguredDetail           = "Unconfigured data source encountered in Read() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errProviderInvalidConfig                  = "Provider configuration invalid"
	errResourceUnconfiguredSummary            = "Resource not configured"
	errResourceUnconfiguredCreateDetail       = "Unconfigured resource encountered in Create() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errResourceUnconfiguredReadDetail         = "Unconfigured resource encountered in Read() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errResourceUnconfiguredUpdateDetail       = "Unconfigured resource encountered in Update() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errResourceUnconfiguredDeleteDetail       = "Unconfigured resource encountered in Delete() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errReadingAllocation                      = "error reading '%s' resource allocation '%s' for blueprint '%s'"
	errSettingAllocation                      = "error setting '%s' resource allocation '%s' for blueprint '%s'"
	errProviderBug                            = "Provider Bug. Please report this issue to the provider maintainers."
	errInvalidConfig                          = "invalid configuration"
	errTemplateTypeInvalidElement             = "template '%s' has type '%s' which never permits '%s' to be set"

	errDataSourceReadFail         = "Data Source Read() failure'"
	errResourceReadFail           = "Resource Read() failure'"
	errInsufficientConfigElements = "Available configuration elements did provide a solution. Please report this issue to the provider maintainers"
)
