package tfapstra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/raw"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure = &resourceRawJSON{}
	_ resourceWithSetClient          = &resourceRawJSON{}
)

type resourceRawJSON struct {
	client *apstra.Client
}

func (o *resourceRawJSON) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_raw_json"
}

func (o *resourceRawJSON) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(ctx, o, req, resp)
}

func (o *resourceRawJSON) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryFootGun + "**!!! Warning !!!**\n" +
			"This is resource is intended only to solve problems not addressed by the normal resources." +
			"Its use is discouraged and not supported. You're on your own with this thing.\n" +
			"**!!! Warning !!!**\n\n" +
			"This resource creates an object from a raw JSON payload via `POST` request. It assumes that the API will " +
			"respond with a payload containing the new object ID: `{\"id\": \"xxxxxxxx\"}`. Config drift detection is " +
			"not implemented, but update-in-place should be possible.. The `Update()` and `Delete()` functions append " +
			"the ID (`/xxxxxxxx`) to the URL.",
		Attributes: raw.RawJSON{}.ResourceAttributes(),
	}
}

func (o *resourceRawJSON) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan raw.RawJSON
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var idResponse struct {
		Id *string `json:"id"`
	}

	u, err := url.Parse(plan.URL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse URL %q", plan.URL.ValueString()), err.Error())
		return
	}

	err = o.client.DoRawJsonTransaction(ctx, apstra.RawJsonRequest{
		Method:  http.MethodPost,
		Url:     u,
		Payload: json.RawMessage(plan.Payload.ValueString()),
	}, &idResponse)
	if err != nil {
		resp.Diagnostics.AddError("Error creating raw JSON object", err.Error())
		return
	}

	if plan.ID.IsUnknown() {
		plan.ID = types.StringPointerValue(idResponse.Id)
	}

	if plan.ID.IsNull() {
		resp.Diagnostics.AddWarning(
			"ID is null",
			"creation did not produce an error, but no ID was specified in the configuration and we failed to find one in the API response",
		)
	}

	resp.State.Set(ctx, &plan)
}

func (o *resourceRawJSON) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state raw.RawJSON
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	u, err := url.Parse(state.URL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse URL %q", state.URL.ValueString()), err.Error())
		return
	}

	err = o.client.DoRawJsonTransaction(ctx, apstra.RawJsonRequest{
		Method: http.MethodGet,
		Url:    u,
	}, nil)
	if err != nil {
		var ace apstra.ClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("error reading raw JSON object", err.Error())
		return
	}

	resp.State.Set(ctx, &state)
}

func (o *resourceRawJSON) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan raw.RawJSON
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.IsNull() {
		resp.Diagnostics.AddWarning("cannot update raw JSON object", "ID is null -- cannot update")
		return
	}

	u, err := url.Parse(fmt.Sprintf("%s/%s", plan.URL.ValueString(), plan.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse URL %q", plan.URL.ValueString()), err.Error())
		return
	}

	err = o.client.DoRawJsonTransaction(ctx, apstra.RawJsonRequest{
		Method:  plan.UpdateMethod.ValueString(),
		Url:     u,
		Payload: json.RawMessage(plan.Payload.ValueString()),
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating raw JSON object", err.Error())
		return
	}

	resp.State.Set(ctx, &plan)
}

func (o *resourceRawJSON) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state raw.RawJSON
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() {
		resp.Diagnostics.AddWarning("cannot delete raw JSON object", "ID is null -- cannot update")
		return
	}

	u, err := url.Parse(fmt.Sprintf("%s/%s", state.URL.ValueString(), state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("cannot parse URL %q", state.URL.ValueString()), err.Error())
		return
	}

	err = o.client.DoRawJsonTransaction(ctx, apstra.RawJsonRequest{
		Method: http.MethodDelete,
		Url:    u,
	}, nil)
	if err != nil {
		var ace apstra.ClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
			return // 404 is okay
		}

		resp.Diagnostics.AddError("error deleting raw JSON object", err.Error())
		return
	}
}

func (o *resourceRawJSON) setClient(client *apstra.Client) {
	o.client = client
}
