package private

import (
	"context"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ResourceDatacenterTag struct {
	Id apstra.ObjectId `json:"id"`
}

func (o *ResourceDatacenterTag) LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	loadPrivateState(ctx, ps, o, diags)
}

func (o ResourceDatacenterTag) SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics) {
	setPrivateState(ctx, ps, o, diags)
}
