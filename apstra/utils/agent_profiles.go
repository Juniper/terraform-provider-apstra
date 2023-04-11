package utils

import "github.com/Juniper/apstra-go-sdk/apstra"

func AgentProfilePlatforms() []string {
	return []string{
		apstra.AgentPlatformNXOS.String(),
		apstra.AgentPlatformJunos.String(),
		apstra.AgentPlatformEOS.String(),
	}
}
