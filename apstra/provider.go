package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var stderr = os.Stderr

const (
	envTlsKeyLogFile  = "APSTRA_TLS_KEYLOG"
	envApstraUsername = "APSTRA_USER"
	envApstraPassword = "APSTRA_PASS"

	dataSourceAgentProfileName    = "apstra_agent_profile"
	dataSourceAgentProfilesName   = "apstra_agent_profiles"
	dataSourceAsnPoolIdName       = "apstra_asn_pool_id"
	dataSourceAsnPoolName         = "apstra_asn_pool"
	dataSourceAsnPoolIdsName      = "apstra_asn_pool_ids"
	dataSourceIp4PoolIdName       = "apstra_ip4_pool_id"
	dataSourceIp4PoolName         = "apstra_ip4_pool"
	dataSourceIp4PoolIdsName      = "apstra_ip4_pool_ids"
	dataSourceLogicalDeviceName   = "apstra_logical_device"
	dataSourceRackTypeName        = "apstra_rack_type"
	dataSourceTemplateL3Collapsed = "apstra_l3collapsed_template"
	//dataSourceTemplatePodBased  = "apstra_podbased_template"
	//dataSourceTemplateRackBased = "apstra_rackbased_template"
	dataSourceTagName = "apstra_tag"

	resourceAgentProfileName  = "apstra_agent_profile"
	resourceAsnPoolName       = "apstra_asn_pool"
	resourceAsnPoolRangeName  = "apstra_asn_pool_range"
	resourceIp4PoolName       = "apstra_ip4_pool"
	resourceIp4PoolSubnetName = "apstra_ip4_pool_subnet"
	resourceBlueprintName     = "apstra_blueprint"
	resourceManagedDeviceName = "apstra_managed_device"
	resourceRackTypeName      = "apstra_rack_type"
	//resourceSourceTemplateL3Collapsed = "apstra_l3collapsed_template"
	//resourceSourceTemplatePodBased    = "apstra_podbased_template"
	//resourceSourceTemplateRackBased   = "apstra_rackbased_template"
	resourceWireframeName = "apstra_template"
)

func New() provider.Provider {
	return &apstraProvider{}
}

type apstraProvider struct {
	configured bool
	client     *goapstra.Client
}

// GetSchema returns provider schema
func (p *apstraProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"url": {
				Type:     types.StringType,
				Required: true,
				MarkdownDescription: "URL of the apstra server, e.g. `http://<user>:<password>@apstra.juniper.net:443/`\n" +
					"If username or password are omitted environment variables `" + envApstraUsername + "` and `" +
					envApstraPassword + "` will be used.",
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
	Url         types.String `tfsdk:"url"`
	TlsNoVerify types.Bool   `tfsdk:"tls_validation_disabled"`
}

func (p *apstraProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// initial client config from URL string
	cfg, err := goapstraClientCfgFromUrlString(config.Url.Value)
	if err != nil {
		resp.Diagnostics.AddError("error parsing apstra url", err.Error())
	}

	// try to fill missing username from environment
	if cfg.User == "" {
		user, userOk := os.LookupEnv(envApstraUsername)
		if !userOk || user == "" {
			resp.Diagnostics.AddError("apstra configuration error", "unable to determine apstra username")
			return
		}
		cfg.User = user
	}

	// try to fill missing password from environment
	if cfg.Pass == "" {
		pass, passOk := os.LookupEnv(envApstraPassword)
		if !passOk || pass == "" {
			resp.Diagnostics.AddError("apstra configuration error", "unable to determine apstra password")
			return
		}
		cfg.Pass = pass
	}

	// set up logger
	logFile, err := os.OpenFile(".terraform.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		resp.Diagnostics.AddError("error opening logfile", err.Error())
	}
	logger := log.New(logFile, "", 0)
	cfg.Logger = logger

	// create client's httpClient with the configured TLS verification switch
	cfg.HttpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: config.TlsNoVerify.Value}}}

	// TLS key log
	if fileName, ok := os.LookupEnv(envTlsKeyLogFile); ok {
		klw, err := newKeyLogWriter(fileName)
		if err != nil {
			resp.Diagnostics.AddError("error setting up TLS key log", err.Error())
		}
		cfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.KeyLogWriter = klw
	}

	// todo: do something with cfg.ErrChan?

	// store the client in the provider
	p.client, err = cfg.NewClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"unable to create client",
			fmt.Sprintf("error creating apstra client - %s", err),
		)
		return
	}

	// set the provider's "configured" flag to indicate it's ready to go
	p.configured = true
}

// GetResources defines provider resources
func (p *apstraProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		resourceAgentProfileName:  resourceAgentProfileType{},
		resourceAsnPoolName:       resourceAsnPoolType{},
		resourceAsnPoolRangeName:  resourceAsnPoolRangeType{},
		resourceIp4PoolName:       resourceIp4PoolType{},
		resourceIp4PoolSubnetName: resourceIp4PoolSubnetType{},
		resourceBlueprintName:     resourceBlueprintType{},
		resourceManagedDeviceName: resourceManagedDeviceType{},
		resourceRackTypeName:      resourceRackTypeType{},
		//resourceSourceTemplateL3Collapsed: resourceSourceTemplateL3CollapsedType{},
		//resourceSourceTemplatePodBased:    resourceSourceTemplatePodBasedType{},
		//resourceSourceTemplateRackBased:   resourceSourceTemplateRackBasedType{},
		resourceWireframeName: resourceWireframeType{},
	}, nil
}

// GetDataSources defines provider data sources
func (p *apstraProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{
		dataSourceAsnPoolIdName:     dataSourceAsnPoolIdType{},
		dataSourceAsnPoolName:       dataSourceAsnPoolType{},
		dataSourceAsnPoolIdsName:    dataSourceAsnPoolsType{},
		dataSourceAgentProfilesName: dataSourceAgentProfilesType{},
		dataSourceAgentProfileName:  dataSourceAgentProfileType{},
		dataSourceIp4PoolIdName:     dataSourceIp4PoolIdType{},
		dataSourceIp4PoolIdsName:    dataSourceIp4PoolsType{},
		dataSourceIp4PoolName:       dataSourceIp4PoolType{},
		dataSourceLogicalDeviceName: dataSourceLogicalDeviceType{},
		//dataSourceRackTypeName:      dataSourceRackTypeType{},
		//dataSourceTemplateL3Collapsed: dataSourceTemplateL3CollapsedType{},
		//dataSourceTemplatePodBased:    dataSourceTemplatePodBasedType{},
		//dataSourceTemplateRackBased:   dataSourceTemplateRackBasedType{},
		dataSourceTagName: dataSourceTagType{},
	}, nil
}

func goapstraClientCfgFromUrlString(urlStr string) (*goapstra.ClientCfg, error) {
	apstraUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.ParseUint(apstraUrl.Port(), 10, 16)
	if err != nil {
		return nil, err
	}
	pass, _ := apstraUrl.User.Password()

	return &goapstra.ClientCfg{
		Scheme: apstraUrl.Scheme,
		User:   apstraUrl.User.Username(),
		Pass:   pass,
		Host:   apstraUrl.Hostname(),
		Port:   uint16(port),
	}, nil
}
