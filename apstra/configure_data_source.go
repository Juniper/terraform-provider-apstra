package tfapstra

import (
	"context"
	"fmt"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

type datasourceWithSetClient interface {
	datasource.DataSourceWithConfigure
	setClient(*apstra.Client)
}

type datasourceWithSetDcBpClientFunc interface {
	datasource.DataSourceWithConfigure
	setBpClientFunc(func(context.Context, string) (*apstra.TwoStageL3ClosClient, error))
}

type datasourceWithSetFfBpClientFunc interface {
	datasource.DataSourceWithConfigure
	setBpClientFunc(func(context.Context, string) (*apstra.FreeformClient, error))
}

func configureDataSource(_ context.Context, ds datasource.DataSourceWithConfigure, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	if ds, ok := ds.(datasourceWithSetClient); ok {
		ds.setClient(pd.client)
	}

	if ds, ok := ds.(datasourceWithSetDcBpClientFunc); ok {
		ds.setBpClientFunc(pd.getTwoStageL3ClosClient)
	}

	if ds, ok := ds.(datasourceWithSetFfBpClientFunc); ok {
		ds.setBpClientFunc(pd.getFreeformClient)
	}
}
