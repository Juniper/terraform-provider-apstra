package apstra

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type versionValidator interface {
	apiVersion() (*version.Version, error)
	cfgVersionMin() (*version.Version, error)
	cfgVersionMax() (*version.Version, error)
}

func checkVersionCompatibility(vv versionValidator, path path.Path, diags *diag.Diagnostics) {
	apiVersion, err := vv.apiVersion()
	if err != nil {
		diags.AddError(errProviderBug, fmt.Sprintf("error determining API version - %s", err.Error()))
		return
	}
	if apiVersion == nil {
		diags.AddError(errProviderBug, "attempt to verify API version compatibility with nil pointer")
	}

	cfgVersionMin, err := vv.cfgVersionMin()
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error determining minimum API version required by configuration - %s",
				err.Error()))
		return
	}

	cfgVersionMax, err := vv.cfgVersionMax()
	if err != nil {
		diags.AddError(errProviderBug,
			fmt.Sprintf("error determining maximum API version required by configuration - %s",
				err.Error()))
		return
	}

	if cfgVersionMin != nil && apiVersion.LessThan(cfgVersionMin) {
		diags.AddAttributeError(path, errApiCompatibility,
			fmt.Sprintf("API version (%q) is less than minimum required by configuration (%q)",
				apiVersion.String(), cfgVersionMin.String()))
		return
	}

	if cfgVersionMax != nil && apiVersion.LessThan(cfgVersionMax) {
		diags.AddAttributeError(path, errApiCompatibility,
			fmt.Sprintf("API version (%q) is less than maximum required by configuration (%q)",
				apiVersion.String(), cfgVersionMax.String()))
		return
	}
}
