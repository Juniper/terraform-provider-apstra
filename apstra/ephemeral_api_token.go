package tfapstra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/authentication"
	"github.com/Juniper/terraform-provider-apstra/apstra/private"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
)

var (
	_ ephemeral.EphemeralResource              = (*ephemeralToken)(nil)
	_ ephemeral.EphemeralResourceWithClose     = (*ephemeralToken)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*ephemeralToken)(nil)
	_ ephemeral.EphemeralResourceWithRenew     = (*ephemeralToken)(nil)
	_ ephemeralWithSetClient                   = (*ephemeralToken)(nil)
)

type ephemeralToken struct {
	client *apstra.Client
}

func (o *ephemeralToken) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_token"
}

func (o *ephemeralToken) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCategoryAuthentication + "This Ephemeral Resource retrieves a unique API token and (optionally) invalidates it on Close.",
		Attributes:          authentication.ApiToken{}.EphemeralAttributes(),
	}
}

func (o *ephemeralToken) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	configureEphemeral(ctx, o, req, resp)
}

func (o *ephemeralToken) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var config authentication.ApiToken
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set default values
	config.SetDefaults()

	// create a new client using the credentials in the embedded client's config
	client, err := o.client.Config().NewClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error creating new client", err.Error())
		return
	}

	// log in so that the new client fetches an API token
	err = client.Login(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error logging in new client", err.Error())
		return
	}

	// extract the token
	token := client.GetApiToken()
	if token == "" {
		resp.Diagnostics.AddError("requested API token is empty", "requested API token is empty")
		return
	}

	// destroy the new client without invalidating the API token we just collected.
	// We use Logout() here because it stops the task monitor goroutine.
	client.SetApiToken("")
	err = client.Logout(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error logging out client", err.Error())
		return
	}

	config.LoadApiData(ctx, token, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// sanity check the token lifetime
	now := time.Now()
	if now.After(config.ExpiresAt) {
		resp.Diagnostics.AddError(
			"Just-fetched API token is expired",
			fmt.Sprintf("Token expired at: %s. Current time is: %s", config.ExpiresAt, now),
		)
		return
	}

	// warn the user about imminent expiration
	warn := time.Duration(config.WarnSeconds.ValueInt64()) * time.Second
	if now.Add(warn).After(config.ExpiresAt) {
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("API token expires within %d second warning threshold", config.WarnSeconds),
			fmt.Sprintf("API token expires at %s. Current time: %s", config.ExpiresAt, now),
		)
	}

	// save the private state
	config.SetPrivateState(ctx, resp.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// set the renew timestamp to the early warning time
	resp.RenewAt = config.ExpiresAt.Add(-1 * warn)

	// set the result
	resp.Diagnostics.Append(resp.Result.Set(ctx, &config)...)
}

func (o *ephemeralToken) Renew(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
	var privateEphemeralApiToken private.EphemeralApiToken
	privateEphemeralApiToken.LoadPrivateState(ctx, req.Private, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	now := time.Now()
	if now.After(privateEphemeralApiToken.ExpiresAt) {
		resp.Diagnostics.AddError(
			"API token has expired",
			fmt.Sprintf("Token expired at: %s. Current time is: %s", privateEphemeralApiToken.ExpiresAt, now),
		)
		return
	}

	if now.Add(privateEphemeralApiToken.WarnThreshold).After(privateEphemeralApiToken.ExpiresAt) {
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("API token expires within %d second warning threshold", privateEphemeralApiToken.WarnThreshold),
			fmt.Sprintf("API token expires at %s. Current time: %s", privateEphemeralApiToken.ExpiresAt, now),
		)
	}

	// set the renew timestamp to the expiration time
	resp.RenewAt = privateEphemeralApiToken.ExpiresAt
}

func (o *ephemeralToken) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	// extract the private state data
	var privateEphemeralApiToken private.EphemeralApiToken
	privateEphemeralApiToken.LoadPrivateState(ctx, req.Private, &resp.Diagnostics)

	if privateEphemeralApiToken.DoNotLogOut {
		return // user doesn't want the token invalidated, so there's nothing to do
	}

	if time.Now().After(privateEphemeralApiToken.ExpiresAt) {
		return // token has already expired, so there's nothing to do
	}

	// create a new client based on the embedded client's config
	client, err := o.client.Config().NewClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error creating new client", err.Error())
		return
	}

	// copy the API token from private state into the new client
	client.SetApiToken(privateEphemeralApiToken.Token)

	// log out the client using the swapped-in token
	err = client.Logout(ctx)
	if err != nil {
		var ace apstra.ClientErr
		if errors.As(err, &ace) && ace.Type() == apstra.ErrAuthFail {
			return // 401 is okay
		}

		resp.Diagnostics.AddError("Error while logging out the API key", err.Error())
		return
	}
}

func (o *ephemeralToken) setClient(client *apstra.Client) {
	o.client = client
}
