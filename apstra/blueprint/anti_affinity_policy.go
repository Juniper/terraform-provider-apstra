package blueprint

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math"
)

type AntiAffinityPolicy struct {
	MaxLinksCountPerSlot          types.Int64 `tfsdk:"max_links_count_per_slot"`
	MaxLinksCountPerSystemPerSlot types.Int64 `tfsdk:"max_links_count_per_system_per_slot"`
	MaxLinksCountPerPort          types.Int64 `tfsdk:"max_links_count_per_port"`
	MaxLinksCountPerSystemPerPort types.Int64 `tfsdk:"max_links_count_per_system_per_port"`
}

func (o AntiAffinityPolicy) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"max_links_count_per_slot":            types.Int64Type,
		"max_links_count_per_system_per_slot": types.Int64Type,
		"max_links_count_per_port":            types.Int64Type,
		"max_links_count_per_system_per_port": types.Int64Type,
	}
}

func (o AntiAffinityPolicy) datasourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"max_links_count_per_slot": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Maximum total number of links connected to ports/interfaces of the specified slot regardless of the system" +
				"they are targeted to. It controls how many links can be connected to one slot of one system. " +
				"Example: A line card slot in a chassis.",
		},
		"max_links_count_per_system_per_slot": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Restricts the number of links to a certain system connected to the ports/interfaces in a specific slot. " +
				"It controls how many links can be connected to one system to one slot of another system.",
		},
		"max_links_count_per_port": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Maximum total number of links connected to the interfaces of the specific port regardless of the system " +
				"they are targeted to. It controls how many links can be connected to one port in one system. " +
				"Example: Several transformations of one port. In this case, it controls how many transformations can be used in links.",
		},
		"max_links_count_per_system_per_port": dataSourceSchema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Restricts the number of interfaces on a port used to connect to a certain system. It controls " +
				"how many links can be connected from one system to one port of another system. This is the one that you will " +
				"most likely use, for port breakouts.",
		},
	}
}
func (o AntiAffinityPolicy) resourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"max_links_count_per_slot": resourceSchema.Int64Attribute{
			Required: true,
			MarkdownDescription: "Maximum total number of links connected to ports/interfaces of the specified slot regardless of the system" +
				"they are targeted to. It controls how many links can be connected to one slot of one system. " +
				"Example: A line card slot in a chassis.",
			Validators: []validator.Int64{int64validator.Between(0, math.MaxUint8)},
		},
		"max_links_count_per_system_per_slot": resourceSchema.Int64Attribute{
			Required: true,
			MarkdownDescription: "Restricts the number of links to a certain system connected to the ports/interfaces in a specific slot. " +
				"It controls how many links can be connected to one system to one slot of another system.",
			Validators: []validator.Int64{int64validator.Between(0, math.MaxUint8)},
		},
		"max_links_count_per_port": resourceSchema.Int64Attribute{
			Required: true,
			MarkdownDescription: "Maximum total number of links connected to the interfaces of the specific port regardless of the system " +
				"they are targeted to. It controls how many links can be connected to one port in one system. " +
				"Example: Several transformations of one port. In this case, it controls how many transformations can be used in links.",
			Validators: []validator.Int64{int64validator.Between(0, math.MaxUint8)},
		},
		"max_links_count_per_system_per_port": resourceSchema.Int64Attribute{
			Required: true,
			MarkdownDescription: "Restricts the number of interfaces on a port used to connect to a certain system. It controls " +
				"how many links can be connected from one system to one port of another system. This is the one that you will " +
				"most likely use, for port breakouts.",
			Validators: []validator.Int64{int64validator.Between(0, math.MaxUint8)},
		},
	}
}

func (o *AntiAffinityPolicy) loadApiData(_ context.Context, in *apstra.AntiAffinityPolicy, _ *diag.Diagnostics) {
	o.MaxLinksCountPerPort = types.Int64Value(int64(in.MaxLinksPerPort))
	o.MaxLinksCountPerSlot = types.Int64Value(int64(in.MaxLinksPerSlot))
	o.MaxLinksCountPerSystemPerPort = types.Int64Value(int64(in.MaxPerSystemLinksPerPort))
	o.MaxLinksCountPerSystemPerSlot = types.Int64Value(int64(in.MaxPerSystemLinksPerSlot))
}
