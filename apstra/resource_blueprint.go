package apstra

import (
	"context"
	"fmt"
	"github.com/chrismarget-j/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

type resourceBlueprintType struct{}

func (r resourceBlueprintType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
			"template_id": {
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
			},
			"leaf_asn_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"leaf_ip_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"link_ip_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"spine_asn_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"spine_ip_pool_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
		},
	}, nil
}

func (r resourceBlueprintType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceBlueprint{
		p: *(p.(*provider)),
	}, nil
}

type resourceBlueprint struct {
	p provider
}

func (r resourceBlueprint) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan ResourceBlueprint
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var missing string
	var err error

	// ensure ASN pools from the plan exist on Apstra
	missing, err = asnPoolsExist(ctx, append(plan.LeafAsns, plan.SpineAsns...), r.p.client)
	if err != nil {
		resp.Diagnostics.AddError("error fetching available ASN pools", err.Error())
		return
	}
	if missing != "" {
		resp.Diagnostics.AddError("cannot assign ASN pool",
			fmt.Sprintf("requested pool '%s' does not exist", missing))
		return
	}

	// ensure IP4 pools from the plan exist on Apstra
	missing, err = ip4PoolsExist(ctx, append(append(plan.LeafIp4s, plan.SpineIp4s...), plan.LinkIp4s...), r.p.client)
	if err != nil {
		resp.Diagnostics.AddError("error fetching available IP pools", err.Error())
		return
	}
	if missing != "" {
		resp.Diagnostics.AddError("cannot assign IP pool",
			fmt.Sprintf("requested pool '%s' does not exist", missing))
		return
	}

	// create blueprint
	id, err := r.p.client.CreateBlueprintFromTemplate(ctx, &goapstra.CreateBluePrintFromTemplate{
		RefDesign:  goapstra.RefDesignDatacenter,
		Label:      plan.Name.Value,
		TemplateId: plan.TemplateId.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError("error creating Blueprint", err.Error())
		return
	}
	plan.Id = types.String{Value: string(id)}

	// assign leaf ASN pool
	err = r.p.client.SetResourceAllocation(ctx, id, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeAsnPool,
		Name:    goapstra.ResourceGroupNameLeafAsn,
		PoolIds: tfStringSliceToSliceObjectId(plan.LeafAsns),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// assign leaf IP4 pool
	err = r.p.client.SetResourceAllocation(ctx, id, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeIp4Pool,
		Name:    goapstra.ResourceGroupNameLeafIps,
		PoolIds: tfStringSliceToSliceObjectId(plan.LeafIp4s),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// assign link IP4 pool
	err = r.p.client.SetResourceAllocation(ctx, id, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeIp4Pool,
		Name:    goapstra.ResourceGroupNameLinkIps,
		PoolIds: tfStringSliceToSliceObjectId(plan.LinkIp4s),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// assign spine ASN pool
	err = r.p.client.SetResourceAllocation(ctx, id, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeAsnPool,
		Name:    goapstra.ResourceGroupNameSpineAsn,
		PoolIds: tfStringSliceToSliceObjectId(plan.SpineAsns),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	// assign spine IP4 pool
	err = r.p.client.SetResourceAllocation(ctx, id, &goapstra.ResourceGroupAllocation{
		Type:    goapstra.ResourceTypeIp4Pool,
		Name:    goapstra.ResourceGroupNameSpineIps,
		PoolIds: tfStringSliceToSliceObjectId(plan.SpineIp4s),
	})
	if err != nil {
		resp.Diagnostics.AddError("error setting resource group allocation", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceBlueprint) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state ResourceBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blueprint, err := r.p.client.GetBlueprint(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError("error fetching blueprint", err.Error())
		return
	}

	leafAsns, err := r.p.client.GetResourceAllocation(ctx, blueprint.Id, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeAsnPool,
		Name: goapstra.ResourceGroupNameLeafAsn,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}
	leafIps, err := r.p.client.GetResourceAllocation(ctx, blueprint.Id, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeIp4Pool,
		Name: goapstra.ResourceGroupNameLeafIps,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}
	linkIps, err := r.p.client.GetResourceAllocation(ctx, blueprint.Id, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeIp4Pool,
		Name: goapstra.ResourceGroupNameLinkIps,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}
	spineAsns, err := r.p.client.GetResourceAllocation(ctx, blueprint.Id, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeAsnPool,
		Name: goapstra.ResourceGroupNameSpineAsn,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}
	spineIps, err := r.p.client.GetResourceAllocation(ctx, blueprint.Id, &goapstra.ResourceGroupAllocation{
		Type: goapstra.ResourceTypeIp4Pool,
		Name: goapstra.ResourceGroupNameSpineIps,
	})
	if err != nil {
		resp.Diagnostics.AddError("error reading blueprint resource allocation", err.Error())
		return
	}

	state.Name = types.String{Value: blueprint.Label}
	state.LeafAsns = resourceGroupAllocationToTfStringSlice(leafAsns)
	state.LeafIp4s = resourceGroupAllocationToTfStringSlice(leafIps)
	state.LinkIp4s = resourceGroupAllocationToTfStringSlice(linkIps)
	state.SpineAsns = resourceGroupAllocationToTfStringSlice(spineAsns)
	state.SpineIp4s = resourceGroupAllocationToTfStringSlice(spineIps)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r resourceBlueprint) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve state
	var state ResourceBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve plan
	var plan ResourceBlueprint
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error

	if !setsOfStringsMatch(plan.LeafAsns, state.LeafAsns) {
		err = r.p.client.SetResourceAllocation(ctx, goapstra.ObjectId(plan.Id.Value), &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeAsnPool,
			Name:    goapstra.ResourceGroupNameLeafIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.LeafAsns),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
	}

	if !setsOfStringsMatch(plan.LeafIp4s, state.LeafIp4s) {
		err = r.p.client.SetResourceAllocation(ctx, goapstra.ObjectId(plan.Id.Value), &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeIp4Pool,
			Name:    goapstra.ResourceGroupNameLeafIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.LeafIp4s),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
	}

	if !setsOfStringsMatch(plan.LinkIp4s, state.LinkIp4s) {
		err = r.p.client.SetResourceAllocation(ctx, goapstra.ObjectId(plan.Id.Value), &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeIp4Pool,
			Name:    goapstra.ResourceGroupNameLinkIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.LinkIp4s),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
	}

	if !setsOfStringsMatch(plan.SpineAsns, state.SpineAsns) {
		err = r.p.client.SetResourceAllocation(ctx, goapstra.ObjectId(plan.Id.Value), &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeAsnPool,
			Name:    goapstra.ResourceGroupNameSpineIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.SpineAsns),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
	}

	if !setsOfStringsMatch(plan.SpineIp4s, state.SpineIp4s) {
		err = r.p.client.SetResourceAllocation(ctx, goapstra.ObjectId(plan.Id.Value), &goapstra.ResourceGroupAllocation{
			Type:    goapstra.ResourceTypeIp4Pool,
			Name:    goapstra.ResourceGroupNameSpineIps,
			PoolIds: tfStringSliceToSliceObjectId(plan.SpineIp4s),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("error allocating resource", err.Error())
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (r resourceBlueprint) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ResourceBlueprint
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.p.client.DeleteBlueprint(ctx, goapstra.ObjectId(state.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError("error deleting blueprint", err.Error())
		return
	}
}

func resourceGroupAllocationToTfStringSlice(in *goapstra.ResourceGroupAllocation) []types.String {
	var result []types.String
	for _, pool := range in.PoolIds {
		result = append(result, types.String{Value: string(pool)})
	}
	return result
}

func tfStringSliceToSliceObjectId(in []types.String) []goapstra.ObjectId {
	var result []goapstra.ObjectId
	for _, s := range in {
		result = append(result, goapstra.ObjectId(s.Value))
	}
	return result
}

func asnPoolsExist(ctx context.Context, in []types.String, client *goapstra.Client) (string, error) {
	poolsPerApi, err := client.ListAsnPoolIds(ctx)
	if err != nil {
		return "", fmt.Errorf("error listing available resource pool IDs")
	}

testPool:
	for _, testPool := range in {
		for _, apiPool := range poolsPerApi {
			if goapstra.ObjectId(testPool.Value) == apiPool {
				continue testPool // this one's good, check the next testPool
			}
		}
		return testPool.Value, nil // we looked at every apiPool, none matched testPool
	}
	return "", nil
}

func ip4PoolsExist(ctx context.Context, in []types.String, client *goapstra.Client) (string, error) {
	poolsPerApi, err := client.ListIp4PoolIds(ctx)
	if err != nil {
		return "", fmt.Errorf("error listing available resource pool IDs")
	}

testPool:
	for _, testPool := range in {
		for _, apiPool := range poolsPerApi {
			if goapstra.ObjectId(testPool.Value) == apiPool {
				continue testPool // this one's good, check the next testPool
			}
		}
		return testPool.Value, nil // we looked at every apiPool, none matched testPool
	}
	return "", nil
}

func setsOfStringsMatch(a []types.String, b []types.String) bool {
	os.Stderr.WriteString(fmt.Sprintf("xxxx len a %d\n", len(a)))
	os.Stderr.WriteString(fmt.Sprintf("xxxx len b %d\n", len(b)))
	if len(a) != len(b) {
		return false
	}

loopA:
	for _, ta := range a {
		for _, tb := range b {
			if ta.Null == tb.Null && ta.Unknown == tb.Unknown && ta.Value == tb.Value {
				continue loopA
			}
		}
		return false
	}
	return true
}
