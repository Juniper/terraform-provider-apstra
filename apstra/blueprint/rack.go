package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type Rack struct {
	Id                types.String `tfsdk:"id"`
	BlueprintId       types.String `tfsdk:"blueprint_id"`
	Name              types.String `tfsdk:"name"`
	PodId             types.String `tfsdk:"pod_id"`
	RackTypeId        types.String `tfsdk:"rack_type_id"`
	SystemNameOneShot types.Bool   `tfsdk:"system_name_one_shot"`
}

func (o Rack) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"blueprint_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the Blueprint where the Rack should be created.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Name of the Rack.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"pod_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Graph node ID of Pod (3-stage topology) where the new rack should be created. " +
				"Required only in Pod-Based (5-stage) Blueprints.",
			Optional:      true,
			Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"rack_type_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the Global Catalog Rack Type design object to use as a template for this Rack.",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"system_name_one_shot": resourceSchema.BoolAttribute{
			MarkdownDescription: "Boolean to set at creation time only change system name to match rack name.",
			Optional:            true,
			Validators:          []validator.Bool{boolvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("name"))},
		},
	}
}

func (o Rack) Request() *apstra.TwoStageL3ClosRackRequest {
	return &apstra.TwoStageL3ClosRackRequest{
		PodId:      apstra.ObjectId(o.PodId.ValueString()),
		RackTypeId: apstra.ObjectId(o.RackTypeId.ValueString()),
	}
}

func (o Rack) SetName(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	// data structure to use when calling PatchNode
	var patch struct {
		Label string `json:"label"`
	}
	patch.Label = o.Name.ValueString()

	err := client.PatchNode(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), apstra.ObjectId(o.Id.ValueString()), &patch, nil)
	if err != nil {
		diags.AddError("failed to rename Rack", err.Error())
		// do not return - we must set the state below
	}
}

func (o Rack) SetSystemNames(ctx context.Context, client *apstra.Client, oldName string, diags *diag.Diagnostics) {
	query := new(apstra.PathQuery).
		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetClient(client).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeRack.QEEAttribute(),
			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRack.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal("switch")},
			{Key: "name", Value: apstra.QEStringVal("n_system")},
		})

	var response struct {
		Items []struct {
			System struct {
				Id       apstra.ObjectId `json:"id"`
				Label    string          `json:"label"`
				Hostname string          `json:"hostname"`
				Role     string          `json:"role"`
			} `json:"n_system"`
		} `json:"items"`
	}

	err := query.Do(ctx, &response)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed querying for switches in rack %s", o.Id), err.Error())
		return
	}

	for _, item := range response.Items {
		// data structure to use when calling PatchNode
		var patch struct {
			Label    string `json:"label"`
			Hostname string `json:"hostname"`
		}
		patch.Label = strings.Replace(item.System.Label, oldName, o.Name.ValueString(), 1)
		patch.Hostname = strings.Replace(strings.Replace(item.System.Label, oldName, o.Name.ValueString(), 1), "_", "-", -1)
		err := client.PatchNode(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), item.System.Id, &patch, nil)
		if err != nil {
			diags.AddError(fmt.Sprintf("failed to rename %s switch in rack %s", item.System.Role, o.Id), err.Error())
		}
	}
}

func (o *Rack) GetName(ctx context.Context, client *apstra.Client) (string, error) {
	// struct used to collect the rack node info
	var node struct {
		Label string `json:"label"`
	}

	// collect the rack node info
	err := client.GetNode(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), apstra.ObjectId(o.Id.ValueString()), &node)
	if err != nil {
		return "", err
	}

	return node.Label, nil
}
