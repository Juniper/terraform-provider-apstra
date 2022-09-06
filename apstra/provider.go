package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"math"
	"net/http"
	"os"
)

var stderr = os.Stderr

const (
	envTlsKeyLogFile = "APSTRA_API_TLS_LOGFILE"

	dataSourceAgentProfileName  = "apstra_agent_profile"
	dataSourceAgentProfilesName = "apstra_agent_profiles"
	dataSourceAsnPoolIdName     = "apstra_asn_pool_id"
	dataSourceAsnPoolName       = "apstra_asn_pool"
	dataSourceAsnPoolIdsName    = "apstra_asn_pool_ids"
	dataSourceIp4PoolIdName     = "apstra_ip4_pool_id"
	dataSourceIp4PoolName       = "apstra_ip4_pool"
	dataSourceIp4PoolIdsName    = "apstra_ip4_pool_ids"
	dataSourceLogicalDeviceName = "apstra_logical_device"
	dataSourceTagName           = "apstra_tag"

	resourceAgentProfileName  = "apstra_agent_profile"
	resourceAsnPoolName       = "apstra_asn_pool"
	resourceAsnPoolRangeName  = "apstra_asn_pool_range"
	resourceIp4PoolName       = "apstra_ip4_pool"
	resourceIp4PoolSubnetName = "apstra_ip4_pool_subnet"
	resourceBlueprintName     = "apstra_blueprint"
	resourceManagedDeviceName = "apstra_managed_device"
	resourceRackTypeName      = "apstra_rack_type"
	resourceWireframeName     = "apstra_template"
)

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *goapstra.Client
}

// GetSchema returns provider schema
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"scheme": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "URL Scheme used to connect to Apstra, default value is 'https'.",
			},
			"host": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Hostname or IP address of the Apstra API server.",
			},
			"port": {
				Type:                types.Int64Type,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "TCP port number of the Apstra API listener.",
			},
			"username": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Apstra API username.",
			},
			"password": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Apstra API password.",
			},
			"tls_validation_disabled": {
				Type:                types.BoolType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set 'true' to disable TLS certificate validation.",
			},
		},
	}, diag.Diagnostics{}
}

// Provider schema struct
type providerData struct {
	Scheme      types.String `tfsdk:"scheme"`
	Host        types.String `tfsdk:"host"`
	Port        types.Int64  `tfsdk:"port"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	TlsNoVerify types.Bool   `tfsdk:"tls_validation_disabled"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch {
	case config.Username.Unknown:
		resp.Diagnostics.AddWarning("Unable to create client", "Cannot use unknown value as 'Username'")
		return
	case config.Password.Unknown:
		resp.Diagnostics.AddWarning("Unable to create client", "Cannot use unknown value as 'Password'")
		return
	case config.Host.Unknown:
		resp.Diagnostics.AddWarning("Unable to create client", "Cannot use unknown value as 'Host'")
		return
	}

	if config.Port.Value < 0 || config.Port.Value > math.MaxUint16 {
		resp.Diagnostics.AddError(
			"invalid port",
			fmt.Sprintf("'Port' %d out of range", config.Port.Value))
	}

	// todo: do something with ClientCfg.ErrChan?
	logFile, err := os.OpenFile(".terraform.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		resp.Diagnostics.AddError("error opening logfile", err.Error())
	}
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(logFile, "", 0)
	klw, err := keyLogWriterFromEnv(envTlsKeyLogFile)
	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{
		InsecureSkipVerify: config.TlsNoVerify.Value,
		KeyLogWriter:       klw,
	}}}

	// Create a new goapstra client and set it to the provider client
	c, err := goapstra.ClientCfg{
		Scheme:     config.Scheme.Value,
		User:       config.Username.Value,
		Pass:       config.Password.Value,
		Host:       config.Host.Value,
		Port:       uint16(config.Port.Value),
		HttpClient: httpClient,
		Logger:     logger,
		Timeout:    0,
		ErrChan:    nil,
	}.NewClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"unable to create client",
			fmt.Sprintf("error creating apstra client - %s", err),
		)
		return
	}

	p.client = c
	p.configured = true
}

// GetResources defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		resourceAgentProfileName:  resourceAgentProfileType{},
		resourceAsnPoolName:       resourceAsnPoolType{},
		resourceAsnPoolRangeName:  resourceAsnPoolRangeType{},
		resourceIp4PoolName:       resourceIp4PoolType{},
		resourceIp4PoolSubnetName: resourceIp4PoolSubnetType{},
		resourceBlueprintName:     resourceBlueprintType{},
		resourceManagedDeviceName: resourceManagedDeviceType{},
		resourceRackTypeName:      resourceRackTypeType{},
		resourceWireframeName:     resourceWireframeType{},
	}, nil
}

// GetDataSources defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		dataSourceAsnPoolIdName:     dataSourceAsnPoolIdType{},
		dataSourceAsnPoolName:       dataSourceAsnPoolType{},
		dataSourceAsnPoolIdsName:    dataSourceAsnPoolsType{},
		dataSourceAgentProfilesName: dataSourceAgentProfilesType{},
		dataSourceAgentProfileName:  dataSourceAgentProfileType{},
		dataSourceIp4PoolIdName:     dataSourceIp4PoolIdType{},
		dataSourceIp4PoolIdsName:    dataSourceIp4PoolsType{},
		dataSourceIp4PoolName:       dataSourceIp4PoolType{},
		dataSourceLogicalDeviceName: dataSourceLogicalDeviceType{},
		dataSourceTagName:           dataSourceTagType{},
	}, nil
}
