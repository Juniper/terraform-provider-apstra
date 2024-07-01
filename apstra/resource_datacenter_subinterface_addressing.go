package tfapstra

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/blueprint"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resourceWithSetBpClientFunc = &resourceDatacenterSubinterfaceAddressing{}
var _ resourceWithSetBpLockFunc = &resourceDatacenterSubinterfaceAddressing{}

type resourceDatacenterSubinterfaceAddressing struct {
	getBpClientFunc func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)
	lockFunc        func(context.Context, string) error
}

func (o *resourceDatacenterSubinterfaceAddressing) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datacenter_subinterface_addressing"
}

func (o *resourceDatacenterSubinterfaceAddressing) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceDatacenterSubinterfaceAddressing) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryDatacenter + "This resource creates IPv4 and IPv6 addressing parameters on L3 " +
			"subinterfaces within a Datacenter Blueprint fabric. It is intended to be used to assign values to " +
			"subinterfaces created as a side-effect of assigning Connectivity Templates containing IP Link primitives."
		Attributes: blueprint.SubinterfaceAddressing{}.ResourceAttributes(),
	}
}

func (o *resourceDatacenterSubinterfaceAddressing) setBpClientFunc(f func(context.Context, string) (*apstra.TwoStageL3ClosClient, error)) {
	o.getBpClientFunc = f
}

func (o *resourceDatacenterSubinterfaceAddressing) setBpLockFunc(f func(context.Context, string) error) {
	o.lockFunc = f
}
