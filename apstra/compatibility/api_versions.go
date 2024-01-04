package compatibility

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/go-version"
	"strings"
)

const (
	minVerForVnL3Mtu = "4.2.0"
)

func MinVerForVnL3Mtu() *version.Version {
	v, _ := version.NewVersion(minVerForVnL3Mtu) // this will not error
	return v
}

func SupportedApiVersions() []string {
	us := []string{
		"4.1.0",
		"4.1.1",
		"4.1.2",
		"4.2.0",
	}
	them := apstra.ApstraApiSupportedVersions()

	var result []string
	for i := range us {
		if them.Includes(us[i]) {
			result = append(result, us[i])
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
