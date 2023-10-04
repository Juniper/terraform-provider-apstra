package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultTag    = "v0.0.0"
	defaultCommit = "devel"

	envApiTimeout            = "APSTRA_API_TIMEOUT"
	envBlueprintMutexEnabled = "APSTRA_BLUEPRINT_MUTEX_ENABLED"
	envBlueprintMutexMessage = "APSTRA_BLUEPRINT_MUTEX_MESSAGE"
	envExperimental          = "APSTRA_EXPERIMENTAL"
	envTlsNoVerify           = "APSTRA_TLS_VALIDATION_DISABLED"

	blueprintMutexMessage = "locked by terraform at $DATE"

	osxCertErrStringMatch = "certificate is not trusted"
	winCertErrStringMatch = "x509: certificate signed by unknown authority"
	linCertErrStringMatch = "x509: cannot validate certificate for"

	defaultApiTimeout = 10

	disableTlsValidationMsg = `!!! BAD IDEA WARNING !!!

If you expected TLS validation to fail because the Apstra server is not
configured with a trusted certificate, you might consider setting...

	tls_validation_disabled = true

...in the provider configuration block.
https://registry.terraform.io/providers/Juniper/apstra/%s/docs#tls_validation_disabled`
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
	Version string
	Commit  string
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
	resp.Version = p.Version + "-" + p.Commit
}

func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the apstra server, e.g. `https://apstra.example.com`\n It is possible " +
					"to include Apstra API credentials in the URL using " +
					"[standard syntax](https://datatracker.ietf.org/doc/html/rfc1738#section-3.1). Care should be " +
					"taken to ensure that these credentials aren't accidentally committed to version control, etc... " +
					"The preferred approach is to pass the credentials as environment variables `" +
					utils.EnvApstraUsername + "`  and `" + utils.EnvApstraPassword + "`.\n If `url` is omitted, " +
					"environment variable `" + utils.EnvApstraUrl + "` can be used to in its place.\n When the " +
					"username or password are embedded in the URL string, any special characters must be " +
					"URL-encoded. For example, `pass^word` would become `pass%5eword`.",
				Optional:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"tls_validation_disabled": schema.BoolAttribute{
				MarkdownDescription: "Set 'true' to disable TLS certificate validation.",
				Optional:            true,
			},
			"blueprint_mutex_enabled": schema.BoolAttribute{
				MarkdownDescription: "Blueprint mutexes are indicators that changes are being made in a staging " +
					"Blueprint and other automation processes (including other instances of Terraform) should wait " +
					"before beginning to make changes of their own. Setting this attribute 'true' causes the " +
					"provider to lock a blueprint-specific mutex before making any changes. [More info here]" +
					"(https://github.com/Juniper/terraform-provider-apstra/blob/main/kb/blueprint_mutex.md).",
				Optional: true,
			},
			"blueprint_mutex_message": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Blueprint mutexes are signals that changes are being made "+
					"in a staging Blueprint and other automation processes (including other instances of Terraform) "+
					"should wait before beginning to make changes of their own. The mutexes embed a human-readable "+
					"field to reduce confusion in the event a mutex needs to be cleared manually. This attribute "+
					"overrides the default message in that field: %q.", blueprintMutexMessage),
				Optional:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"experimental": schema.BoolAttribute{
				MarkdownDescription: fmt.Sprintf("Sets a flag in the underlying Apstra SDK client object "+
					"which enables *experimental* features. At this time, the only effect is bypassing version "+
					"compatibility checks in the SDK. This provider release is tested with Apstra versions %s.",
					compatibility.SupportedApiVersionsPretty()),
				Optional: true,
			},
			"api_timeout": schema.Int64Attribute{
				MarkdownDescription: fmt.Sprintf("Timeout in seconds for completing API transactions "+
					"with the Apstra server. Omit for default value of %d seconds. Value of 0 results in "+
					"infinite timeout.",
					defaultApiTimeout),
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(0)},
			},
		},
	}
}

