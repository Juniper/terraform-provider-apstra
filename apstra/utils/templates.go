package utils

import "bitbucket.org/apstrktr/goapstra"

func AllTemplateTypes() []string {
	templates := goapstra.AllTemplateTypes()
	result := make([]string, len(templates))
	for i := range templates {
		result[i] = templates[i].String()
	}
	return result
}

func AllOverlayControlProtocols() []string {
	overlayControlProtocols := goapstra.AllOverlayControlProtocols()
	result := make([]string, len(overlayControlProtocols))
	for i := range overlayControlProtocols {
		result[i] = overlayControlProtocols[i].String()
	}
	return result
}
