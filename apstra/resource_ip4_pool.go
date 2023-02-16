package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

var _ resource.ResourceWithConfigure = &resourceAsnPool{}
var _ resource.ResourceWithValidateConfig = &resourceAsnPool{}

type resourceIp4Pool struct {
	client *goapstra.Client
}

func (o *resourceIp4Pool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip4_pool"
}

func (o *resourceIp4Pool) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if pd, ok := req.ProviderData.(*providerData); ok {
		o.client = pd.client
	} else {
		resp.Diagnostics.AddError(
			errResourceConfigureProviderDataDetail,
			fmt.Sprintf(errResourceConfigureProviderDataDetail, pd, req.ProviderData),
		)
	}
}

func (o *resourceIp4Pool) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource creates an IPv4 resource pool",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Apstra ID number of the resource pool",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Pool name displayed in the Apstra web UI",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"subnets": schema.SetAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "IPv4 allocation in CIDR notation",
			},
		},
	}
}

func (o *resourceIp4Pool) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config rIp4Pool
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnets := make([]string, len(config.Subnets.Elements()))
	d := config.Subnets.ElementsAs(ctx, &subnets, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i := range subnets {
		_, netI, err := net.ParseCIDR(subnets[i])
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("subnets").AtSetValue(types.StringValue(subnets[i])),
				"failure parsing cidr notation", fmt.Sprintf("error parsing '%s' - %s", subnets[i], err.Error()))
			return
		}

		for j := range subnets {
			if j == i {
				continue
			}
			_, netJ, err := net.ParseCIDR(subnets[j])
			if err != nil {
				resp.Diagnostics.AddAttributeError(
					path.Root("subnets").AtSetValue(types.StringValue(subnets[i])),
					"failure parsing cidr notation", fmt.Sprintf("error parsing '%s' - %s", subnets[i], err.Error()))
				return
			}
			if netI.Contains(netJ.IP) || netJ.Contains(netI.IP) {
				resp.Diagnostics.AddAttributeError(
					path.Root("subnets"),
					"pool has overlapping subnets",
					fmt.Sprintf("subnet '%s' and '%s' overlap", subnets[i], subnets[j]))
				return
			}
		}
	}

	// todo check for overlapping subnets
}

func (o *resourceIp4Pool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan rIp4Pool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new Ip4 Pool
	id, err := o.client.CreateIp4Pool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("error creating new IPv4 Pool", err.Error())
		return
	}

	plan.Id = types.StringValue(string(id))
	resp.State.Set(ctx, &plan)
}

func (o *resourceIp4Pool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredReadDetail)
		return
	}

	// Get current state
	var state rIp4Pool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Ip4 pool from API and then update what is in state from what the API returns
	pool, err := o.client.GetIp4Pool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			// resource deleted outside of terraform
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError("error reading IPv4 pool", err.Error())
			return
		}
	}

	// create new state object
	var newState rIp4Pool
	newState.loadApiResponse(ctx, pool, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update resource
func (o *resourceIp4Pool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan rIp4Pool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := plan.request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := o.client.UpdateIp4Pool(ctx, goapstra.ObjectId(plan.Id.ValueString()), request)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // deleted manually since 'plan'?
			resp.State.RemoveResource(ctx)
			return
		}
		// some other unknown error
		resp.Diagnostics.AddError("API error",
			fmt.Sprintf("error updating IPv4 pool '%s' - %s", plan.Id.ValueString(), err.Error()),
		)
		return
	}

	// set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceIp4Pool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}

	var state rIp4Pool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete IPv4 pool by calling API
	err := o.client.DeleteIp4Pool(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound {
			resp.Diagnostics.AddError(
				"error deleting IPv4 pool", err.Error())
		}
		return
	}
}

type rIp4Pool struct {
	Id      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Subnets types.Set    `tfsdk:"subnets"`
}

func (o *rIp4Pool) request(ctx context.Context, diags *diag.Diagnostics) *goapstra.NewIpPoolRequest {
	subnets := make([]string, len(o.Subnets.Elements()))
	d := o.Subnets.ElementsAs(ctx, &subnets, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	result := goapstra.NewIpPoolRequest{
		DisplayName: o.Name.ValueString(),
		Subnets:     make([]goapstra.NewIpSubnet, len(subnets)),
	}

	for i, subnet := range subnets {
		result.Subnets[i].Network = subnet
	}

	return &result
}

func (o *rIp4Pool) loadApiResponse(ctx context.Context, in *goapstra.IpPool, diags *diag.Diagnostics) {
	subnets := make([]string, len(in.Subnets))
	for i, subnet := range in.Subnets {
		subnets[i] = subnet.Network.String()
	}
	if diags.HasError() {
		return
	}

	var d diag.Diagnostics
	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.DisplayName)
	o.Subnets, d = types.SetValueFrom(ctx, types.StringType, subnets)
	diags.Append(d...)
}
