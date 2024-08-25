package utils

import "github.com/Juniper/apstra-go-sdk/apstra"

func AllTemplateTypes() []string {
	templates := apstra.AllTemplateTypes()
	result := make([]string, len(templates))
	for i := range templates {
		result[i] = templates[i].String()
	}
	return result
}

func AllOverlayControlProtocols() []string {
	overlayControlProtocols := apstra.AllOverlayControlProtocols()
	result := make([]string, len(overlayControlProtocols))
	for i := range overlayControlProtocols {
		result[i] = overlayControlProtocols[i].String()
	}
	return result
}
