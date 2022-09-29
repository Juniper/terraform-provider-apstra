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
)

const (
	DefaultVersion = "0.0.0"
	DefaultCommit  = "devel"

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
	resp.Version = p.Version + "_" + p.Commit
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

	parsedUrl, err := url.Parse(config.Url.Value)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error parsing URL '%s'", config.Url.Value), err.Error())
		return
	}

	// determine apstra username
	user := parsedUrl.User.Username()
	if user == "" {
		if val, ok := os.LookupEnv(envApstraUsername); ok {
			user = val
		} else {
			resp.Diagnostics.AddError(errProviderInvalidConfig, "unable to determine apstra username")
			return
		}
	}

	// determine apstra password
	pass, found := parsedUrl.User.Password()
	if !found {
		if val, ok := os.LookupEnv(envApstraPassword); ok {
			pass = val
		} else {
			resp.Diagnostics.AddError(errProviderInvalidConfig, "unable to determine apstra password")
			return
		}
	}

	// remove credentials from URL prior to rendering it into ClientCfg
	parsedUrl.User = nil

	clientCfg := &goapstra.ClientCfg{
		Url:  parsedUrl.String(),
		User: user,
		Pass: pass,
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

	// login after creation so that future parallel workflows don't trigger TOO MANY REQUESTS threshold
	err = client.Login(ctx)
	if err != nil {
		resp.Diagnostics.AddError("apstra login failure", err.Error())
		return
	}

	// data passed to Resource and DataSource Configure() methods
	pd := &providerData{client: client}
	resp.ResourceData = pd
	resp.DataSourceData = pd
}

// DataSources defines provider data sources
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &dataSourceAgentProfile{} },
		func() datasource.DataSource { return &dataSourceAgentProfiles{} },
		func() datasource.DataSource { return &dataSourceAsnPool{} },
		func() datasource.DataSource { return &dataSourceAsnPools{} },
		func() datasource.DataSource { return &dataSourceIp4Pools{} },
		func() datasource.DataSource { return &dataSourceIp4Pool{} },
		func() datasource.DataSource { return &dataSourceLogicalDevice{} },
		//func() datasource.DataSource { return &dataSourceRackType{} },
		//func() datasource.DataSource { return &dataSourceTemplateL3Collapsed{}},
		//func() datasource.DataSource { return &dataSourceTemplatePodBased{}},
		//func() datasource.DataSource { return &dataSourceTemplateRackBased{}},
		//func() datasource.DataSource { return &dataSourceTag{} },
	}
}

// Resources defines provider resources
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		//func() resource.Resource { return &resourceAgentProfile{} },
		//func() resource.Resource { return &resourceAsnPool{} },
		//func() resource.Resource { return &resourceAsnPoolRange{} },
		//func() resource.Resource { return &resourceBlueprint{} },
		//func() resource.Resource { return &resourceIp4Pool{} },
		//func() resource.Resource { return &resourceIp4PoolSubnet{} },
		//func() resource.Resource { return &resourceManagedDevice{} },
		//func() resource.Resource { return &ResourceRackType{} },
		//func() resource.Resource { return &resourceSourceTemplateL3Collapsed{} },
		//func() resource.Resource { return &resourceSourceTemplatePodBased{} },
		//func() resource.Resource { return &resourceSourceTemplateRackBased{} },
		//func() resource.Resource { return &resourceWireframe{} },
	}
}
