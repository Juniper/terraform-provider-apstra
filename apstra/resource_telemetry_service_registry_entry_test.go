package tfapstra

import (
	"fmt"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	resourceTelemetryServiceRegistryEntryHCL = `
// resource config
resource "apstra_telemetry_service_registry_entry" "test" {
   description         = "Test Registry Entry"
   service_name        = "%s"
 	application_schema  = jsonencode(%s)
   storage_schema_path = "%s"
  
}
`
	as1 = `{
            properties = {
                key = {
                    properties = {
                        authenticated_vlan = {
                            type = "string"
                        }
                        authorization_status = {
                            type = "string"
                        }
                        fallback_vlan_active = {
                            enum = [
                                "True",
                                "False",
                            ]
                            type = "string"
                        }
                        port_status = {
                            enum = [
                                "authorized",
                                "blocked",
                            ]
                            type = "string"
                        }
                        supplicant_mac = {
                            pattern = "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
                            type    = "string"
                        }
                    }
                    required = [
                        "supplicant_mac",
                        "authenticated_vlan",
                        "authorization_status",
                        "port_status",
                        "fallback_vlan_active",
                    ]
                    type = "object"
                }
                value = {
                    description = "0 in case of blocked, 1 in case of authorized"
                    type        = "integer"
                }
            }
            required = [
                "key",
                "value",
            ]
            type = "object"
        }`
	as2 = `{
            properties = {
                key = {
                    properties = {
                        authenticated_vlan = {
                            type = "string"
                        }
                        authorization_status = {
                            type = "string"
                        }
                        fallback_vlan_active = {
                            enum = [
                                "True",
                                "False",
                            ]
                            type = "string"
                        }
                        port_status = {
                            enum = [
                                "authorized",
                                "blocked",
                            ]
                            type = "string"
                        }
                    }
                    required = [
                        "authenticated_vlan",
                        "authorization_status",
                        "port_status",
                        "fallback_vlan_active",
                    ]
                    type = "object"
                }
                value = {
                    description = "0 in case of blocked, 1 in case of authorized"
                    type        = "string"
                }
            }
            required = [
                "key",
                "value",
            ]
            type = "object"
        }`
	as1_string = `{
			   "properties":{
				  "key":{
					 "properties":{
						"authenticated_vlan":{
						   "type":"string"
						},
						"authorization_status":{
						   "type":"string"
						},
						"fallback_vlan_active":{
						   "enum":[
							  "True",
							  "False"
						   ],
						   "type":"string"
						},
						"port_status":{
						   "enum":[
							  "authorized",
							  "blocked"
						   ],
						   "type":"string"
						},
						"supplicant_mac":{
						   "pattern":"^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$",
						   "type":"string"
						}
					 },
					 "required":[
						"supplicant_mac",
						"authenticated_vlan",
						"authorization_status",
						"port_status",
						"fallback_vlan_active"
					 ],
					 "type":"object"
				  },
				  "value":{
					 "description":"0 in case of blocked, 1 in case of authorized",
					 "type":"integer"
				  }
			   },
			   "required":[
				  "key",
				  "value"
			   ],
			   "type":"object"
			}`
	as2_string = `{
		   "properties":{
			  "key":{
				 "properties":{
					"authenticated_vlan":{
					   "type":"string"
					},
					"authorization_status":{
					   "type":"string"
					},
					"fallback_vlan_active":{
					   "enum":[
						  "True",
						  "False"
					   ],
					   "type":"string"
					},
					"port_status":{
					   "enum":[
						  "authorized",
						  "blocked"
					   ],
					   "type":"string"
					}
				 },
				 "required":[
					"authenticated_vlan",
					"authorization_status",
					"port_status",
					"fallback_vlan_active"
				 ],
				 "type":"object"
			  },
			  "value":{
				 "description":"0 in case of blocked, 1 in case of authorized",
				 "type":"string"
			  }
		   },
		   "required":[
			  "key",
			  "value"
		   ],
		   "type":"object"
		}`
	ss1 = "aos.sdk.telemetry.schemas.iba_integer_data"
	ss2 = "aos.sdk.telemetry.schemas.iba_string_data"
)

func TestAccResourceTelemetryServiceRegistryEntry(t *testing.T) {
	var (
		testAccResourceServiceName         = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		testAccResourceServiceRegistryCfg1 = fmt.Sprintf(resourceTelemetryServiceRegistryEntryHCL, testAccResourceServiceName, as1, ss1)
		testAccResourceServiceRegistryCfg2 = fmt.Sprintf(resourceTelemetryServiceRegistryEntryHCL, testAccResourceServiceName, as2, ss2)
	)

	d := diag.Diagnostics{}
	TestSR1 := func(state string) error {
		if !utils.JSONEqual(types.StringValue(state), types.StringValue(as1_string), &d) {
			return fmt.Errorf("input Data does not match output Input %v. Output %v", as1_string, state)
		}
		return nil
	}

	TestSR2 := func(state string) error {
		if !utils.JSONEqual(types.StringValue(state), types.StringValue(as2_string), &d) {
			return fmt.Errorf("input Data does not match output Input %v. Output %v", as2_string, state)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceServiceRegistryCfg1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name and data
					resource.TestCheckResourceAttr("apstra_telemetry_service_registry_entry.test", "service_name", testAccResourceServiceName),
					resource.TestCheckResourceAttr("apstra_telemetry_service_registry_entry.test", "storage_schema_path", ss1),
					resource.TestCheckResourceAttrWith("apstra_telemetry_service_registry_entry.test", "application_schema", TestSR1),
				),
			},
			// Update and Read testing
			{
				Config: insecureProviderConfigHCL + testAccResourceServiceRegistryCfg2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name and data
					resource.TestCheckResourceAttr("apstra_telemetry_service_registry_entry.test", "service_name", testAccResourceServiceName),
					resource.TestCheckResourceAttr("apstra_telemetry_service_registry_entry.test", "storage_schema_path", ss2),
					resource.TestCheckResourceAttrWith("apstra_telemetry_service_registry_entry.test", "application_schema", TestSR2),
				),
			},
		},
	})
}
