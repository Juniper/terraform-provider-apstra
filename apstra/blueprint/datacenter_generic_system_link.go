package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type DatacenterGenericSystemLink struct {
	SwitchEndpoint types.Object `tfsdk:"switch_endpoint"`
	LagMode        types.String `tfsdk:"lag_mode"`
	GroupLabel     types.String `tfsdk:"group_label"`
	Id             types.String `tfsdk:"id"`
}

func (o DatacenterGenericSystemLink) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"switch_endpoint": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Identifies the physical switch, physical interface and operational mode " +
				"of the *switch* end of the link.",
			Attributes: DatacenterGenericSystemLinkEndpoint{}.ResourceAttributes(),
			Required:   true,
		},
		"lag_mode": resourceSchema.StringAttribute{
			MarkdownDescription: "LAG negotiation mode of the Link.",
			Optional:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				apstra.RackLinkLagModeActive.String(),
				apstra.RackLinkLagModePassive.String(),
				apstra.RackLinkLagModeStatic.String(),
			)},
		},
		"group_label": resourceSchema.StringAttribute{
			MarkdownDescription: "This field can be used to force multiple links into different groups. For " +
				"example, to create two LACP pairs from four physical links, you might use `group_label` " +
				"value \"bond0\" on two links and \"bond1\" on the other two links",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Graph datastore ID of the link node.",
			Computed:            true,
		},
	}
}

func (o DatacenterGenericSystemLink) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"switch_endpoint": types.ObjectType{AttrTypes: DatacenterGenericSystemLinkEndpoint{}.attrTypes()},
		"lag_mode":        types.StringType,
		"group_label":     types.StringType,
		"id":              types.StringType,
	}
}

func (o DatacenterGenericSystemLink) Request(ctx context.Context, diags *diag.Diagnostics) *apstra.SwitchLink {
	switchEndpoint := o.endpoint(ctx, diags)
	if diags.HasError() {
		return nil
	}

	result := &apstra.SwitchLink{
		//Tags:           nil, // todo
		//SystemEndpoint: apstra.SwitchLinkEndpoint{}, // no need with logical-device-based cabling
		//LagMode:        0, // included below
		SwitchEndpoint: switchEndpoint.Request(ctx, diags),
		GroupLabel:     o.GroupLabel.ValueString(),
	}
	if diags.HasError() {
		return nil
	}

	err := result.LagMode.FromString(o.LagMode.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("failed converting %s to LAG Mode", o.LagMode), err.Error())
		return nil
	}

	return result
}

func (o *DatacenterGenericSystemLink) GenericSystemQuery() *apstra.PathQuery {
	query := new(apstra.PathQuery).
		Node([]apstra.QEEAttribute{apstra.NodeTypeLink.QEEAttribute(),
			{"id", apstra.QEStringVal(o.Id.ValueString())},
		}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
		Node([]apstra.QEEAttribute{apstra.NodeTypeSystem.QEEAttribute(),
			{"role", apstra.QEStringVal("generic")},
			{"name", apstra.QEStringVal("n_generic")},
		})

	return query
}

func (o DatacenterGenericSystemLink) GenericSystemQueryResponse() struct {
	Items []struct {
		Generic struct {
			Hostname string `json:"hostname"`
			Id       string `json:"id"`
			Label    string `json:"label"`
		} `json:"n_generic"`
	} `json:"items"`
} {
	return struct {
		Items []struct {
			Generic struct {
				Hostname string `json:"hostname"`
				Id       string `json:"id"`
				Label    string `json:"label"`
			} `json:"n_generic"`
		} `json:"items"`
	}{}
}

func (o *DatacenterGenericSystemLink) endpoint(ctx context.Context, diags *diag.Diagnostics) *DatacenterGenericSystemLinkEndpoint {
	var result DatacenterGenericSystemLinkEndpoint
	diags.Append(o.SwitchEndpoint.As(ctx, &result, basetypes.ObjectAsOptions{})...)
	return &result
}
