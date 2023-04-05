package utils

import "github.com/Juniper/apstra-go-sdk/apstra"

func FcdModes() []string {
	return []string{
		apstra.FabricConnectivityDesignL3Clos.String(),
		apstra.FabricConnectivityDesignL3Collapsed.String()}
}
