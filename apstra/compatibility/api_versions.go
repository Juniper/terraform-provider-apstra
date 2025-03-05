package compatibility

import (
	"sort"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra/compatibility"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
)

func SupportedApiVersions() []string {
	providerVersions := []string{
		apiversions.Apstra420,
		apiversions.Apstra421,
		apiversions.Apstra4211,
		apiversions.Apstra422,
		apiversions.Apstra500,
		apiversions.Apstra501,
		apiversions.Apstra510,
	}

	sdkVersions := compatibility.SupportedApiVersions()

	bothVersions := utils.SliceIntersectionOfAB(providerVersions, sdkVersions)

	sort.Slice(bothVersions, func(i, j int) bool {
		iv := version.Must(version.NewVersion(bothVersions[i]))
		jv := version.Must(version.NewVersion(bothVersions[j]))
		return iv.LessThan(jv)
	})

	return bothVersions
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