// Provider configuration struct. Matches GetSchema() output.
type providerConfig struct {
	Url          types.String `tfsdk:"url"`
	TlsNoVerify  types.Bool   `tfsdk:"tls_validation_disabled"`
	MutexEnable  types.Bool   `tfsdk:"blueprint_mutex_enabled"`
	MutexMessage types.String `tfsdk:"blueprint_mutex_message"`
	Experimental types.Bool   `tfsdk:"experimental"`
	ApiTimeout   types.Int64  `tfsdk:"api_timeout"`
}

func (o *providerConfig) fromEnv(_ context.Context, diags *diag.Diagnostics) {
	if s, ok := os.LookupEnv(envTlsNoVerify); ok && o.TlsNoVerify.IsNull() {
		v, err := strconv.ParseBool(s)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envTlsNoVerify), err.Error())
		}
		o.TlsNoVerify = types.BoolValue(v)
	}

	if s, ok := os.LookupEnv(envBlueprintMutexEnabled); ok && o.MutexEnable.IsNull() {
		v, err := strconv.ParseBool(s)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envBlueprintMutexEnabled), err.Error())
		}
		o.MutexEnable = types.BoolValue(v)
	}

	if s, ok := os.LookupEnv(envBlueprintMutexMessage); ok && o.MutexMessage.IsNull() {
		if len(s) < 1 {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envBlueprintMutexMessage),
				fmt.Sprintf("minimum string length 1; got %q", s))
		}
		o.MutexMessage = types.StringValue(s)
	}

	if s, ok := os.LookupEnv(envExperimental); ok && o.Experimental.IsNull() {
		v, err := strconv.ParseBool(s)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envExperimental), err.Error())
		}
		o.Experimental = types.BoolValue(v)
	}

	if s, ok := os.LookupEnv(envApiTimeout); ok && o.ApiTimeout.IsNull() {
		v, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envApiTimeout), err.Error())
		}
		if v < 0 {
			diags.AddError(fmt.Sprintf("invalid value in environment variable %q", envApiTimeout),
				fmt.Sprintf("minimum permitted value is 0, got %d", v))
		}
		o.ApiTimeout = types.Int64Value(v)
	}
}

