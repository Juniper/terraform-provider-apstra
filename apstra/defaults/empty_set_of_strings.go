package apstradefault

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ defaults.Set = emptyStringSetDefaulter{}

type emptyStringSetDefaulter struct{}

func (o emptyStringSetDefaulter) Description(_ context.Context) string {
	return "an empty set of types.String"
}

func (o emptyStringSetDefaulter) MarkdownDescription(ctx context.Context) string {
	return o.MarkdownDescription(ctx)
}

func (o emptyStringSetDefaulter) DefaultSet(_ context.Context, _ defaults.SetRequest, resp *defaults.SetResponse) {
	resp.PlanValue = types.SetValueMust(types.StringType, []attr.Value{})
}

func EmptyStringSet() defaults.Set {
	return emptyStringSetDefaulter{}
}
