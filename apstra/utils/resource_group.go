package utils

import "github.com/Juniper/apstra-go-sdk/apstra"

func AllResourceGroupNameStrings() []string {
	argn := apstra.AllResourceGroupNames()
	result := make([]string, len(argn))
	for i := range argn {
		result[i] = argn[i].String()
	}
	return result
}
