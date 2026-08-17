package tfapstra

import (
	"context"

	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type resourceTag struct {
	resourceDesignTag
}

func (o *resourceTag) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (o *resourceTag) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: "This resource will be deprecated in a future release, no earlier than v2.0.0. " +
			"Users are encouraged to migrate their configurations to use `apstra_design_tag`, which is a drop-in replacement.",
		MarkdownDescription: docCategoryDesign + "This resource creates a Tag in the Apstra Design tab.",
		Attributes:          design.Tag{}.ResourceAttributes(),
	}
}
