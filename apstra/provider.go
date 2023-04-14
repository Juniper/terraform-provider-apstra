package tfapstra

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"terraform-provider-apstra/apstra/compatibility"
	"time"
)

const (
	defaultTag    = "v0.0.0"
	defaultCommit = "devel"

	envTlsKeyLogFile  = "SSLKEYLOGFILE"
	envApstraUsername = "APSTRA_USER"
	envApstraPassword = "APSTRA_PASS"
	envApstraLogfile  = "APSTRA_LOG"
	envApstraUrl      = "APSTRA_URL"

	blueprintMutexMessage = "locked by terraform at $DATE"
)

var commit, tag string // populated by goreleaser

var _ provider.Provider = &Provider{}

// NewProvider instantiates the provider in main
func NewProvider() provider.Provider {
	l := len(commit)
	switch {
	case l == 0:
		commit = defaultCommit
	case l > 7:
		commit = commit[:8]
	}

	if len(tag) == 0 {
		tag = defaultTag
	}

	return &Provider{
		Version: tag,
		Commit:  commit,
	}
}

// map of mutexes keyed by blueprint ID
var blueprintMutexes map[string]apstra.Mutex

// mutex which we use to control access to blueprintMutexes
var blueprintMutexesMutex sync.Mutex

// Provider fulfils the provider.Provider interface
type Provider struct {
	Version    string
	Commit     string
	configured bool
	client     *apstra.Client
}

// providerData gets instantiated in Provider's Configure() method and
// is made available to the Configure() method of implementations of
// datasource.DataSource and resource.Resource
type providerData struct {
	client           *apstra.Client
	providerVersion  string
	terraformVersion string
	bpLockFunc       func(context.Context, string) error
	bpUnlockFunc     func(context.Context, string) error
}

func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apstra"
	resp.Version = p.Version + "_" + p.Commit
}

func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the apstra server, e.g. `https://apstra.example.com`\n\n" +
					"It is possible to include Apstra API credentials in the URL using [standard syntax]" +
					"(https://datatracker.ietf.org/doc/html/rfc1738#section-3.1). Care should be taken to ensure " +
					"that these credentials aren't accidentally committed to version control, etc... The preferred " +
					"approach is to pass the credentials as environment variables `" + envApstraUsername + "` and `" +
					envApstraPassword + "`.\n\nIf `url` is omitted, environment variable `" + envApstraUrl + "` can " +
					"be used to in its place.\n\nWhen the username or password are embedded in the URL string, any " +
					"special characters must be URL-encoded. For example, `pass^word` would become `pass%5eword`.",
				Optional: true,
			},
			"tls_validation_disabled": schema.BoolAttribute{
				MarkdownDescription: "Set 'true' to disable TLS certificate validation.",
				Optional:            true,
			},
			"blueprint_mutex_disabled": schema.BoolAttribute{
				MarkdownDescription: "Blueprint mutexes are signals that changes are being made in the staging " +
					"Blueprint and other automation processes (including other instances of Terraform)  should wait " +
					"before beginning to make changes of their own. Set this attribute 'true' to skip locking the " +
					"mutex(es) which signal exclusive Blueprint access for all Blueprint changes made in this project.",
				Optional: true,
			},
			"blueprint_mutex_message": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Blueprint mutexes are signals that changes are being made "+
					"in the staging Blueprint and other automation processes (including other instances of "+
					"Terraform)  should wait before beginning to make changes of their own. The mutexes embed a "+
					"human-readable field to reduce confusion in the event a mutex needs to be cleared manually. "+
					"This attribute overrides the default message in that field: %q.", blueprintMutexMessage),
				Optional: true,
			},
			"experimental": schema.BoolAttribute{
				MarkdownDescription: fmt.Sprintf("Sets a flag in the underlying Apstra SDK client object "+
					"which enables *experimental* features. At this time, the only effect is bypassing version "+
					"compatibility checks in the SDK. This provider release is tested with Apstra versions %s.",
					compatibility.SupportedApiVersionsPretty()),
				Optional: true,
			},
		},
	}
}

