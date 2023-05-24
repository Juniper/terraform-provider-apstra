package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"terraform-provider-apstra/apstra/compatibility"
	"terraform-provider-apstra/apstra/utils"
	"time"
)

const (
	defaultTag    = "v0.0.0"
	defaultCommit = "devel"

	blueprintMutexMessage = "locked by terraform at $DATE"

	osxCertErrStringMatch = "certificate is not trusted"
	winCertErrStringMatch = "x509: certificate signed by unknown authority"
	linCertErrStringMatch = "x509: cannot validate certificate for"

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
	resp.Version = p.Version + "-" + p.Commit
}

func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the apstra server, e.g. `https://apstra.example.com`\n\nIt is possible " +
					"to include Apstra API credentials in the URL using " +
					"[standard syntax](https://datatracker.ietf.org/doc/html/rfc1738#section-3.1). Care should be " +
					"taken to ensure that these credentials aren't accidentally committed to version control, etc... " +
					"The preferred approach is to pass the credentials as environment variables `" +
					utils.EnvApstraUsername + "`  and `" + utils.EnvApstraPassword + "`.\n\nIf `url` is omitted, " +
					"environment variable `" + utils.EnvApstraUrl + "` can be used to in its place.\n\nWhen the " +
					"username or password are embedded in the URL string, any special characters must be " +
					"URL-encoded. For example, `pass^word` would become `pass%5eword`.",
				Optional:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
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

	// Create the Apstra client configuration from the URL and the environment.
	clientCfg, err := utils.NewClientConfig(config.Url.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating Apstra client configuration", err.Error())
		return
	}

	// Set the experimental flag according to the configuration
	clientCfg.Experimental = config.Experimental.ValueBool()

	// Set the TLS InsecureSkipVerify according to the configuration
	if transport, ok := clientCfg.HttpClient.Transport.(*http.Transport); ok {
		transport.TLSClientConfig.InsecureSkipVerify = config.TlsNoVerify.ValueBool()
		clientCfg.HttpClient.Transport = transport
	}

	// Create the Apstra client.
	client, err := clientCfg.NewClient()
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
		func() datasource.DataSource { return &dataSourceAgentProfile{} },
		func() datasource.DataSource { return &dataSourceAgentProfiles{} },
		func() datasource.DataSource { return &dataSourceAsnPool{} },
		func() datasource.DataSource { return &dataSourceAsnPools{} },
		func() datasource.DataSource { return &dataSourceBlueprintDeploy{} },
		func() datasource.DataSource { return &dataSourceBlueprints{} },
		func() datasource.DataSource { return &dataSourceBlueprintSystemNodes{} },
		func() datasource.DataSource { return &dataSourceBlueprintSystemNode{} },
		func() datasource.DataSource { return &dataSourceConfiglet{} },
		func() datasource.DataSource { return &dataSourceConfiglets{} },
		func() datasource.DataSource { return &dataSourceVirtualNetworkBindingConstructor{} },
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
		//func() datasource.DataSource { return &dataSourceTemplateL3Collapsed{} },
		//func() datasource.DataSource { return &dataSourceTemplatePodBased{}},
		func() datasource.DataSource { return &dataSourceDatacenterRoutingZone{} },
		func() datasource.DataSource { return &dataSourceDatacenterRoutingZones{} },
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
		func() resource.Resource { return &resourceDatacenterVirtualNetwork{} },
		func() resource.Resource { return &resourceDeviceAllocation{} },
		func() resource.Resource { return &resourceInterfaceMap{} },
		func() resource.Resource { return &resourceIpv4Pool{} },
		func() resource.Resource { return &resourceIpv6Pool{} },
		func() resource.Resource { return &resourceLogicalDevice{} },
		func() resource.Resource { return &resourceManagedDevice{} },
		func() resource.Resource { return &resourcePoolAllocation{} },
		func() resource.Resource { return &resourcePropertySet{} },
		func() resource.Resource { return &resourceRackType{} },
		//func() resource.Resource { return &resourceSourceTemplateL3Collapsed{} },
		//func() resource.Resource { return &resourceSourceTemplatePodBased{} },
		func() resource.Resource { return &resourceTemplateRackBased{} },
		func() resource.Resource { return &resourceTag{} },
		func() resource.Resource { return &resourceVniPool{} },
	}
}
