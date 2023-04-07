package compatibility

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"strings"
)

func SupportedApiVersions() []string {
	us := []string{
		"4.1.0",
		"4.1.1",
		"4.1.2",
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
