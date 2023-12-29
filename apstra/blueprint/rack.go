package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	apstravalidator "github.com/Juniper/terraform-provider-apstra/apstra/apstra_validator"
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
	"strings"
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
			DeprecationMessage: "Please migrate your configuration to use `rack_elements_name_one_shot`.",
			MarkdownDescription: "Because this resource only manages the Rack, names of Systems defined within the Rack " +
				"are not within this resource's control. When `system_name_one_shot` is `true` during initial Rack " +
				"creation, Systems within the Rack will be renamed to match the rack's `name`. Subsequent modifications " +
				"to the `name` attribute will not affect the names of those systems. It's a create-time one-shot operation.",
			Optional: true,
			Validators: []validator.Bool{
				boolvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("name")),
				apstravalidator.MustBeOneOf([]attr.Value{
					types.BoolValue(true),
					types.BoolNull(),
				}),
				boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("rack_elements_name_one_shot")),
			},
		},
		"rack_elements_name_one_shot": resourceSchema.BoolAttribute{
			MarkdownDescription: "Because this resource only manages the Rack, names of Systems and other embedded " +
				"elements with names derived from the Rack name are not within this resource's control. When `true` " +
				"during initial Rack creation, those elements will be renamed to match the `name` attribute. Subsequent " +
				"modifications to the `name` attribute will not affect those elements. It's a create-time operation only.",
			Optional: true,
			Validators: []validator.Bool{
				boolvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("name")),
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

	err = client.PatchNodes(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), reSlice)
	if err != nil {
		diags.AddError("failed while renaming new rack nodes", err.Error())
	}
}

//func (o Rack) SetSystemNames(ctx context.Context, client *apstra.Client, oldName string, diags *diag.Diagnostics) {
//	query := new(apstra.PathQuery).
//		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
//		SetClient(client).
//		SetBlueprintType(apstra.BlueprintTypeStaging).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeRack.QEEAttribute(),
//			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
//		}).
//		In([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRack.QEEAttribute()}).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeSystem.QEEAttribute(),
//			{Key: "system_type", Value: apstra.QEStringVal("switch")},
//			{Key: "name", Value: apstra.QEStringVal("n_system")},
//		})
//
//	var response struct {
//		Items []struct {
//			System struct {
//				Id       apstra.ObjectId `json:"id"`
//				Label    string          `json:"label"`
//				Hostname string          `json:"hostname"`
//				Role     string          `json:"role"`
//			} `json:"n_system"`
//		} `json:"items"`
//	}
//
//	err := query.Do(ctx, &response)
//	if err != nil {
//		diags.AddError(fmt.Sprintf("failed querying for switches in rack %s", o.Id), err.Error())
//		return
//	}
//
//	// data structure to use when calling PatchNode
//	var patch struct {
//		Label    string `json:"label"`
//		Hostname string `json:"hostname"`
//	}
//
//	// loop over each discovered switch, set the label and hostname
//	for _, item := range response.Items {
//		patch.Label = strings.Replace(item.System.Label, oldName, o.Name.ValueString(), 1)
//		patch.Hostname = strings.Replace(strings.Replace(item.System.Label, oldName, o.Name.ValueString(), 1), "_", "-", -1)
//		err := client.PatchNode(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), item.System.Id, &patch, nil)
//		if err != nil {
//			diags.AddError(fmt.Sprintf("failed to rename %s switch %s in rack %s", item.System.Role, item.System.Id, o.Id), err.Error())
//		}
//	}
//}

