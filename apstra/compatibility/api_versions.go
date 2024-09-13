package compatibility

import (
	"github.com/Juniper/apstra-go-sdk/apstra/compatibility"
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"sort"
	"strings"
)

func SupportedApiVersions() []string {
	providerVersions := []string{
		apiversions.Apstra420,
		apiversions.Apstra421,
		apiversions.Apstra4211,
		apiversions.Apstra422,
	}

	sdkVersions := compatibility.SupportedApiVersions()

	return utils.SliceIntersectionOfAB(providerVersions, sdkVersions)
}

func SupportedApiVersionsPretty() string {
	supportedVers := SupportedApiVersions()
	sort.Slice(supportedVers, func(i, j int) bool {
		iv := version.Must(version.NewVersion(supportedVers[i]))
		jv := version.Must(version.NewVersion(supportedVers[j]))
		return iv.LessThan(jv)
	})

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
