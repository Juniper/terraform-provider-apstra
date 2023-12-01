package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func DataSourceGetClient(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *apstra.Client {
	if req.ProviderData == nil {
		return nil
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.client
	}

	resp.Diagnostics.AddError(
		errDataSourceConfigureProviderDataSummary,
		fmt.Sprintf(errDataSourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return nil
}

func ResourceGetClient(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) *apstra.Client {
	if req.ProviderData == nil {
		return nil
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.client
	}

	resp.Diagnostics.AddError(
		errResourceConfigureProviderDataSummary,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return nil
}

func ResourceGetProviderVersion(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) string {
	if req.ProviderData == nil {
		return ""
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.providerVersion
	}

	resp.Diagnostics.AddError(
		errResourceConfigureProviderDataSummary,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return ""
}

func ResourceGetTerraformVersion(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) string {
	if req.ProviderData == nil {
		return ""
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.terraformVersion
	}

	resp.Diagnostics.AddError(
		errResourceConfigureProviderDataSummary,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return ""
}

func ResourceGetBlueprintLockFunc(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) func(context.Context, string) error {
	if req.ProviderData == nil {
		return nil
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.bpLockFunc
	}

	resp.Diagnostics.AddError(
		errResourceConfigureProviderDataSummary,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return nil
}

func ResourceGetBlueprintUnlockFunc(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) func(context.Context, string) error {
	if req.ProviderData == nil {
		return nil
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.bpUnlockFunc
	}

	resp.Diagnostics.AddError(
		errResourceConfigureProviderDataSummary,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return nil
}
