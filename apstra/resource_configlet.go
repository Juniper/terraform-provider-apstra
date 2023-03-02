package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &resourceConfiglet{}

//var _ resource.ResourceWithValidateConfig = &resourceConfiglet{}

type resourceConfiglet struct {
	client *goapstra.Client
}

func (o *resourceConfiglet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configlet"
}

func (o *resourceConfiglet) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	o.client = ResourceGetClient(ctx, req, resp)
}

func (o *resourceConfiglet) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This  resource provides details of a specific Configlet.\n\n" +
			"At least one optional attribute is required. ",
		Attributes: Configlet{}.resourceAttributes(),
	}
}
func (o Configlet) make_configlet_request(ctx context.Context, diags *diag.Diagnostics) *goapstra.ConfigletRequest {
	var tf_gen []configletGenerator
	var r *goapstra.ConfigletRequest = &goapstra.ConfigletRequest{}

	diags.Append(o.Generators.ElementsAs(ctx, &tf_gen, true)...)
	r.DisplayName = o.Name.ValueString()
	r.RefArchs = make([]goapstra.RefDesign, len(o.RefArchs.Elements()))
	refArches := make([]string, len(o.RefArchs.Elements()))
	d := o.RefArchs.ElementsAs(ctx, &refArches, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	for i, j := range refArches {
		e := r.RefArchs[i].FromString(j)
		if e != nil {
			diags.AddError(fmt.Sprintf("error parsing reference architecture : %q", j), e.Error())
		}
	}
	r.Generators = make([]goapstra.ConfigletGenerator, len(o.Generators.Elements()))
	dCG := make([]configletGenerator, len(o.Generators.Elements()))
	d = o.Generators.ElementsAs(ctx, &dCG, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	for i, j := range dCG {
		var a goapstra.ApstraPlatformOS
		e := a.FromString(j.ConfigStyle.ValueString())
		if e != nil {
			diags.AddError(fmt.Sprintf("error parsing configlet style : '%s'", j.ConfigStyle.ValueString()), e.Error())
		}
		var s goapstra.ApstraConfigletSection

		e = s.FromString(j.Section.ValueString())
		if e != nil {
			diags.AddError(fmt.Sprintf("error parsing configlet section : '%s'", j.Section.ValueString()), e.Error())
		}
		r.Generators[i] = goapstra.ConfigletGenerator{
			ConfigStyle:          a,
			Section:              s,
			TemplateText:         j.TemplateText.ValueString(),
			NegationTemplateText: j.NegationTemplateText.ValueString(),
			Filename:             j.FileName.ValueString(),
		}
	}
	return r
}

func (o *resourceConfiglet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredCreateDetail)
		return
	}

	// Retrieve values from plan
	var plan Configlet
	req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var r *goapstra.ConfigletRequest

	r = plan.make_configlet_request(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	id, err := o.client.CreateConfiglet(ctx, r)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Configlet", err.Error())
		return
	}
	plan.Id = types.StringValue(id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (o *resourceConfiglet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errDataSourceUnconfiguredSummary, errDatasourceUnconfiguredDetail)
		return
	}

	var state Configlet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var api *goapstra.Configlet
	var ace goapstra.ApstraClientErr

	api, err = o.client.GetConfiglet(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil && errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Id = types.StringValue(string(api.Id))
	state.loadApiData(ctx, api.Data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update resource
func (o *resourceConfiglet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredUpdateDetail)
		return
	}

	// Get plan values
	var plan Configlet
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	c := plan.make_configlet_request(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Update Configlet

	err := o.client.UpdateConfiglet(ctx, goapstra.ObjectId(plan.Id.ValueString()), c)
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error updating Configlet", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resource
func (o *resourceConfiglet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if o.client == nil {
		resp.Diagnostics.AddError(errResourceUnconfiguredSummary, errResourceUnconfiguredDeleteDetail)
		return
	}
	var state Configlet

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Delete Configlet by calling API
	err := o.client.DeleteConfiglet(ctx, goapstra.ObjectId(state.Id.ValueString()))
	if err != nil {
		var ace goapstra.ApstraClientErr
		if errors.As(err, &ace) && ace.Type() != goapstra.ErrNotfound { // 404 is okay - it's the objective
			resp.Diagnostics.AddError("error deleting Configlet", err.Error())
			return
		}
	}
}
