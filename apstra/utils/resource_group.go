package utils

import "bitbucket.org/apstrktr/goapstra"

func AllResourceGroupNameStrings() []string {
	argn := goapstra.AllResourceGroupNames()
	result := make([]string, len(argn))
	for i := range argn {
		result[i] = argn[i].String()
	}
	return result
}
