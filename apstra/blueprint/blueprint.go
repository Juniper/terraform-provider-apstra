package blueprint

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Blueprint struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	TemplateId       types.String `tfsdk:"template_id"`
	FabricAddressing types.String `tfsdk:"fabric_addressing"`
	Status           types.String `tfsdk:"status"`
	SuperspineCount  types.Int64  `tfsdk:"superspine_count"`
	SpineCount       types.Int64  `tfsdk:"spine_count"`
	LeafCount        types.Int64  `tfsdk:"leaf_switch_count"`
	AccessCount      types.Int64  `tfsdk:"access_switch_count"`
	GenericCount     types.Int64  `tfsdk:"generic_system_count"`
	ExternalCount    types.Int64  `tfsdk:"external_router_count"`
}

func (o Blueprint) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Blueprint: Either as a result of a lookup, or user-specified.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ExactlyOneOf(path.Expressions{
					path.MatchRelative(),
					path.MatchRoot("name"),
				}...),
			},
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Name of the Blueprint: Either as a result of a lookup, or user-specified.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"template_id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Template ID will always be null in 'data source' context.",
			Computed:            true,
		},
		"fabric_addressing": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Fabric Addressing will always be null in 'data source' context.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status of the Blueprint",
			Computed:            true,
		},
		"superspine_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
			Computed:            true,
		},
		"spine_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of spine devices in the topology.",
			Computed:            true,
		},
		"leaf_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of leaf switches in the topology.",
			Computed:            true,
		},
		"access_switch_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of access switches in the topology.",
			Computed:            true,
		},
		"generic_system_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of generic systems in the topology.",
			Computed:            true,
		},
		"external_router_count": dataSourceSchema.Int64Attribute{
			MarkdownDescription: "The count of external routers attached to the topology.",
			Computed:            true,
		},
	}
}

func (o Blueprint) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Blueprint ID assigned by Apstra.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Blueprint name.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"template_id": schema.StringAttribute{
			MarkdownDescription: "ID of Rack Based Template used to instantiate the Blueprint.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"fabric_addressing": schema.StringAttribute{
			MarkdownDescription: "Addressing scheme for both superspine/spine and spine/leaf  links. Only " +
				"applicable to Apstra versions 4.1.1 and later.",
			Optional: true,
			Validators: []validator.String{stringvalidator.OneOf(
				goapstra.AddressingSchemeIp4.String(),
				goapstra.AddressingSchemeIp6.String(),
				goapstra.AddressingSchemeIp46.String())},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status of the Blueprint",
			Computed:            true,
		},
		"superspine_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "For 5-stage topologies, the count of superspine devices",
			Computed:            true,
		},
		"spine_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of spine devices in the topology.",
			Computed:            true,
		},
		"leaf_switch_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of leaf switches in the topology.",
			Computed:            true,
		},
		"access_switch_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of access switches in the topology.",
			Computed:            true,
		},
		"generic_system_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of generic systems in the topology.",
			Computed:            true,
		},
		"external_router_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "The count of external routers attached to the topology.",
			Computed:            true,
		},
	}
}

func (o Blueprint) Request(_ context.Context, diags *diag.Diagnostics) *goapstra.CreateBlueprintFromTemplateRequest {
	var fap *goapstra.FabricAddressingPolicy
	if !o.FabricAddressing.IsNull() {
		var ap goapstra.AddressingScheme
		err := ap.FromString(o.FabricAddressing.ValueString())
		if err != nil {
			diags.AddError(
				errProviderBug,
				fmt.Sprintf("error parsing fabric_addressing %q - %s",
					o.FabricAddressing.ValueString(), err.Error()))
			return nil
		}
		fap = &goapstra.FabricAddressingPolicy{
			SpineSuperspineLinks: ap,
			SpineLeafLinks:       ap,
		}
	}

	return &goapstra.CreateBlueprintFromTemplateRequest{
		RefDesign:              goapstra.RefDesignDatacenter,
		Label:                  o.Name.ValueString(),
		TemplateId:             goapstra.ObjectId(o.TemplateId.ValueString()),
		FabricAddressingPolicy: fap,
	}
}

func (o *Blueprint) LoadApiData(_ context.Context, in *goapstra.BlueprintStatus, _ *diag.Diagnostics) {
	o.Id = types.StringValue(in.Id.String())
	o.Name = types.StringValue(in.Label)
	o.TemplateId = types.StringNull()
	o.FabricAddressing = types.StringNull()
	o.Status = types.StringValue(in.Status)
	o.SuperspineCount = types.Int64Value(int64(in.SuperspineCount))
	o.SpineCount = types.Int64Value(int64(in.SpineCount))
	o.LeafCount = types.Int64Value(int64(in.LeafCount))
	o.AccessCount = types.Int64Value(int64(in.AccessCount))
	o.GenericCount = types.Int64Value(int64(in.GenericCount))
	o.ExternalCount = types.Int64Value(int64(in.ExternalRouterCount))
}

func (o *Blueprint) SetName(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	// create a client specific to the reference design
	bpClient, err := client.NewTwoStageL3ClosClient(ctx, goapstra.ObjectId(o.Id.ValueString()))
	if err != nil {
		diags.AddError("error creating Blueprint client", err.Error())
		return
	}

	type node struct {
		Label string            `json:"label,omitempty"`
		Id    goapstra.ObjectId `json:"id,omitempty"`
	}
	response := &struct {
		Nodes map[string]node `json:"nodes"`
	}{}

	err = bpClient.GetNodes(ctx, goapstra.NodeTypeMetadata, response)
	if err != nil {
		diags.AddError(
			fmt.Sprintf(errApiGetWithTypeAndId, "Blueprint Node", bpClient.Id()),
			err.Error(),
		)
		return
	}
	if len(response.Nodes) != 1 {
		diags.AddError(fmt.Sprintf("wrong number of %s nodes", goapstra.NodeTypeMetadata.String()),
			fmt.Sprintf("expecting 1 got %d nodes", len(response.Nodes)))
		return
	}
	var nodeId goapstra.ObjectId
	for _, v := range response.Nodes {
		nodeId = v.Id
	}
	err = bpClient.PatchNode(ctx, nodeId, &node{Label: o.Name.ValueString()}, nil)
	if err != nil {
		diags.AddError(
			fmt.Sprintf(errApiPatchWithTypeAndId, bpClient.Id(), nodeId),
			err.Error(),
		)
		return
	}
}

func (o *Blueprint) MinMaxApiVersions(_ context.Context, diags *diag.Diagnostics) (*version.Version, *version.Version) {
	var min, max *version.Version
	var err error
	if o.FabricAddressing.IsNull() {
		min, err = version.NewVersion("4.1.1")
	} else {
		max, err = version.NewVersion("4.1.0")
	}
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error parsing min/max version - %s", err.Error()))
	}

	return min, max
}
