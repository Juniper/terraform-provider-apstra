package apstra

const (
	errDataSourceConfigureProviderDataSummary = "Unexpected Data Source Configure Type."
	errDataSourceConfigureProviderDataDetail  = "Expected '%T', got: '%T'. Please report this issue to the provider developers."
	errReourceConfigureProviderDataSummary    = "Unexpected Resource Configure Type."
	errReourceConfigureProviderDataDetail     = "Expected '%T', got: '%T'. Please report this issue to the provider developers."
	errDataSourceUnonfiguredSummary           = "Data Source not configured"
	errDatasourceUnonfiguredCreateDetail      = "Unconfigured data source encountered in Create() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errDatasourceUnonfiguredReadDetail        = "Unconfigured data source encountered in Read() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errDatasourceUnonfiguredUpdateDetail      = "Unconfigured data source encountered in Update() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errDatasourceUnonfiguredDeleteDetail      = "Unconfigured data source encountered in Delete() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errProviderInvalidConfig                  = "Provider configuration invalid"
	errResourceUnonfiguredSummary             = "Resource not configured"
	errResourceUnonfiguredCreateDetail        = "Unconfigured resource encountered in Create() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errResourceUnonfiguredReadDetail          = "Unconfigured resource encountered in Read() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errResourceUnonfiguredUpdateDetail        = "Unconfigured resource encountered in Update() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
	errResourceUnonfiguredDeleteDetail        = "Unconfigured resource encountered in Delete() operation, possibly because it depends on an unknown value from another object. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!"
)
