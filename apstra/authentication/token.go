package authentication

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	ephemeralSchema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiTokenDefaultWarning = 60

type ApiToken struct {
	Value       types.String `tfsdk:"value"`
	SessionId   types.String `tfsdk:"session_id"`
	UserName    types.String `tfsdk:"user_name"`
	WarnSeconds types.Int64  `tfsdk:"warn_seconds"`
	ExpiresAt   time.Time    `tfsdk:"-"`
	DoNotLogOut types.Bool   `tfsdk:"do_not_log_out"`
}

func (o ApiToken) EphemeralAttributes() map[string]ephemeralSchema.Attribute {
	return map[string]ephemeralSchema.Attribute{
		"value": ephemeralSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The API token value.",
		},
		"session_id": ephemeralSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The API session ID associated with the token.",
		},
		"user_name": ephemeralSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The user name associated with the session ID.",
		},
		"warn_seconds": ephemeralSchema.Int64Attribute{
			Optional: true,
			Computed: true,
			MarkdownDescription: fmt.Sprintf("Terraform will produce a warning when the token value is "+
				"referenced with less than this amount of time remaining before expiration. Note that "+
				"determination of remaining token lifetime depends on clock sync between the Apstra server and "+
				"the Terraform host. Value `0` disables warnings. Default value is `%d`.", apiTokenDefaultWarning),
			Validators: []validator.Int64{int64validator.AtLeast(0)},
		},
		"do_not_log_out": ephemeralSchema.BoolAttribute{
			Optional: true,
			MarkdownDescription: "By default, tokens / API sessions produced by this resource are invalidated by " +
				"calling Apstra's `logout` API when Terraform invokes `Close` on this resource. Setting this " +
				"attribute to `true` changes that behavior. `logout` will not be called, and the token produced by " +
				"this resource will remain valid until it expires or something else invalidates it.",
		},
	}
}

func (o *ApiToken) LoadApiData(_ context.Context, in string, diags *diag.Diagnostics) {
	parts := strings.Split(in, ".")
	if len(parts) != 3 {
		diags.AddError("unexpected API response", fmt.Sprintf("JWT should have 3 parts, got %d", len(parts)))
		return
	}

	claimsB64 := parts[1] + strings.Repeat("=", (4-len(parts[1])%4)%4) // pad the b64 part as necessary
	claimsBytes, err := base64.StdEncoding.DecodeString(claimsB64)
	if err != nil {
		diags.AddError("failed base64 decoding token claims", err.Error())
		return
	}

	var claims struct {
		Username    string `json:"username"`
		UserSession string `json:"user_session"`
		Expiration  int64  `json:"exp"`
	}
	err = json.Unmarshal(claimsBytes, &claims)
	if err != nil {
		diags.AddError("failed unmarshaling token claims JSON payload", err.Error())
		return
	}

	o.Value = types.StringValue(in)
	o.UserName = types.StringValue(claims.Username)
	o.SessionId = types.StringValue(claims.UserSession)
	o.ExpiresAt = time.Unix(claims.Expiration, 0)
}

func (o *ApiToken) SetDefaults() {
	if o.WarnSeconds.IsNull() {
		o.WarnSeconds = types.Int64Value(apiTokenDefaultWarning)
	}
}

func (o *ApiToken) SetPrivateState(ctx context.Context, ps private.State, diags *diag.Diagnostics) {
	privateEphemeralApiToken := private.EphemeralApiToken{
		Token:         o.Value.ValueString(),
		ExpiresAt:     o.ExpiresAt,
		WarnThreshold: time.Duration(o.WarnSeconds.ValueInt64()) * time.Second,
		DoNotLogOut:   o.DoNotLogOut.ValueBool(),
	}
	privateEphemeralApiToken.SetPrivateState(ctx, ps, diags)
}
