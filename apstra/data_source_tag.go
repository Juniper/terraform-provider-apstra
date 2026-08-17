package tfapstra

import (
	"context"

	"github.com/Juniper/terraform-provider-apstra/apstra/design"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type dataSourceTag struct {
	dataSourceDesignTag
}

func (o *dataSourceTag) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (o *dataSourceTag) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: "This resource will be deprecated in a future release, no earlier than v2.0.0. " +
			"Users are encouraged to migrate their configurations to use `apstra_design_tag`, which is a drop-in replacement.",
		MarkdownDescription: docCategoryDesign + "This data source provides details of a specific Tag.\n\n" +
			"At least one optional attribute is required.",
		Attributes: design.Tag{}.DataSourceAttributes(),
	}
}