//func (o Rack) SetRedundancyGroupNames(ctx context.Context, client *apstra.Client, oldName string, diags *diag.Diagnostics) {
//	query := new(apstra.PathQuery).
//		SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
//		SetClient(client).
//		SetBlueprintType(apstra.BlueprintTypeStaging).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeRack.QEEAttribute(),
//			{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())},
//		}).
//		In([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRack.QEEAttribute()}).
//		Node([]apstra.QEEAttribute{
//			apstra.NodeTypeRedundancyGroup.QEEAttribute(),
//			{Key: "name", Value: apstra.QEStringVal("n_redundancy_group")},
//		})
//
//	var response struct {
//		Items []struct {
//			RedundancyGroup struct {
//				Id    apstra.ObjectId `json:"id"`
//				Label string          `json:"label"`
//			} `json:"n_redundancy_group"`
//		} `json:"items"`
//	}
//
//	err := query.Do(ctx, &response)
//	if err != nil {
//		diags.AddError(fmt.Sprintf("failed querying for redundancy groups in rack %s", o.Id), err.Error())
//		return
//	}
//
//	// data structure to use when calling PatchNode
//	var patch struct {
//		Label string `json:"label"`
//	}
//
//	// loop over each discovered redundancy group, set the label and hostname
//	for _, item := range response.Items {
//		patch.Label = strings.Replace(item.RedundancyGroup.Label, oldName, o.Name.ValueString(), 1)
//		err := client.PatchNode(ctx, apstra.ObjectId(o.BlueprintId.ValueString()), item.RedundancyGroup.Id, &patch, nil)
//		if err != nil {
//			diags.AddError(fmt.Sprintf("failed to rename redundancy group %s in rack %s", item.RedundancyGroup.Id, o.Id), err.Error())
//		}
//	}
//}

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

