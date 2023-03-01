package apstra

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

type blueprint struct {
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

func (o blueprint) dataSourceAttributes() map[string]dataSourceSchema.Attribute {
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
			MarkdownDescription: "Deployment status of the blueprint",
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

type blueprintRackBased struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	TemplateId       types.String `tfsdk:"template_id"`
	FabricAddressing types.String `tfsdk:"fabric_addressing"`
}

func (o blueprint) resourceAttributes() map[string]resourceSchema.Attribute {
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
			MarkdownDescription: "ID of Rack Based Template used to instantiate the blueprint.",
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
			MarkdownDescription: "Deployment status of the blueprint",
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

func (o blueprint) request(_ context.Context, diags *diag.Diagnostics) *goapstra.CreateBlueprintFromTemplateRequest {
	var fap *goapstra.FabricAddressingPolicy
	if !o.FabricAddressing.IsNull() {
		var ap goapstra.AddressingScheme
		err := ap.FromString(o.FabricAddressing.ValueString())
		if err != nil {
			diags.AddError(errProviderBug, fmt.Sprintf("error parsing fabric_addressing %q - %s",
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

func (o *blueprint) loadApiData(_ context.Context, in *goapstra.BlueprintStatus, _ *diag.Diagnostics) {
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

// validateTemplateId ensures that the specified Template exists and ensures
// that it is a Rack Based Template
func (o *blueprint) validateTemplateId(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
	// ensure template ID exists and learn its type
	templateType, err := client.GetTemplateType(ctx, goapstra.ObjectId(o.TemplateId.ValueString()))
	if err != nil {
		diags.AddAttributeError(path.Root("template_id"), errApiData,
			fmt.Sprintf("error retrieving Template %q - %s", o.TemplateId.ValueString(), err.Error()))
		return
	}

	// validate expected type
	if templateType != goapstra.TemplateTypeRackBased {
		diags.AddAttributeError(path.Root("template_id"), errInvalidConfig,
			fmt.Sprintf("Template %q has wrong type %q for use in a Rack Based Blueprint",
				o.TemplateId.ValueString(), templateType.String()))
		return
	}
}

//// populateDeviceProfileIds uses the user supplied device_key for each switch to
//// populate device_profile_id (hardware type)
//func (o *blueprintRackBased) populateDeviceProfileIds(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
//	allSystemsInfo := getAllSystemsInfo(ctx, client, diags)
//	if diags.HasError() {
//		return
//	}
//
//	switchMap := make(map[string]blueprintSwitch, len(o.Switches.Elements()))
//	d := o.Switches.ElementsAs(ctx, &switchMap, false)
//	diags.Append(d...)
//	if diags.HasError() {
//		return
//	}
//
//	for switchLabel, plannedSwitch := range switchMap {
//		deviceKey := plannedSwitch.DeviceKey.ValueString()
//		var msi goapstra.ManagedSystemInfo
//		var found bool
//		if msi, found = allSystemsInfo[deviceKey]; !found {
//			diags.AddAttributeError(
//				path.Root("switches").AtMapKey(switchLabel),
//				"managed device not found",
//				fmt.Sprintf("Switch with device_key %q not found among managed devices", deviceKey),
//			)
//			return
//		}
//		plannedSwitch.DeviceProfileId = types.StringValue(string(msi.Facts.AosHclModel))
//		switchMap[switchLabel] = plannedSwitch
//	}
//
//	o.Switches = MapValueOrNull(ctx, types.ObjectType{AttrTypes: blueprintSwitch{}.attrTypes()}, switchMap, diags)
//}
//
//// extractResourcePoolElementByTfsdkTag returns the value (types.Set)
//// identified by fieldName (a tfsdk tag) found within the blueprintRackBased object
//func (o *blueprintRackBased) extractResourcePoolElementByTfsdkTag(fieldName string, diags *diag.Diagnostics) types.Set {
//	v := reflect.ValueOf(o).Elem()
//	// It's possible we can cache this, which is why precompute all these ahead of time.
//	findTfsdkName := func(t reflect.StructTag) string {
//		if tfsdkTag, ok := t.Lookup("tfsdk"); ok {
//			return tfsdkTag
//		}
//		diags.AddError(errProviderBug, fmt.Sprintf("attempt to lookupg nonexistent tfsdk tag '%s'", fieldName))
//		return ""
//	}
//	fieldNames := map[string]int{}
//	for i := 0; i < v.NumField(); i++ {
//		typeField := v.Type().Field(i)
//		tag := typeField.Tag
//		tname := findTfsdkName(tag)
//		fieldNames[tname] = i
//	}
//
//	fieldNum, ok := fieldNames[fieldName]
//	if !ok {
//		diags.AddError(errProviderBug, fmt.Sprintf("field '%s' does not exist within the provided item", fieldName))
//	}
//	fieldVal := v.Field(fieldNum)
//	return fieldVal.Interface().(types.Set)
//}
//
//// setApstraPoolAllocationByTfsdkTag reads the named pool allocation element
//// from the blueprintRackBased object and sets that value in apstra.
//func (o *blueprintRackBased) setApstraPoolAllocationByTfsdkTag(ctx context.Context, tag string, client *goapstra.TwoStageL3ClosClient, diags *diag.Diagnostics) {
//	// extract the poolSet matching 'tag' from o
//	poolSet := o.extractResourcePoolElementByTfsdkTag(tag, diags)
//
//	// create a goapstra.ResourceGroupAllocation object for (a) query and (b) assignment
//	rga := newRga(tfsdkTagToRgn(tag, diags), &poolSet, diags)
//	name := tfsdkTagToRgn(tag, diags)
//	if diags.HasError() {
//		return
//	}
//	rga := &goapstra.ResourceGroupAllocation{
//		ResourceGroup: goapstra.ResourceGroup{
//			Type: resourceTypeNameFromResourceGroupName(name, diags),
//			Name: name,
//		},
//		PoolIds: poolIds,
//	}
//
//	// get apstra's opinion on the resource group in question
//	_, err := client.GetResourceAllocation(ctx, &rga.ResourceGroup)
//	if err != nil {
//		var ace goapstra.ApstraClientErr
//		if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound { // 404?
//			// the blueprint doesn't need/want this resource group.
//			if poolSet.IsNull() {
//				return // apstra does not want, and the set is null - nothing to do
//			}
//			if poolSet.IsUnknown() {
//				// can't have unknown values in TF state
//				o.setResourcePoolElementByTfsdkTag(tag, types.Set{ElemType: types.StringType, Null: true}, diags)
//				return
//			}
//			// blueprint doesn't want it, but the pool is non-null (appears in the TF config) warn the user.
//			diags.AddWarning(warnUnwantedResourceSummary, fmt.Sprintf(warnUnwantedResourceDetail, tag))
//			// overwrite the planned value so that TF state doesn't reflect the unwanted group
//			o.setResourcePoolElementByTfsdkTag(tag, types.Set{ElemType: types.StringType, Null: true}, diags)
//			return
//		} else {
//			// not a 404
//			diags.AddWarning("error reading resource allocation", err.Error())
//			return
//		}
//	}
//
//	if poolSet.IsUnknown() && len(poolSet.Elements()) != 0 {
//		diags.AddError("oh snap", "an unknown, but non-empty poolSet")
//	}
//
//	if poolSet.IsUnknown() {
//		// can't have unknown values in TF state
//		o.setResourcePoolElementByTfsdkTag(tag, types.Set{ElemType: types.StringType, Null: true}, diags)
//		return
//	}
//
//	// Apstra was expecting something, so set it
//	err = client.SetResourceAllocation(ctx, rga)
//	if err != nil {
//		diags.AddError(errSettingAllocation, err.Error())
//	}
//}
//
//func (o *blueprintRackBased) allocateResources(ctx context.Context, client *goapstra.Client, id goapstra.ObjectId, diags *diag.Diagnostics) {
//	// create a client specific to the reference design
//	blueprint, err := client.NewTwoStageL3ClosClient(ctx, id)
//	if err != nil {
//		diags.AddError("error creating blueprint client", err.Error())
//		return
//	}
//
//	// set user-configured resource group allocations
//	for _, t := range listOfResourceGroupAllocationTags() {
//		o.setApstraPoolAllocationByTfsdkTag(ctx, t, blueprint, diags)
//		if diags.HasError() {
//			return
//		}
//	}
//
//}

func (o *blueprint) setName(ctx context.Context, client *goapstra.Client, diags *diag.Diagnostics) {
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
		diags.AddError(errApiError, fmt.Sprintf("error getting nodes of Blueprint %q - %s",
			bpClient.Id(), err.Error()))
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
		diags.AddError(errApiError, fmt.Sprintf("error setting Blueprint %q name in node %q - %s",
			bpClient.Id(), nodeId, err.Error()))
		return
	}
}

func (o *blueprint) minMaxApiVersions(_ context.Context, diags *diag.Diagnostics) (*version.Version, *version.Version) {
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
