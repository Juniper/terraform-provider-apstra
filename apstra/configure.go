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
		errResourceConfigureProviderDataDetail,
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
		errResourceConfigureProviderDataDetail,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return ""
}

func ResourceGetBlueprintLockFunc(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) func(context.Context, *goapstra.TwoStageL3ClosMutex) error {
	if req.ProviderData == nil {
		return nil
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.bpLockFunc
	}

	resp.Diagnostics.AddError(
		errResourceConfigureProviderDataDetail,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return nil
}

func ResourceGetBlueprintUnlockFunc(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) func(context.Context, goapstra.ObjectId) error {
	if req.ProviderData == nil {
		return nil
	}

	var pd *providerData
	var ok bool
	if pd, ok = req.ProviderData.(*providerData); ok {
		return pd.bpUnlockFunc
	}

	resp.Diagnostics.AddError(
		errResourceConfigureProviderDataDetail,
		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
	)
	return nil
}

//func ResourceGetMutexes(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) *[]goapstra.TwoStageL3ClosMutex {
//	if req.ProviderData == nil {
//		return nil
//	}
//
//	var pd *providerData
//	var ok bool
//	if pd, ok = req.ProviderData.(*providerData); ok {
//		return pd.mutexes
//	}
//
//	resp.Diagnostics.AddError(
//		errResourceConfigureProviderDataDetail,
//		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
//	)
//	return nil
//}
//
//func ResourceGetProviderUUID(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) uuid.UUID {
//	if req.ProviderData == nil {
//		return uuid.UUID{}
//	}
//
//	var pd *providerData
//	var ok bool
//	if pd, ok = req.ProviderData.(*providerData); ok {
//		return pd.uuid
//	}
//
//	resp.Diagnostics.AddError(
//		errResourceConfigureProviderDataDetail,
//		fmt.Sprintf(errResourceConfigureProviderDataDetail, *pd, req.ProviderData),
//	)
//	return uuid.UUID{}
//}
