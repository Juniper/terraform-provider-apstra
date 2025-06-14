package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/compatibility"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

const (
	defaultTag    = "v0.0.0"
	defaultCommit = "devel"

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

var gitCommit, gitTag string // populated by goreleaser

var (
	_ provider.Provider                       = (*Provider)(nil)
	_ provider.ProviderWithEphemeralResources = (*Provider)(nil)
)

// NewProvider instantiates the provider in main
func NewProvider() provider.Provider {
	l := len(gitCommit)
	switch {
	case l == 0:
		gitCommit = defaultCommit
	case l > 7:
		gitCommit = gitCommit[:8]
	}

	if len(gitTag) == 0 {
		gitTag = defaultTag
	}

	return &Provider{
		Version: gitTag,
		Commit:  gitCommit,
	}
}

// map of mutexes keyed by blueprint ID
var blueprintMutexes map[string]apstra.Mutex

// mutex which we use to control access to blueprintMutexes
var blueprintMutexesMutex sync.Mutex

// maps of blueprint clients keyed by blueprint ID
var (
	twoStageL3ClosClients map[string]apstra.TwoStageL3ClosClient
	freeformClients       map[string]apstra.FreeformClient
)

// mutex which we use to control access to twoStageL3ClosClients
var blueprintClientsMutex sync.Mutex

// Provider fulfils the provider.Provider interface
type Provider struct {
	Version string
	Commit  string
}

// providerData gets instantiated in Provider's Configure() method and
// is made available to the Configure() method of implementations of
// datasource.DataSource and resource.Resource
type providerData struct {
	client                  *apstra.Client
	providerVersion         string
	terraformVersion        string
	bpLockFunc              func(context.Context, string) error
	bpUnlockFunc            func(context.Context, string) error
	getTwoStageL3ClosClient func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	getFreeformClient       func(context.Context, string) (*apstra.FreeformClient, error)
	experimental            bool
}

func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apstra"
	resp.Version = p.Version + "-" + p.Commit
}

func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf("The Apstra Provider allows Terraform to manage Juniper Apstra fabrics.\n\n"+
			"It covers day 0 and day 1 operations (design and deployment), and a growing list of day 2 "+
			"capabilities within *Datacenter* and *Freeform* Apstra reference designs Blueprints.\n\n"+
			"Use the navigation tree on the left to read about the available resources and data sources.\n\n"+
			"This release has been tested with Apstra versions %s.\n\n"+
			"Some example projects which make use of this provider can be found [here](https://github.com/Juniper/terraform-apstra-examples).",
			compatibility.SupportedApiVersionsPretty()),
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the apstra server, e.g. `https://apstra.example.com`\n It is possible " +
					"to include Apstra API credentials in the URL using " +
					"[standard syntax](https://datatracker.ietf.org/doc/html/rfc1738#section-3.1). Care should be " +
					"taken to ensure that these credentials aren't accidentally committed to version control, etc... " +
					"The preferred approach is to pass the credentials as environment variables `" +
					constants.EnvUsername + "`  and `" + constants.EnvPassword + "`.\n If `url` is omitted, " +
					"environment variable `" + constants.EnvUrl + "` can be used to in its place.\n When the " +
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
				MarkdownDescription: "Enable *experimental* features. In this release that means:\n" +
					"  - Set the `experimental` flag in the underlying Apstra SDK client object. Doing so permits " +
					"connections to Apstra instances not supported by the SDK.\n",
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
			"env_var_prefix": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("This attribute defines a prefix which redefines all of the " +
					"`APSTRA_*` environment variables. For example, setting `env_var_prefix = \"FOO_\"` will cause " +
					"the provider to learn the Apstra service URL from the `FOO_APSTRA_URL` environment variable " +
					"rather than the `APSTRA_URL` environment variable. This capability is intended to be used " +
					"when configuring multiple instances of the Apstra provider (which talk to multiple Apstra " +
					"servers) in a single Terraform project."),
				Optional:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
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
	EnvVarPrefix types.String `tfsdk:"env_var_prefix"`
}