//	func (o *Rack) getPartOfRackElements(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) map[apstra.ObjectId]RackElement {
//		query := new(apstra.PathQuery).
//			SetClient(client).
//			SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
//			SetBlueprintType(apstra.BlueprintTypeStaging).
//			Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())}}). // rack node
//			In([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRack.QEEAttribute()}).
//			Node([]apstra.QEEAttribute{ // switch and redundancy group nodes
//				//apstra.NodeTypeSystem.QEEAttribute(),
//				//{Key: "system_type", Value: apstra.QEStringVal("switch")},
//				{Key: "name", Value: apstra.QEStringVal("n_obj")},
//			})
//
//		var response struct {
//			Items []struct {
//				Obj RackElement `json:"n_obj"`
//			} `json:"items"`
//		}
//
//		err := query.Do(ctx, &response)
//		if err != nil {
//			diags.AddError(fmt.Sprintf("failed querying for switches in rack %s", o.Id), err.Error())
//			return nil
//		}
//
//		result := make(map[apstra.ObjectId]RackElement, len(response.Items))
//		for _, item := range response.Items {
//			result[item.Obj.Id] = item.Obj
//		}
//
//		return result
//	}
//
//	func (o *Rack) getAllRackElements(ctx context.Context, switchIds []string, client *apstra.Client, diags *diag.Diagnostics) map[apstra.ObjectId]RackElement {
//		partOfRackQuery := new(apstra.PathQuery).
//			Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(o.Id.ValueString())}}). // rack node
//			In([]apstra.QEEAttribute{apstra.RelationshipTypePartOfRack.QEEAttribute()}).
//			Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_part_of_rack")}})
//
//		linkQuery := new(apstra.PathQuery).
//			Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_part_of_rack")}}).
//			Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
//			Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
//			Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
//			Node([]apstra.QEEAttribute{
//				apstra.NodeTypeLink.QEEAttribute(),
//				{Key: "name", Value: apstra.QEStringVal("n_link")},
//			})
//
//		serverQuery := new(apstra.PathQuery).
//			Node([]apstra.QEEAttribute{{Key: "name", Value: apstra.QEStringVal("n_link")}}).
//			In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
//			Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
//			In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
//			Node([]apstra.QEEAttribute{
//				apstra.NodeTypeSystem.QEEAttribute(),
//				{Key: "system_type", Value: apstra.QEStringVal("server")},
//				{Key: "external", Value: apstra.QEBoolVal(false)},
//				{Key: "name", Value: apstra.QEStringVal("n_server")},
//			})
//
//		query := new(apstra.MatchQuery).
//			Match(partOfRackQuery).
//			Optional(new(apstra.MatchQuery).
//				Match(linkQuery).
//				Optional(serverQuery))
//
//		qString := query.String()
//		_ = qString
//
//		var result struct {
//			Items []struct {
//				PartOfRack *RackElement `json:"n_part_of_rack"`
//				Link       *RackElement `json:"n_link"`
//				Server     *RackElement `json:"n_server"`
//			} `json:"items"`
//		}
//
//		err := query.Do(ctx, &result)
//		if err != nil {
//			diags.AddError("failed querying for rack elements", err.Error())
//			return nil
//		}
//
// }
//
//	func (o *Rack) GetRackElementsIdsAndLabels(ctx context.Context, client *apstra.Client, diags *diag.Diagnostics) []RackElement {
//		// partsOfRack are RackElements discovered using 'part_of_rack' graph relationship.
//		// They include leaf switches, access switches, and leaf/access redundancy groups,
//		// each of which need to be re-labeled to match the rack name.
//		result := o.getPartOfRackElements(ctx, client, diags)
//		if diags.HasError() {
//			return nil
//		}
//
//		// partOfRackIds are used for tracking down links and servers
//		partOfRackIds := make([]string, len(result))
//		i := 0
//		for _, partOfRack := range result {
//			partOfRackIds[i] = partOfRack.Id.String()
//			i++
//		}
//
//		// we want links from leaf/access systems (Ethernet and LAG) and from redundancy groups (MLAG)
//		links := o.getLinkRackElements(ctx, switchIds, client, diags)
//		if diags.HasError() {
//			return nil
//		}
//
//		maps.Copy(result, links) // add links to the result map
//
//		for id, link := range links {
//
//		}
//
//		for i := len(links) - 1; i >= 0; i-- {
//			if !ids[links[i].Id] {
//				// this is a new ID; add it to the list
//				ids[links[i].Id] = true
//				continue
//			}
//
//			// this id has been seen previously; remove it from the slice
//			links[i] = links[len(links)-1]
//			links = links[:len(links)-1]
//		}
//
//		response := make([]RackElement, len(rPartOfRack.Items))
//		for i, item := range rPartOfRack.Items {
//			response[i] = item.Obj
//			ids[item.Obj.Id] = true
//		}
//
//		for _, node := range response {
//			if node.Type != "system" {
//				continue // we're looking only for leaf/access at this stage
//			}
//
//			qLink := new(apstra.PathQuery).
//				SetClient(client).
//				SetBlueprintId(apstra.ObjectId(o.BlueprintId.ValueString())).
//				SetBlueprintType(apstra.BlueprintTypeStaging).
//				Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(node.Id.String())}}).
//				Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(node.Id.String())}}).
//				Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
//				Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
//				Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
//				Node([]apstra.QEEAttribute{
//					apstra.NodeTypeLink.QEEAttribute(),
//					{Key: "name", Value: apstra.QEStringVal("n_link")},
//				})
//
//			var rLink struct {
//				Items []struct {
//					Link RackElement `json:"n_link"`
//				} `json:"items"`
//			}
//
//			err = qLink.Do(ctx, &rLink)
//			if err != nil {
//				diags.AddError(fmt.Sprintf("failed querying for links from node %s", node.Id), err.Error())
//				return nil
//			}
//
//			// add discovered links to the response
//			for _, item := range rLink.Items {
//				if ids[item.Link.Id] {
//					continue
//				}
//
//				response = append(response, item.Link)
//				ids[item.Link.Id] = true
//			}
//
//			qGeneric := qLink.
//				In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
//				Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
//				In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
//				Node([]apstra.QEEAttribute{
//					apstra.NodeTypeSystem.QEEAttribute(),
//					{Key: "system_type", Value: apstra.QEStringVal("server")},
//					{Key: "role", Value: apstra.QEStringVal("generic")},
//					{Key: "external", Value: apstra.QEBoolVal(false)},
//					{Key: "name", Value: apstra.QEStringVal("n_generic")},
//				})
//
//			var rGeneric struct {
//				Items []struct {
//					Link RackElement `json:"n_generic"`
//				} `json:"items"`
//			}
//
//			err = qLink.Do(ctx, &rGeneric)
//			if err != nil {
//				diags.AddError(fmt.Sprintf("failed querying for servers from node %s", node.Id), err.Error())
//				return nil
//			}
//
//			for _, item := range qResponse.Items {
//				if ids[item.Id] {
//					continue
//				}
//
//				response = append(response, item)
//				ids[item.Id] = true
//			}
//
//		}
//
//		return response
//	}
type RackElement struct {
	Id       apstra.ObjectId `json:"id"`
	Type     string          `json:"type,omitempty"`
	Label    string          `json:"label"`
	Hostname *string         `json:"hostname,omitempty"`
}
