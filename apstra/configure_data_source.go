package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type datasourceWithSetClient interface {
	datasource.DataSourceWithConfigure
	setClient(*apstra.Client)
}

type datasourceWithSetBpClientFunc interface {
	datasource.DataSourceWithConfigure
	setBpClientFunc(func(context.Context, string) (*apstra.TwoStageL3ClosClient, error))
}

type resourceWithSetClient interface {
	resource.ResourceWithConfigure
	setClient(*apstra.Client)
}

type resourceWithSetBpClientFunc interface {
	resource.ResourceWithConfigure
	setBpClientFunc(func(context.Context, string) (*apstra.TwoStageL3ClosClient, error))
}

type resourceWithSetBpLockFunc interface {
	resource.ResourceWithConfigure
	setBpLockFunc(func(context.Context, string) error)
}

type resourceWithSetBpUnlockFunc interface {
	resource.ResourceWithConfigure
	setBpUnlockFunc(func(context.Context, string) error)
}

type resourceWithSetProviderVersion interface {
	resource.ResourceWithConfigure
	setProviderVersion(string)
}

type resourceWithSetTerraformVersion interface {
	resource.ResourceWithConfigure
	setTerraformVersion(string)
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

	if ds, ok := ds.(datasourceWithSetBpClientFunc); ok {
		ds.setBpClientFunc(pd.getTwoStageL3ClosClient)
	}
}

func configureResource(_ context.Context, rs resource.ResourceWithConfigure, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return // cannot continue
	}

	var pd *providerData
	var ok bool

	if pd, ok = req.ProviderData.(*providerData); !ok {
		resp.Diagnostics.AddError(
			errResourceConfigureProviderDataSummary,
			fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
		)
	}

	if rs, ok := rs.(resourceWithSetClient); ok {
		rs.setClient(pd.client)
	}

	if rs, ok := rs.(resourceWithSetBpClientFunc); ok {
		rs.setBpClientFunc(pd.getTwoStageL3ClosClient)
	}

	if rs, ok := rs.(resourceWithSetBpLockFunc); ok {
		rs.setBpLockFunc(pd.bpLockFunc)
	}

	if rs, ok := rs.(resourceWithSetBpUnlockFunc); ok {
		rs.setBpUnlockFunc(pd.bpUnlockFunc)
	}

	if rs, ok := rs.(resourceWithSetProviderVersion); ok {
		rs.setProviderVersion(pd.providerVersion)
	}

	if rs, ok := rs.(resourceWithSetTerraformVersion); ok {
		rs.setTerraformVersion(pd.terraformVersion)
	}
}
