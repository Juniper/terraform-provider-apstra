package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func DataSourceGetClient(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *goapstra.Client {
	if req.ProviderData == nil {
		return nil
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.client
	}

	resp.Diagnostics.AddError(
		errDataSourceConfigureProviderDataDetail,
		fmt.Sprintf(errDataSourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return nil
}

func ResourceGetClient(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) *goapstra.Client {
	if req.ProviderData == nil {
		return nil
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.client
	}

	resp.Diagnostics.AddError(
		errResourceConfigureProviderDataDetail,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return nil
}
