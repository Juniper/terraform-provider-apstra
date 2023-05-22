package apstradefault

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DefaultAny interface {
	defaults.Bool
	defaults.Int64
	defaults.Float64
	defaults.List
	defaults.Map
	defaults.Number
	defaults.Object
	defaults.Set
	defaults.String
}

var _ DefaultAny = StaticAnyTypeDefaulter{}

type StaticAnyTypeDefaulter struct {
	value attr.Value
}

func (o StaticAnyTypeDefaulter) Description(_ context.Context) string {
	return fmt.Sprintf("set the default value to a %T with value %q", o.value, o.value)
}

func (o StaticAnyTypeDefaulter) MarkdownDescription(ctx context.Context) string {
	return o.MarkdownDescription(ctx)
}

func (o StaticAnyTypeDefaulter) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = o.value.(types.Bool)
}

func (o StaticAnyTypeDefaulter) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = o.value.(types.Int64)
}

func (o StaticAnyTypeDefaulter) DefaultFloat64(_ context.Context, _ defaults.Float64Request, resp *defaults.Float64Response) {
	resp.PlanValue = o.value.(types.Float64)
}

func (o StaticAnyTypeDefaulter) DefaultList(_ context.Context, _ defaults.ListRequest, resp *defaults.ListResponse) {
	resp.PlanValue = o.value.(types.List)
}

func (o StaticAnyTypeDefaulter) DefaultMap(_ context.Context, _ defaults.MapRequest, resp *defaults.MapResponse) {
	resp.PlanValue = o.value.(types.Map)
}

func (o StaticAnyTypeDefaulter) DefaultNumber(_ context.Context, _ defaults.NumberRequest, resp *defaults.NumberResponse) {
	resp.PlanValue = o.value.(types.Number)
}

func (o StaticAnyTypeDefaulter) DefaultObject(_ context.Context, _ defaults.ObjectRequest, resp *defaults.ObjectResponse) {
	resp.PlanValue = o.value.(types.Object)
}

func (o StaticAnyTypeDefaulter) DefaultSet(_ context.Context, _ defaults.SetRequest, resp *defaults.SetResponse) {
	resp.PlanValue = o.value.(types.Set)
}

func (o StaticAnyTypeDefaulter) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = o.value.(types.String)
}

func StaticDefaultAny(v attr.Value) StaticAnyTypeDefaulter {
	return StaticAnyTypeDefaulter{
		value: v,
	}
}
