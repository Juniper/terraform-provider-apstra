package blueprint

import (
	"context"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Rack struct {
	Id                      types.String `tfsdk:"id"`
	BlueprintId             types.String `tfsdk:"blueprint_id"`
	Name                    types.String `tfsdk:"name"`
	PodId                   types.String `tfsdk:"pod_id"`
	RackTypeId              types.String `tfsdk:"rack_type_id"`
	SystemNameOneShot       types.Bool   `tfsdk:"system_name_one_shot"`
	RackElementsNameOneShot types.Bool   `tfsdk:"rack_elements_name_one_shot"`
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
			Required:            true,
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
			DeprecationMessage: "The `system_name_one_shot` attribute is deprecated. Please migrate your configuration " +
				"to use `rack_elements_name_one_shot` instead.",
			MarkdownDescription: "Because this resource only manages the Rack, names of Systems defined within the Rack " +
				"are not within this resource's control. When `system_name_one_shot` is `true` during initial Rack " +
				"creation, Systems within the Rack will be renamed to match the rack's `name`. Subsequent modifications " +
				"to the `name` attribute will not affect the names of those systems. It's a create-time one-shot operation.",
			Optional: true,
			Validators: []validator.Bool{
				boolvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("name").Resolve()),
				apstravalidator.MustBeOneOf([]attr.Value{
					types.BoolValue(true),
					types.BoolNull(),
				}),
				boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("rack_elements_name_one_shot").Resolve()),
			},
		},
		"rack_elements_name_one_shot": resourceSchema.BoolAttribute{
			MarkdownDescription: "Because this resource only manages the Rack, names of Systems and other embedded " +
				"elements with names derived from the Rack name are not within this resource's control. When `true` " +
				"during initial Rack creation, those elements will be renamed to match the `name` attribute. Subsequent " +
				"changes to the `name` attribute will not affect those elements. It's a create-time operation only.",
			Optional: true,
			Validators: []validator.Bool{
				boolvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("name").Resolve()),
				apstravalidator.MustBeOneOf([]attr.Value{
					types.BoolValue(true),
					types.BoolNull(),
				}),
			},
		},
	}
}

func (o *Rack) Request() *apstra.TwoStageL3ClosRackRequest {
	return &apstra.TwoStageL3ClosRackRequest{
		PodId:      apstra.ObjectId(o.PodId.ValueString()),
		RackTypeId: apstra.ObjectId(o.RackTypeId.ValueString()),
	}
}

// SetName sets the name of the rack and (optionally) elements within the rack.
// If oldName is empty, only the rack will be renamed.
func (o *Rack) SetName(ctx context.Context, oldName string, client *apstra.Client, diags *diag.Diagnostics) {
	if oldName == "" {
		o.setRackNameOnly(ctx, client, diags)
	} else {
		o.setRackAndChildNames(ctx, oldName, client, diags)
	}
}

func (o *Rack) setRackNameOnly(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) {
	// data structure to use when calling PatchNode
	var patch struct {
		Label string `json:"label"`
	}
	patch.Label = o.Name.ValueString()

	err := client.PatchNode(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), apstra.ObjectId(o.Id.ValueString()), &patch, nil)
	if err != nil {
		diags.AddError("failed to rename Rack", err.Error())
		return
	}
}

func (o *Rack) setRackAndChildNames(ctx context.Context, oldName string, client *apstra.Client, diags *diag.Diagnostics) {
	partOfRackQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRack.QEEAttribute()}).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_part_of_rack")}})

	// The linkQuery doesn't currently catch the link between subinterfaces between an ESI access switch pair.
	// This might be okay. The link in question doesn't seem to appear in the UI.
	// That graph traversal might look like:
	// node(type='system', role='access').out().node(type='interface').out().node(type='interface', if_type='subinterface').out().node(type='link', name='n_link')
	linkQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_part_of_rack")}}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
		Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeLink.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_link")},
		})

	serverQuery := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_link")}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeSystem.QEEAttribute(),
			{Key: "system_type", Value: apstra.QEStringVal("server")},
			{Key: "external", Value: apstra.QEBoolVal(false)},
			{Key: "name", Value: apstra.QEStringVal("n_server")},
		})

	query := new(apstra.MatchQuery).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
		SetClient(client).
		Match(partOfRackQuery).
		Optional(new(apstra.MatchQuery).
			Match(linkQuery).
			Optional(serverQuery))

	var response struct {
		Items []struct {
			PartOfRack *RackElement `json:"n_part_of_rack"`
			Link       *RackElement `json:"n_link"`
			Server     *RackElement `json:"n_server"`
		} `json:"items"`
	}

	err := query.Do(ctx, &response)
	if err != nil {
		diags.AddError("failed querying for rack elements", err.Error())
	}

	// Create a map of RackElements keyed by graph node ID.
	// We use a map here to eliminate duplicates which appear in the graph query response.
	// The first map element is the rack node, which will not be discovered by the graph query.
	reMap := make(map[apstra.ObjectId]RackElement, len(response.Items)+1)
	reMap[apstra.ObjectId(o.Id.ValueString())] = RackElement{
		Id:    apstra.ObjectId(o.Id.ValueString()),
		Type:  "rack",
		Label: oldName,
	}
	for _, item := range response.Items {
		if item.PartOfRack != nil {
			reMap[item.PartOfRack.Id] = *item.PartOfRack
		}
		if item.Link != nil {
			reMap[item.Link.Id] = *item.Link
		}
		if item.Server != nil {
			reMap[item.Server.Id] = *item.Server
		}
	}

	// Reduce the RackElement map to a slice, performing string
	// substitution to rename the nodes as we go.
	reSlice := make([]interface{}, len(reMap))
	i := 0
	for _, v := range reMap {
		v.Label = strings.Replace(v.Label, oldName, o.Name.ValueString(), -1)
		if v.Hostname != nil {
			hostname := strings.Replace(*v.Hostname, oldName, o.Name.ValueString(), -1)
			v.Hostname = &hostname
		}
		reSlice[i] = v
		i++
	}

	// Send the slice to Apstra to rename the rack elements all in one batch.
	err = client.PatchNodes(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), reSlice)
	if err != nil {
		diags.AddError("failed while renaming new rack nodes", err.Error())
	}
}

// GetName fetches the rack node's label field
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

type RackElement struct {
	Id       apstra.ObjectId `json:"id"`
	Type     string          `json:"type,omitempty"`
	Label    string          `json:"label"`
	Hostname *string         `json:"hostname,omitempty"`
}