// Provider configuration struct. Matches GetSchema() output.
type providerConfig struct {
	Url          types.String `tfsdk:"url"`
	TlsNoVerify  types.Bool   `tfsdk:"tls_validation_disabled"`
	MutexDisable types.Bool   `tfsdk:"blueprint_mutex_disabled"`
	MutexMessage types.String `tfsdk:"blueprint_mutex_message"`
	Experimental types.Bool   `tfsdk:"experimental"`
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config providerConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default the mutex message if needed.
	if config.MutexMessage.IsNull() {
		config.MutexMessage = types.StringValue(fmt.Sprintf(blueprintMutexMessage))
	}

	// Populate raw URL string from config or environment.
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

	// Either config or env could have sent us an empty string.
	if apstraUrl == "" {
		resp.Diagnostics.AddError(errInvalidConfig, "Apstra URL: empty string")
		return
	}

	// Parse the URL.
	parsedUrl, err := url.Parse(apstraUrl)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error parsing URL '%s'", config.Url.ValueString()), err.Error())
		return
	}

	// Determine the Apstra username.
	user := parsedUrl.User.Username()
	if user == "" {
		if val, ok := os.LookupEnv(envApstraUsername); ok {
			user = val
		} else {
			resp.Diagnostics.AddError(errProviderInvalidConfig, "unable to determine apstra username")
			return
		}
	}

	// Determine  the Apstra password.
	pass, found := parsedUrl.User.Password()
	if !found {
		if val, ok := os.LookupEnv(envApstraPassword); ok {
			pass = val
		} else {
			resp.Diagnostics.AddError(errProviderInvalidConfig, "unable to determine apstra password")
			return
		}
	}

	// Remove credentials from the URL prior to rendering it into ClientCfg.
	parsedUrl.User = nil

	// Create the clientCfg
	clientCfg := &apstra.ClientCfg{
		Url:          parsedUrl.String(),
		User:         user,
		Pass:         pass,
		Experimental: config.Experimental.ValueBool(),
	}

	// Set up a logger.
	if logFileName, ok := os.LookupEnv(envApstraLogfile); ok {
		logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			resp.Diagnostics.AddError("error opening logfile", err.Error())
			return
		}
		logger := log.New(logFile, "", 0)
		clientCfg.Logger = logger
	}

	// Create the client's httpClient with(out) TLS verification.
	clientCfg.HttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.TlsNoVerify.ValueBool(),
			}}}

	// Set up the TLS session key log.
	if fileName, ok := os.LookupEnv(envTlsKeyLogFile); ok {
		klw, err := newKeyLogWriter(fileName)
		if err != nil {
			resp.Diagnostics.AddError("error setting up TLS key log", err.Error())
			return
		}
		clientCfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.KeyLogWriter = klw
	}

	// Create the Apstra client.
	client, err := clientCfg.NewClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"unable to create client",
			fmt.Sprintf("error creating apstra client - %s", err),
		)
		return
	}

	// Login after client creation so that future parallel
	// workflows don't trigger TOO MANY REQUESTS threshold.
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

	if !config.MutexDisable.ValueBool() {
		// non-nil slice signals intent to lock and track mutexes
		blueprintMutexes = make(map[string]apstra.Mutex, 0)
	}

	bpLockFunc := func(ctx context.Context, id string) error {
		if blueprintMutexes == nil {
			// A nil map indicates we're not configured to lock the mutex.
			return nil
		}

		blueprintMutexesMutex.Lock()
		defer blueprintMutexesMutex.Unlock()

		if _, ok := blueprintMutexes[id]; ok {
			// We have a map entry, so the mutex must be locked already.
			return nil
		}

		bpClient, err := client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(id))
		if err != nil {
			return fmt.Errorf("error creating blueprint client while attempting to lock blueprint mutex - %w", err)
		}

		// Shove the date into the environment so it's available to ExpandEnv.
		// This should probably be in a text/template configuration.
		err = os.Setenv("DATE", time.Now().UTC().String())
		if err != nil {
			return fmt.Errorf("error setting the 'DATE' environment variable - %w", err)
		}

		// Set the mutex message.
		err = bpClient.Mutex.SetMessage(os.ExpandEnv(config.MutexMessage.ValueString()))
		if err != nil {
			return fmt.Errorf("error setting mutex message - %w", err)
		}

		// This is a blocking call. We get the lock, we hit an error, or we wait.
		err = bpClient.Mutex.Lock(ctx)
		if err != nil {
			return fmt.Errorf("error locking blueprint mutex - %w", err)
		}

		// Drop the Mutex into the map so that it can be unlocked after deployment.
		blueprintMutexes[id] = bpClient.Mutex

		return nil
	}

	bpUnlockFunc := func(ctx context.Context, id string) error {
		blueprintMutexesMutex.Lock()
		defer blueprintMutexesMutex.Unlock()
		if m, ok := blueprintMutexes[id]; ok {
			delete(blueprintMutexes, id)
			return m.Unlock(ctx)
		}
		return nil
	}

	// data passed to Resource and DataSource Configure() methods
	pd := &providerData{
		client:           client,
		providerVersion:  version,
		terraformVersion: req.TerraformVersion,
		bpLockFunc:       bpLockFunc,
		bpUnlockFunc:     bpUnlockFunc,
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
		func() resource.Resource { return &resourceDatacenterRoutingZone{} },
		func() resource.Resource { return &resourceDatacenterRoutingPolicy{} },
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
