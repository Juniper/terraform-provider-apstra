package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	envTlsKeyLogFile  = "APSTRA_TLS_KEYLOG"
	envApstraUsername = "APSTRA_USER"
	envApstraPassword = "APSTRA_PASS"
	envApstraLogfile  = "APSTRA_LOG"
)

var _ provider.ProviderWithMetadata = &Provider{}

// Provider fulfils the provider.Provider interface
type Provider struct {
	Version    string
	Commit     string
	configured bool
	client     *goapstra.Client
}

// providerData gets instantiated in Provider's Configure() method and
// is made available to the Configure() method of implementations of
// datasource.DataSource and resource.Resource
type providerData struct {
	client *goapstra.Client
}

func (p *Provider) Metadata(_ context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apstra"
	resp.Version = p.Version + "-" + p.Commit
}

func (p *Provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
				MarkdownDescription: "Set 'true' to disable TLS certificate validation.",
			},
		},
	}, diag.Diagnostics{}
}

// Provider configuration struct. Matches GetSchema() output.
type providerConfig struct {
	Url         types.String `tfsdk:"url"`
	TlsNoVerify types.Bool   `tfsdk:"tls_validation_disabled"`
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config providerConfig
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apstraUrl, err := url.Parse(config.Url.Value)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error parsing URL '%s'", config.Url.Value), err.Error())
		return
	}

	clientCfg := &goapstra.ClientCfg{
		Scheme: apstraUrl.Scheme,
		User:   apstraUrl.User.Username(),
		Host:   apstraUrl.Hostname(),
	}

	// parse password from URL, if it exists
	if pass, ok := apstraUrl.User.Password(); ok {
		clientCfg.Pass = pass
	}

	// parse port string from URL, if it exists
	portStr := apstraUrl.Port()
	if portStr == "" {
		switch apstraUrl.Scheme {
		case "http":
			clientCfg.Port = 80
		case "https":
			clientCfg.Port = 443
		default:
			resp.Diagnostics.AddError("cannot guess port number",
				fmt.Sprintf("url scheme is '%s'", apstraUrl.Scheme))
			return
		}
	} else {
		port64, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("error parsing port string", portStr), err.Error())
			return
		}
		clientCfg.Port = uint16(port64)
	}

	// try to fill missing username from environment
	if clientCfg.User == "" {
		if user, ok := os.LookupEnv(envApstraUsername); ok {
			clientCfg.User = user
		} else {
			resp.Diagnostics.AddError(errProviderInvalidConfig, "unable to determine apstra username")
			return
		}
	}

	// try to fill missing password from environment
	if clientCfg.Pass == "" {
		//resp.Diagnostics.AddWarning("yep", "no password")
		if pass, ok := os.LookupEnv(envApstraPassword); ok {
			clientCfg.Pass = pass
		} else {
			resp.Diagnostics.AddError(errProviderInvalidConfig, "unable to determine apstra password")
			return
		}
	}

	// set up logger
	if logFileName, ok := os.LookupEnv(envApstraLogfile); ok {
		logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			resp.Diagnostics.AddError("error opening logfile", err.Error())
			return
		}
		logger := log.New(logFile, "", 0)
		clientCfg.Logger = logger
	}

	// create client's httpClient with the configured TLS verification switch
	clientCfg.HttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.TlsNoVerify.Value,
			}}}

	// TLS key log
	if fileName, ok := os.LookupEnv(envTlsKeyLogFile); ok {
		klw, err := newKeyLogWriter(fileName)
		if err != nil {
			resp.Diagnostics.AddError("error setting up TLS key log", err.Error())
			return
		}
		clientCfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.KeyLogWriter = klw
	}

	// create the goapstra client
	client, err := clientCfg.NewClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"unable to create client",
			fmt.Sprintf("error creating apstra client - %s", err),
		)
		return
	}

	// data passed to Resource and DataSource Configure() methods
	providerData := &providerData{client: client}
	resp.ResourceData = providerData
	resp.DataSourceData = providerData
}

// Resources defines provider resources
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		//func() resource.Resource { return resourceAgentProfile{} },
		//func() resource.Resource { return resourceAsnPool{} },
		//func() resource.Resource { return resourceAsnPoolRange{} },
		//func() resource.Resource { return resourceBlueprint{} },
		//func() resource.Resource { return resourceIp4Pool{} },
		//func() resource.Resource { return resourceIp4PoolSubnet{} },
		//func() resource.Resource { return resourceManagedDevice{} },
		//func() resource.Resource { return &ResourceRackType{} },
		//func() resource.Resource { return resourceSourceTemplateL3Collapsed{} },
		//func() resource.Resource { return resourceSourceTemplatePodBased{} },
		//func() resource.Resource { return resourceSourceTemplateRackBased{} },
		//func() resource.Resource { return resourceWireframe{} },
	}
}

// DataSources defines provider data sources
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		//func() datasource.DataSource { return dataSourceAgentProfile{} },
		//func() datasource.DataSource { return dataSourceAgentProfiles{} },
		//func() datasource.DataSource { return dataSourceAsnPool{} },
		//func() datasource.DataSource { return dataSourceAsnPoolId{} },
		//func() datasource.DataSource { return dataSourceAsnPools{} },
		//func() datasource.DataSource { return dataSourceIp4PoolId{} },
		//func() datasource.DataSource { return dataSourceIp4Pools{} },
		//func() datasource.DataSource { return dataSourceIp4Pool{} },
		//func() datasource.DataSource { return dataSourceLogicalDevice{} },
		//func() datasource.DataSource { return dataSourceRackType{} },
		//func() datasource.DataSource { return dataSourceTemplateL3Collapsed{}},
		//func() datasource.DataSource { return dataSourceTemplatePodBased{}},
		//func() datasource.DataSource { return dataSourceTemplateRackBased{}},
		//func() datasource.DataSource { return dataSourceTag{} },
	}
}
