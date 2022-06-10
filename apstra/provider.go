package apstra

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var stderr = os.Stderr

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *goapstra.Client
}

// GetSchema
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"scheme": {
				Type:     types.StringType,
				Optional: false,
				Computed: true,
			},
			"host": {
				Type:     types.StringType,
				Optional: false,
				Computed: true,
			},
			"port": {
				Type:     types.Int64Type,
				Optional: false,
				Computed: true,
			},
			"username": {
				Type:     types.StringType,
				Optional: false,
				Computed: true,
			},
			"password": {
				Type:      types.StringType,
				Optional:  false,
				Computed:  true,
				Sensitive: true,
			},
			"tls_no_verify": {
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	Scheme      types.String `tfsdk:"scheme"`
	Host        types.String `tfsdk:"host"`
	Port        types.Int64  `tfsdk:"port"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	TlsNoVerify types.Bool   `tfsdk:"tls_no_verify"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide a user to the provider
	var username string
	if config.Username.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as username",
		)
		return
	}

	if config.Username.Null {
		username = os.Getenv("HASHICUPS_USERNAME")
	} else {
		username = config.Username.Value
	}

	if username == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find username",
			"Username cannot be an empty string",
		)
		return
	}

	// User must provide a password to the provider
	var password string
	if config.Password.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Cannot use unknown value as password",
		)
		return
	}

	if config.Password.Null {
		password = os.Getenv("HASHICUPS_PASSWORD")
	} else {
		password = config.Password.Value
	}

	if password == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find password",
			"password cannot be an empty string",
		)
		return
	}

	// User must specify a host
	var host string
	if config.Host.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Cannot use unknown value as host",
		)
		return
	}

	if config.Host.Null {
		host = os.Getenv("HASHICUPS_HOST")
	} else {
		host = config.Host.Value
	}

	if host == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find host",
			"Host cannot be an empty string",
		)
		return
	}

	// todo: verify 'port' not out of range for uint16
	// todo: import from environment if unset
	// todo: do something with ClientCfg.ErrChan

	// Create a new goapstra client and set it to the provider client
	c, err := goapstra.NewClient(&goapstra.ClientCfg{
		Scheme:    config.Scheme.Value,
		Host:      config.Host.Value,
		Port:      uint16(config.Port.Value),
		User:      config.Username.Value,
		Pass:      config.Password.Value,
		TlsConfig: &tls.Config{InsecureSkipVerify: config.TlsNoVerify.Value},
		Timeout:   0,
		ErrChan:   nil,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			fmt.Sprintf("error creating apstra client - %s", err),
		)
		return
	}

	p.client = c
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		//"asn_pool": resourceAsnPoolType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"apstra_asn_pools": dataSourceAsnPoolsType{},
	}, nil
}
