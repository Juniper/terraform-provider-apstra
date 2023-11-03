package tfapstra_test

import (
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	insecureProviderConfigHCL = `
provider "apstra" {
  tls_validation_disabled = true
  blueprint_mutex_enabled = false
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"apstra": providerserver.NewProtocol6WithError(tfapstra.NewProvider()),
	}
)
