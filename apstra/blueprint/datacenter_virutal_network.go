package blueprint

import (
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	apstravalidator "terraform-provider-apstra/apstra/apstra_validator"
)

type DatacenterVirtualNetwork struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	BlueprintId types.String `tfsdk:"blueprint_id"`
}

func (o DatacenterVirtualNetwork) ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Apstra graph node ID.",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Name",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthBetween(1, 30)},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Virtual Network Type",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(apstra.VnTypeVxlan.String()),
			Validators:          []validator.String{apstravalidator.OneOfStringers(apstra.AllNodeDeployModes())},
			//Validators: []validator.String{stringvalidator.OneOf([]string{"foo", "bar"}...)},
			//Validators: []validator.String{
			//	stringvalidator.LengthAtLeast(5),
			//	stringvalidator.RegexMatches(),
			//	stringvalidator.ConflictsWith("name"),
			//	stringvalidator.OneOf([]string{"foo", "bar"}...),
			//},
		},
		"blueprint_id": resourceSchema.StringAttribute{},
	}
}

type stringyThing struct {
	s string
}

func (o stringyThing) String() string {
	return o.s
}

func printStringers(s []fmt.Stringer) {
	for i := range s {
		log.Println(s[i])
	}
}

func aSliceOfStringyThings() []*stringyThing {
	var result []*stringyThing
	result = append(result, &stringyThing{s: "foo"})
	result = append(result, &stringyThing{s: "bar"})
	return result
}