func (o providerConfig) handleMutexFlag(_ context.Context, diags *diag.Diagnostics) {
	if o.MutexEnable.IsNull() {
		diags.AddWarning("Possibly unsafe default - No exclusive Blueprint access",
			"The provider's 'blueprint_mutex_enabled' configuration attribute is not set. This attribute is used "+
				"to explicitly opt-in to, or opt-out of, signaling exclusive Blueprint access via a mutex. The default "+
				"behavior (false) does not use a mutex, and is appropriate for learning, development environments, and "+
				"anywhere there's no risk of multiple automation systems attempting to make changes within a single "+
				"Blueprint at the same time.\n\nSet `blueprint_mutex_enabled` to either `true` or `false` to suppresss "+
				"this warning.",
		)
		return
	}

	if !o.MutexEnable.IsNull() && o.MutexEnable.ValueBool() {
		blueprintMutexes = make(map[string]apstra.Mutex, 0) // non-nil slice signals intent use blueprint mutexes
	}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	terraformVersionWarnings(ctx, req.TerraformVersion, &resp.Diagnostics)

	// Retrieve provider data from configuration
	var config providerConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve missing config elements from environment
	config.fromEnv(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default the mutex message if needed.
	if config.MutexMessage.IsNull() {
		config.MutexMessage = types.StringValue(blueprintMutexMessage)
	}

	// Create the Apstra client configuration from the URL and the environment.
	clientCfg, err := utils.NewClientConfig(config.Url.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating Apstra client configuration", err.Error())
		return
	}

	// Set the experimental flag according to the configuration
	clientCfg.Experimental = config.Experimental.ValueBool()

	// Set http client transport configuration
	if transport, ok := clientCfg.HttpClient.Transport.(*http.Transport); ok {
		// Set the TLS InsecureSkipVerify according to the configuration
		transport.TLSClientConfig.InsecureSkipVerify = config.TlsNoVerify.ValueBool()

		// Set HTTP/HTTPS proxies according to the environment
		transport.Proxy = http.ProxyFromEnvironment

		clientCfg.HttpClient.Transport = transport
	}

	// Set the API timeout
	if config.ApiTimeout.IsNull() {
		config.ApiTimeout = types.Int64Value(defaultApiTimeout)
	}

	cfgValue := config.ApiTimeout.ValueInt64()
	switch cfgValue {
	case 0:
		clientCfg.Timeout = -1 * time.Second // negative value is infinite timeout
	default:
		clientCfg.Timeout = time.Duration(cfgValue) * time.Second
	}

	// Create the Apstra client.
	client, err := clientCfg.NewClient(ctx)
	if err != nil {
		ver := p.Version
		if ver == "0.0.0" {
			ver = "latest"
		}
		suggestion := fmt.Sprintf(disableTlsValidationMsg, ver)

		var msg string
		//goland:noinspection GoBoolExpressions // runtime.GOOS is not handled correctly
		switch {
		case runtime.GOOS == "windows" && strings.Contains(err.Error(), winCertErrStringMatch):
			msg = fmt.Sprintf("error creating apstra client - %s\n\n%s", err.Error(), suggestion)
		case runtime.GOOS == "darwin" && strings.Contains(err.Error(), osxCertErrStringMatch):
			msg = fmt.Sprintf("error creating apstra client - %s\n\n%s", err.Error(), suggestion)
		case runtime.GOOS == "linux" && strings.Contains(err.Error(), linCertErrStringMatch):
			msg = fmt.Sprintf("error creating apstra client - %s\n\n%s", err.Error(), suggestion)
		default:
			msg = fmt.Sprintf("error creating apstra client - %s", err.Error())
		}

		resp.Diagnostics.AddError("unable to create client", msg)
		return
	}

	// Login after client creation so that future parallel
	// workflows don't trigger TOO MANY REQUESTS threshold.
	err = client.Login(ctx)
	if err != nil {
		resp.Diagnostics.AddError("apstra login failure", err.Error())
		return
	}

	config.handleMutexFlag(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
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
		providerVersion:  p.Version + "-" + p.Commit,
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
		func() datasource.DataSource { return &dataSourceAgent{} },
		func() datasource.DataSource { return &dataSourceAgents{} },
		func() datasource.DataSource { return &dataSourceAgentProfile{} },
		func() datasource.DataSource { return &dataSourceAgentProfiles{} },
		func() datasource.DataSource { return &dataSourceAnomalies{} },
		func() datasource.DataSource { return &dataSourceAsnPool{} },
		func() datasource.DataSource { return &dataSourceAsnPools{} },
		func() datasource.DataSource { return &dataSourceBlueprintDeploy{} },
		func() datasource.DataSource { return &dataSourceBlueprints{} },
		func() datasource.DataSource { return &dataSourceConfiglet{} },
		func() datasource.DataSource { return &dataSourceConfiglets{} },
		func() datasource.DataSource { return &dataSourceDatacenterBlueprint{} },
		func() datasource.DataSource { return &dataSourceDatacenterConfiglet{} },
		func() datasource.DataSource { return &dataSourceDatacenterConfiglets{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtBgpPeeringGenericSystem{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtBgpPeeringIpEndpoint{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtCustomStaticRoute{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtDynamicBgpPeering{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtIpLink{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtRoutingPolicy{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtRoutingZoneConstraint{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtStaticRoute{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtVnSingle{} },
		func() datasource.DataSource { return &dataSourceDatacenterCtVnMultiple{} },
		func() datasource.DataSource { return &dataSourceDatacenterIbaWidget{} },
		func() datasource.DataSource { return &dataSourceDatacenterIbaWidgets{} },
		func() datasource.DataSource { return &dataSourceDatacenterIbaDashboard{} },
		func() datasource.DataSource { return &dataSourceDatacenterIbaDashboards{} },
		func() datasource.DataSource { return &dataSourceDatacenterPropertySet{} },
		func() datasource.DataSource { return &dataSourceDatacenterPropertySets{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingPolicy{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingZone{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingZones{} },
		func() datasource.DataSource { return &dataSourceDatacenterSystemNode{} },
		func() datasource.DataSource { return &dataSourceDatacenterSystemNodes{} },
		func() datasource.DataSource { return &dataSourceDatacenterSvis{} },
		func() datasource.DataSource { return &dataSourceDatacenterVirtualNetworks{} },
		func() datasource.DataSource { return &dataSourceIntegerPool{} },
		func() datasource.DataSource { return &dataSourceInterfacesByLinkTag{} },
		func() datasource.DataSource { return &dataSourceInterfacesBySystem{} },
		func() datasource.DataSource { return &dataSourceIntegerPools{} },
		func() datasource.DataSource { return &dataSourceInterfaceMap{} },
		func() datasource.DataSource { return &dataSourceInterfaceMaps{} },
		func() datasource.DataSource { return &dataSourceIpv4Pool{} },
		func() datasource.DataSource { return &dataSourceIpv4Pools{} },
		func() datasource.DataSource { return &dataSourceIpv6Pool{} },
		func() datasource.DataSource { return &dataSourceIpv6Pools{} },
		func() datasource.DataSource { return &dataSourceLogicalDevice{} },
		func() datasource.DataSource { return &dataSourcePropertySet{} },
		func() datasource.DataSource { return &dataSourcePropertySets{} },
		func() datasource.DataSource { return &dataSourceRackType{} },
		func() datasource.DataSource { return &dataSourceRackTypes{} },
		func() datasource.DataSource { return &dataSourceTag{} },
		// func() datasource.DataSource { return &dataSourceTemplateL3Collapsed{} },
		// func() datasource.DataSource { return &dataSourceTemplatePodBased{}},
		func() datasource.DataSource { return &dataSourceTemplateRackBased{} },
		func() datasource.DataSource { return &dataSourceTemplates{} },
		func() datasource.DataSource { return &dataSourceVirtualNetworkBindingConstructor{} },
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
		func() resource.Resource { return &resourceDatacenterConfiglet{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplate{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplateAssignment{} },
		func() resource.Resource { return &resourceDatacenterGenericSystem{} },
		func() resource.Resource { return &resourceDatacenterPropertySet{} },
		func() resource.Resource { return &resourceDatacenterRoutingZone{} },
		func() resource.Resource { return &resourceDatacenterRoutingPolicy{} },
		func() resource.Resource { return &resourceDatacenterVirtualNetwork{} },
		func() resource.Resource { return &resourceDeviceAllocation{} },
		func() resource.Resource { return &resourceIntegerPool{} },
		func() resource.Resource { return &resourceInterfaceMap{} },
		func() resource.Resource { return &resourceIpv4Pool{} },
		func() resource.Resource { return &resourceIpv6Pool{} },
		func() resource.Resource { return &resourceLogicalDevice{} },
		func() resource.Resource { return &resourceManagedDevice{} },
		func() resource.Resource { return &resourceManagedDeviceAck{} },
		func() resource.Resource { return &resourcePoolAllocation{} },
		func() resource.Resource { return &resourcePropertySet{} },
		func() resource.Resource { return &resourceRackType{} },
		func() resource.Resource { return &resourceTag{} },
		// func() resource.Resource { return &resourceSourceTemplateL3Collapsed{} },
		// func() resource.Resource { return &resourceSourceTemplatePodBased{} },
		func() resource.Resource { return &resourceTemplateRackBased{} },
		func() resource.Resource { return &resourceVniPool{} },
	}
}

func terraformVersionWarnings(ctx context.Context, version string, diags *diag.Diagnostics) {
	const tf150warning = "" +
		"You're using Terraform %s. Terraform 1.5.0 has a known issue calculating " +
		"plans in certain situations. More info at: https://github.com/hashicorp/terraform/issues/33371"

	warnings := map[string]string{
		"1.5.0": tf150warning,
	}

	if w, ok := warnings[version]; ok {
		diags.AddWarning("known compatibility issue", fmt.Sprintf(w, version))
	}
}