func (o *providerConfig) fromEnv(_ context.Context, diags *diag.Diagnostics) {
	envVarPrefix := o.EnvVarPrefix.ValueString()

	if s, ok := os.LookupEnv(envVarPrefix + constants.EnvTlsNoVerify); ok && o.TlsNoVerify.IsNull() {
		v, err := strconv.ParseBool(s)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envVarPrefix+constants.EnvTlsNoVerify), err.Error())
		}
		o.TlsNoVerify = types.BoolValue(v)
	}

	if s, ok := os.LookupEnv(envVarPrefix + constants.EnvBlueprintMutexEnabled); ok && o.MutexEnable.IsNull() {
		v, err := strconv.ParseBool(s)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envVarPrefix+constants.EnvBlueprintMutexEnabled), err.Error())
		}
		o.MutexEnable = types.BoolValue(v)
	}

	if s, ok := os.LookupEnv(envVarPrefix + constants.EnvBlueprintMutexMessage); ok && o.MutexMessage.IsNull() {
		if len(s) < 1 {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envVarPrefix+constants.EnvBlueprintMutexMessage),
				fmt.Sprintf("minimum string length 1; got %q", s))
		}
		o.MutexMessage = types.StringValue(s)
	}

	if s, ok := os.LookupEnv(envVarPrefix + constants.EnvExperimental); ok && o.Experimental.IsNull() {
		v, err := strconv.ParseBool(s)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envVarPrefix+constants.EnvExperimental), err.Error())
		}
		o.Experimental = types.BoolValue(v)
	}

	if s, ok := os.LookupEnv(envVarPrefix + constants.EnvApiTimeout); ok && o.ApiTimeout.IsNull() {
		v, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			diags.AddError(fmt.Sprintf("error parsing environment variable %q", envVarPrefix+constants.EnvApiTimeout), err.Error())
		}
		if v < 0 {
			diags.AddError(fmt.Sprintf("invalid value in environment variable %q", envVarPrefix+constants.EnvApiTimeout),
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
	clientCfg, err := utils.NewClientConfig(config.Url.ValueString(), config.EnvVarPrefix.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating Apstra client configuration", err.Error())
		return
	}
	clientCfg.UserAgent = fmt.Sprintf("terraform-provider-apstra/%s", p.Version)

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
		var ace apstra.ClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrCompatibility {
			resp.Diagnostics.AddError( // SDK incompatibility detected
				err.Error(),
				"You may be trying to use an unsupported version of Apstra. Setting `experimental = true` "+
					"in the provider configuration block will bypass compatibility checks.")
			return
		}

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

	if !slices.Contains(compatibility.SupportedApiVersions(), client.ApiVersion()) && !config.Experimental.ValueBool() {
		resp.Diagnostics.AddError( // provider incompatibility detected
			fmt.Sprintf("Incompatible Apstra API Version %s", client.ApiVersion()),
			"You may be trying to use an unsupported version of Apstra. Setting `experimental = true` "+
				"in the provider configuration block will bypass compatibility checks.")
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

	// This function is made available to resources and data sources via providerData.
	// It maintains a cache of TwoStageL3ClosClients so that NewTwoStageL3ClosClient()
	// needs to be invoked only once per blueprint, rather than invoking it in every
	// resource or data source.
	getTwoStageL3ClosClient := func(ctx context.Context, bpId string) (*apstra.TwoStageL3ClosClient, error) {
		// ensure exclusive access to the blueprint client cache
		blueprintClientsMutex.Lock()
		defer blueprintClientsMutex.Unlock()

		// do we already have this client?
		if twoStageL3ClosClient, ok := twoStageL3ClosClients[bpId]; ok {
			return &twoStageL3ClosClient, nil // client found. return it.
		}

		// create new client (this is the expensive-ish API call we're trying to avoid)
		twoStageL3ClosClient, err := client.NewTwoStageL3ClosClient(ctx, apstra.ObjectId(bpId))
		if err != nil {
			return nil, err
		}

		// create the cache if necessary
		if twoStageL3ClosClients == nil {
			twoStageL3ClosClients = make(map[string]apstra.TwoStageL3ClosClient)
		}

		// save a copy of the client in the map / cache
		twoStageL3ClosClients[bpId] = *twoStageL3ClosClient

		return twoStageL3ClosClient, nil
	}

	getFreeformClient := func(ctx context.Context, bpId string) (*apstra.FreeformClient, error) {
		// ensure exclusive access to the blueprint client cache
		blueprintClientsMutex.Lock()
		defer blueprintClientsMutex.Unlock()

		// do we already have this client?
		if freeformClient, ok := freeformClients[bpId]; ok {
			return &freeformClient, nil // client found. return it.
		}

		// create new client (this is the expensive-ish API call we're trying to avoid)
		freeformClient, err := client.NewFreeformClient(ctx, apstra.ObjectId(bpId))
		if err != nil {
			return nil, err
		}

		// create the cache if necessary
		if freeformClients == nil {
			freeformClients = make(map[string]apstra.FreeformClient)
		}

		// save a copy of the client in the map / cache
		freeformClients[bpId] = *freeformClient

		return freeformClient, nil
	}

	// data passed to Resource, DataSource, and Ephemeral Configure() methods
	pd := providerData{
		client:                  client,
		providerVersion:         p.Version + "-" + p.Commit,
		terraformVersion:        req.TerraformVersion,
		bpLockFunc:              bpLockFunc,
		bpUnlockFunc:            bpUnlockFunc,
		getTwoStageL3ClosClient: getTwoStageL3ClosClient,
		getFreeformClient:       getFreeformClient,
		experimental:            config.Experimental.ValueBool(),
	}
	resp.ResourceData = &pd
	resp.DataSourceData = &pd
	resp.EphemeralResourceData = &pd
}

// DataSources defines provider data sources
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &dataSourceAgent{} },
		func() datasource.DataSource { return &dataSourceAgents{} },
		func() datasource.DataSource { return &dataSourceAgentProfile{} },
		func() datasource.DataSource { return &dataSourceAgentProfiles{} },
		func() datasource.DataSource { return &dataSourceAsnPool{} },
		func() datasource.DataSource { return &dataSourceAsnPools{} },
		func() datasource.DataSource { return &dataSourceBlueprintAnomalies{} },
		func() datasource.DataSource { return &dataSourceBlueprintDeploy{} },
		func() datasource.DataSource { return &dataSourceBlueprintIbaPredefinedProbe{} },
		// func() datasource.DataSource { return &dataSourceBlueprintIbaWidget{} },
		// func() datasource.DataSource { return &dataSourceBlueprintIbaWidgets{} },
		// func() datasource.DataSource { return &dataSourceBlueprintIbaDashboard{} },
		func() datasource.DataSource { return &dataSourceBlueprintIbaDashboards{} },
		func() datasource.DataSource { return &dataSourceBlueprintNodeConfig{} },
		func() datasource.DataSource { return &dataSourceBlueprints{} },
		func() datasource.DataSource { return &dataSourceConfiglet{} },
		func() datasource.DataSource { return &dataSourceConfiglets{} },
		func() datasource.DataSource { return &dataSourceDatacenterBlueprint{} },
		func() datasource.DataSource { return &dataSourceDatacenterConfiglet{} },
		func() datasource.DataSource { return &dataSourceDatacenterConfiglets{} },
		func() datasource.DataSource { return &dataSourceDatacenterConnectivityTemplatesStatus{} },
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
		func() datasource.DataSource { return &dataSourceDatacenterExternalGateway{} },
		func() datasource.DataSource { return &dataSourceDatacenterExternalGateways{} },
		func() datasource.DataSource { return &dataSourceDatacenterInterconnectDomain{} },
		func() datasource.DataSource { return &dataSourceDatacenterInterconnectDomains{} },
		func() datasource.DataSource { return &dataSourceDatacenterInterconnectDomainGateway{} },
		func() datasource.DataSource { return &dataSourceDatacenterInterconnectDomainGateways{} },
		func() datasource.DataSource { return &dataSourceDatacenterGraphQuery{} },
		func() datasource.DataSource { return &dataSourceDatacenterPropertySet{} },
		func() datasource.DataSource { return &dataSourceDatacenterPropertySets{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingPolicies{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingPolicy{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingZone{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingZoneConstraint{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingZoneConstraints{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingZones{} },
		func() datasource.DataSource { return &dataSourceDatacenterSecurityPolicies{} },
		func() datasource.DataSource { return &dataSourceDatacenterSecurityPolicy{} },
		func() datasource.DataSource { return &dataSourceDatacenterSystemNode{} },
		func() datasource.DataSource { return &dataSourceDatacenterSystemNodes{} },
		func() datasource.DataSource { return &dataSourceDatacenterSvis{} },
		func() datasource.DataSource { return &dataSourceDatacenterTag{} },
		func() datasource.DataSource { return &dataSourceDatacenterTags{} },
		func() datasource.DataSource { return &dataSourceDatacenterVirtualNetwork{} },
		func() datasource.DataSource { return &dataSourceDatacenterVirtualNetworks{} },
		func() datasource.DataSource { return &dataSourceDeviceConfig{} },
		func() datasource.DataSource { return &dataSourceFreeformAllocGroup{} },
		func() datasource.DataSource { return &dataSourceFreeformBlueprint{} },
		func() datasource.DataSource { return &dataSourceFreeformConfigTemplate{} },
		func() datasource.DataSource { return &dataSourceFreeformGroupGenerator{} },
		func() datasource.DataSource { return &dataSourceFreeformLink{} },
		func() datasource.DataSource { return &dataSourceFreeformPropertySet{} },
		func() datasource.DataSource { return &dataSourceFreeformResourceGenerator{} },
		func() datasource.DataSource { return &dataSourceFreeformResourceGroup{} },
		func() datasource.DataSource { return &dataSourceFreeformResource{} },
		func() datasource.DataSource { return &dataSourceFreeformSystem{} },
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
		func() datasource.DataSource { return &dataSourceTelemetryServiceRegistryEntries{} },
		func() datasource.DataSource { return &dataSourceTelemetryServiceRegistryEntry{} },
		func() datasource.DataSource { return &dataSourceTemplateCollapsed{} },
		func() datasource.DataSource { return &dataSourceTemplatePodBased{} },
		func() datasource.DataSource { return &dataSourceTemplateRackBased{} },
		func() datasource.DataSource { return &dataSourceTemplates{} },
		func() datasource.DataSource { return &dataSourceVirtualNetworkBindingConstructor{} },
		func() datasource.DataSource { return &dataSourceVniPool{} },
		func() datasource.DataSource { return &dataSourceVniPools{} },
	}
}

// EphemeralResources defines provider ephemeral resources
func (p *Provider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		func() ephemeral.EphemeralResource { return &ephemeralToken{} },
	}
}

// Resources defines provider resources
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &resourceAgentProfile{} },
		func() resource.Resource { return &resourceAsnPool{} },
		func() resource.Resource { return &resourceBlueprintDeploy{} },
		// func() resource.Resource { return &resourceBlueprintIbaDashboard{} },
		func() resource.Resource { return &resourceBlueprintIbaProbe{} },
		// func() resource.Resource { return &resourceBlueprintIbaWidget{} },
		func() resource.Resource { return &resourceConfiglet{} },
		func() resource.Resource { return &resourceDatacenterBlueprint{} },
		func() resource.Resource { return &resourceDatacenterConfiglet{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplateAssignments{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplateInterface{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplateLoopback{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplateProtocolEndpoint{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplateSvi{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplateSystem{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplatesAssignment{} },
		func() resource.Resource { return &resourceDatacenterConnectivityTemplate{} },
		func() resource.Resource { return &resourceDatacenterExternalGateway{} },
		func() resource.Resource { return &resourceDatacenterGenericSystem{} },
		func() resource.Resource { return &resourceDatacenterInterconnectDomain{} },
		func() resource.Resource { return &resourceDatacenterInterconnectDomainGateway{} },
		func() resource.Resource { return &resourceDatacenterIpLinkAddressing{} },
		func() resource.Resource { return &resourceDatacenterPropertySet{} },
		func() resource.Resource { return &resourceDatacenterRack{} },
		func() resource.Resource { return &resourceDatacenterRoutingZone{} },
		func() resource.Resource { return &resourceDatacenterRoutingZoneConstraint{} },
		func() resource.Resource { return &resourceDatacenterRoutingZoneLoopbackAddresses{} },
		func() resource.Resource { return &resourceDatacenterRoutingPolicy{} },
		func() resource.Resource { return &resourceDatacenterSecurityPolicy{} },
		func() resource.Resource { return &resourceDatacenterTag{} },
		func() resource.Resource { return &resourceDatacenterVirtualNetwork{} },
		func() resource.Resource { return &resourceDeviceAllocation{} },
		func() resource.Resource { return &resourceFreeformAllocGroup{} },
		func() resource.Resource { return &resourceFreeformBlueprint{} },
		func() resource.Resource { return &resourceFreeformConfigTemplate{} },
		func() resource.Resource { return &resourceFreeformDeviceProfile{} },
		func() resource.Resource { return &resourceFreeformGroupGenerator{} },
		func() resource.Resource { return &resourceFreeformLink{} },
		func() resource.Resource { return &resourceFreeformPropertySet{} },
		func() resource.Resource { return &resourceFreeformResourceGenerator{} },
		func() resource.Resource { return &resourceFreeformResourceGroup{} },
		func() resource.Resource { return &resourceFreeformResource{} },
		func() resource.Resource { return &resourceFreeformSystem{} },
		func() resource.Resource { return &resourceIntegerPool{} },
		func() resource.Resource { return &resourceInterfaceMap{} },
		func() resource.Resource { return &resourceIpv4Pool{} },
		func() resource.Resource { return &resourceIpv6Pool{} },
		func() resource.Resource { return &resourceLogicalDevice{} },
		func() resource.Resource { return &resourceManagedDevice{} },
		func() resource.Resource { return &resourceManagedDeviceAck{} },
		func() resource.Resource { return &resourceModularDeviceProfile{} },
		func() resource.Resource { return &resourceRawJson{} },
		func() resource.Resource { return &resourceResourcePoolAllocation{} },
		func() resource.Resource { return &resourcePropertySet{} },
		func() resource.Resource { return &resourceRackType{} },
		func() resource.Resource { return &resourceTag{} },
		func() resource.Resource { return &resourceTelemetryServiceRegistryEntry{} },
		func() resource.Resource { return &resourceTemplateCollapsed{} },
		func() resource.Resource { return &resourceTemplatePodBased{} },
		func() resource.Resource { return &resourceTemplateRackBased{} },
		func() resource.Resource { return &resourceVniPool{} },
	}
}

func terraformVersionWarnings(_ context.Context, version string, diags *diag.Diagnostics) {
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
