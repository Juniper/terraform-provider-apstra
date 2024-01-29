package compatibility

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"strings"
)

func SupportedApiVersions() []string {
	providerVersions := []string{
		apiversions.Apstra410,
		apiversions.Apstra411,
		apiversions.Apstra412,
		apiversions.Apstra420,
	}

	sdkVersions := apstra.ApstraApiSupportedVersions()

	var result []string
	for i := range providerVersions {
		if sdkVersions.Includes(providerVersions[i]) {
			result = append(result, providerVersions[i])
		}
	}

	return result
}

func SupportedApiVersionsPretty() string {
	supportedVers := SupportedApiVersions()
	stop := len(supportedVers) - 1

	for i := range supportedVers {
		if i == stop {
			supportedVers[i] = "and " + supportedVers[i]
			break
		}
		supportedVers[i] = supportedVers[i] + ","
	}

	return strings.Join(supportedVers, " ")
}
