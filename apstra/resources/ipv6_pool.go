package resources

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Ipv6Pool struct {
	Id          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Subnets     types.Set     `tfsdk:"subnets"`
	Total       types.Number  `tfsdk:"total"`
	Status      types.String  `tfsdk:"status"`
	Used        types.Number  `tfsdk:"used"`
	UsedPercent types.Float64 `tfsdk:"used_percentage"`
}

func (o Ipv6Pool) DataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"id": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID of the desired IPv6 Pool. Required when `name` is omitted.",
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
			MarkdownDescription: "Web UI Name of the IPv6 pool. Required when `id` is omitted.",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"subnets": dataSourceSchema.SetNestedAttribute{
			MarkdownDescription: "Detailed info about individual IPv6 CIDR allocations within the IPv6 Pool.",
			Computed:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: dataSourceSchema.NestedAttributeObject{
				Attributes: Ipv6PoolSubnet{}.DataSourceAttributes(),
			},
		},
		"total": dataSourceSchema.NumberAttribute{
			MarkdownDescription: "Total number of addresses in the IPv6 pool.",
			Computed:            true,
		},
		"status": dataSourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv6 pool.",
			Computed:            true,
		},
		"used": dataSourceSchema.NumberAttribute{
			MarkdownDescription: "Count of used addresses in the IPv6 pool.",
			Computed:            true,
		},
		"used_percentage": dataSourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used addresses in the IPv6 pool.",
			Computed:            true,
		},
	}
}

func (o Ipv6Pool) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra ID number of the pool",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Pool name displayed in the Apstra web UI",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"subnets": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "Detailed info about individual IPv6 CIDR allocations within the IPv6 Pool.",
			Required:            true,
			Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: Ipv6PoolSubnet{}.ResourceAttributes(),
			},
		},
		"total": resourceSchema.NumberAttribute{
			MarkdownDescription: "Total number of addresses in the IPv6 pool.",
			Computed:            true,
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Status of the IPv6 pool. " +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
		"used": resourceSchema.NumberAttribute{
			MarkdownDescription: "Count of used addresses in the IPv6 pool. " +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
		"used_percentage": resourceSchema.Float64Attribute{
			MarkdownDescription: "Percent of used addresses in the IPv6 pool. " +
				"Note that this element is probably better read from a `data` source because it will be more up-to-date.",
			Computed: true,
		},
	}
}

func (o *Ipv6Pool) LoadApiData(ctx context.Context, in *apstra.IpPool, diags *diag.Diagnostics) {
	subnets := make([]Ipv6PoolSubnet, len(in.Subnets))
	for i, s := range in.Subnets {
		subnets[i].LoadApiData(ctx, &s, diags)
		if diags.HasError() {
			return
		}
	}

	o.Id = types.StringValue(string(in.Id))
	o.Name = types.StringValue(in.DisplayName)
	o.Status = types.StringValue(in.Status.String())
	o.UsedPercent = types.Float64Value(float64(in.UsedPercentage))
	o.Used = types.NumberValue(utils.BigIntToBigFloat(&in.Used))
	o.Total = types.NumberValue(utils.BigIntToBigFloat(&in.Total))
	o.Subnets = utils.SetValueOrNull(ctx, types.ObjectType{AttrTypes: Ipv6PoolSubnet{}.AttrTypes()}, subnets, diags)
}

func (o *Ipv6Pool) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.NewIpPoolRequest {
	response := apstra.NewIpPoolRequest{
		DisplayName: o.Name.ValueString(),
		Subnets:     make([]apstra.NewIpSubnet, len(o.Subnets.Elements())),
	}

	subnets := make([]Ipv6PoolSubnet, len(o.Subnets.Elements()))
	d := o.Subnets.ElementsAs(ctx, &subnets, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	for i, subnet := range subnets {
		response.Subnets[i] = *subnet.Request(ctx, diags)
	}

	return &response
}
