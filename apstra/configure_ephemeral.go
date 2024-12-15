package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

type ephemeralWithSetClient interface {
	ephemeral.EphemeralResourceWithConfigure
	setClient(*apstra.Client)
}

type ephemeralWithSetDcBpClientFunc interface {
	ephemeral.EphemeralResourceWithConfigure
	setBpClientFunc(func(context.Context, string) (*apstra.TwoStageL3ClosClient, error))
}

type ephemeralWithSetFfBpClientFunc interface {
	ephemeral.EphemeralResourceWithConfigure
	setBpClientFunc(func(context.Context, string) (*apstra.FreeformClient, error))
}

func configureEphemeral(_ context.Context, ep ephemeral.EphemeralResourceWithConfigure, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return // cannot continue
	}

	var pd *providerData
	var ok bool

	if pd, ok = req.ProviderData.(*providerData); !ok {
		resp.Diagnostics.AddError(
			errDataSourceConfigureProviderDataSummary,
			fmt.Sprintf(errDataSourceConfigureProviderDataDetail, *pd, req.ProviderData),
		)
	}

	if ep, ok := ep.(ephemeralWithSetClient); ok {
		ep.setClient(pd.client)
	}

	if ep, ok := ep.(ephemeralWithSetDcBpClientFunc); ok {
		ep.setBpClientFunc(pd.getTwoStageL3ClosClient)
	}

	if ep, ok := ep.(ephemeralWithSetFfBpClientFunc); ok {
		ep.setBpClientFunc(pd.getFreeformClient)
	}
}
