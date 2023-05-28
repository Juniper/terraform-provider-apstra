package tfapstra

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	insecureProviderConfigHCL = `
provider "apstra" {
  tls_validation_disabled = true
  blueprint_mutex_disabled = true
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"apstra": providerserver.NewProtocol6WithError(NewProvider()),
	}
)
