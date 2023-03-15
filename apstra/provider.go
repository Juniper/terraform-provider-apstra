package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	DefaultVersion = "0.0.0"
	DefaultCommit  = "devel"

	envTlsKeyLogFile  = "SSLKEYLOGFILE"
	envApstraUsername = "APSTRA_USER"
	envApstraPassword = "APSTRA_PASS"
	envApstraLogfile  = "APSTRA_LOG"
	envApstraUrl      = "APSTRA_URL"
)

var _ provider.Provider = &Provider{}

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
	client           *goapstra.Client
	providerVersion  string
	terraformVersion string
	mutexes          *[]goapstra.TwoStageL3ClosMutex
}

func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apstra"
	resp.Version = p.Version + "_" + p.Commit
}

func (p *Provider) Schema(_ context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "URL of the apstra server, e.g. `https://<user>:<password>@apstra.juniper.net:443/`\n" +
					"If username or password are omitted from URL string, environment variables `" + envApstraUsername +
					"` and `" + envApstraPassword + "` will be used.  If `url` is omitted, environment variable " +
					envApstraUrl + " will be used.",
			},
			"tls_validation_disabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set 'true' to disable TLS certificate validation.",
			},
			"blueprint_mutex_disabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set 'true' to skip locking the mutex(es) which signal exclusive blueprint access",
			},
		},
	}
}

// Provider configuration struct. Matches GetSchema() output.
type providerConfig struct {
	Url          types.String `tfsdk:"url"`
	TlsNoVerify  types.Bool   `tfsdk:"tls_validation_disabled"`
	MutexDisable types.Bool   `tfsdk:"blueprint_mutex_disabled"`
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config providerConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// populate raw URL string from config or environment
	var apstraUrl string
	var ok bool
	if config.Url.IsNull() {
		if apstraUrl, ok = os.LookupEnv(envApstraUrl); !ok {
			resp.Diagnostics.AddError(errInvalidConfig, "missing Apstra URL")
			return
		}
	} else {
		apstraUrl = config.Url.ValueString()
	}

	// either config or env could have sent us an empty string
	if apstraUrl == "" {
		resp.Diagnostics.AddError(errInvalidConfig, "Apstra URL: empty string")
		return
	}

	// parse the URL
	parsedUrl, err := url.Parse(apstraUrl)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error parsing URL '%s'", config.Url.ValueString()), err.Error())
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
				InsecureSkipVerify: config.TlsNoVerify.ValueBool(),
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

	var version string
	if p.Commit == "" {
		version = p.Version
	} else {
		version = p.Version + "-" + p.Commit
	}

	var mutexes []goapstra.TwoStageL3ClosMutex
	if !config.MutexDisable.ValueBool() {
		// non-nil slice signals resources to lock and log mutexes
		mutexes = make([]goapstra.TwoStageL3ClosMutex, 0)
	}

	// data passed to Resource and DataSource Configure() methods
	pd := &providerData{
		client:           client,
		providerVersion:  version,
		terraformVersion: req.TerraformVersion,
		mutexes:          &mutexes,
	}
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
		func() datasource.DataSource { return &dataSourceBlueprintDeploy{} },
		func() datasource.DataSource { return &dataSourceBlueprints{} },
		func() datasource.DataSource { return &dataSourceConfiglet{} },
		func() datasource.DataSource { return &dataSourceConfiglets{} },
		func() datasource.DataSource { return &dataSourceInterfaceMap{} },
		func() datasource.DataSource { return &dataSourceInterfaceMaps{} },
		func() datasource.DataSource { return &dataSourceIpv4Pool{} },
		func() datasource.DataSource { return &dataSourceIpv4Pools{} },
		func() datasource.DataSource { return &dataSourceIpv6Pool{} },
		func() datasource.DataSource { return &dataSourceIpv6Pools{} },
		func() datasource.DataSource { return &dataSourceLogicalDevice{} },
		func() datasource.DataSource { return &dataSourceRackType{} },
		func() datasource.DataSource { return &dataSourceRackTypes{} },
		//func() datasource.DataSource { return &dataSourceTemplateL3Collapsed{} },
		////func() datasource.DataSource { return &dataSourceTemplatePodBased{}},
		func() datasource.DataSource { return &dataSourceTemplateRackBased{} },
		func() datasource.DataSource { return &dataSourceTemplates{} },
		func() datasource.DataSource { return &dataSourceDatacenterBlueprint{} },
		func() datasource.DataSource { return &dataSourceTag{} },
		func() datasource.DataSource { return &dataSourceVniPool{} },
		func() datasource.DataSource { return &dataSourceVniPools{} },
	}
}

// Resources defines provider resources
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &resourceAgentProfile{} },
		func() resource.Resource { return &resourceAsnPool{} },
		func() resource.Resource { return &resourceBlueprintDeploy{} },
		func() resource.Resource { return &resourceConfiglet{} },
		func() resource.Resource { return &resourceDatacenterBlueprint{} },
		func() resource.Resource { return &resourceDeviceAllocation{} },
		func() resource.Resource { return &resourceInterfaceMap{} },
		func() resource.Resource { return &resourceIpv4Pool{} },
		func() resource.Resource { return &resourceIpv6Pool{} },
		func() resource.Resource { return &resourceLogicalDevice{} },
		func() resource.Resource { return &resourceManagedDevice{} },
		func() resource.Resource { return &resourcePoolAllocation{} },
		func() resource.Resource { return &resourceRackType{} },
		////func() resource.Resource { return &resourceSourceTemplateL3Collapsed{} },
		////func() resource.Resource { return &resourceSourceTemplatePodBased{} },
		func() resource.Resource { return &resourceTemplateRackBased{} },
		func() resource.Resource { return &resourceTag{} },
		func() resource.Resource { return &resourceVniPool{} },
	}
}
