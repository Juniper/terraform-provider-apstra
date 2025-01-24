//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const (
	resourceFreeformLinkHcl = `
resource %q %q {
 blueprint_id = %q
  name        = %q
  tags        = %s
  endpoints   = {
    %q = {
      interface_name    = %s
      transformation_id = %s
      ipv4_address      = %s
      ipv6_address      = %s
    },
    %q = {
      interface_name    = %s
      transformation_id = %s
      ipv4_address      = %s
      ipv6_address      = %s
    }
  }
}
`
)

type resourceFreeformLink struct {
	blueprintId string
	name        string
	endpoints   []apstra.FreeformEthernetEndpoint
	tags        []string
}

func (o resourceFreeformLink) render(rType, rName string) string {
	return fmt.Sprintf(resourceFreeformLinkHcl,
		rType, rName,
		o.blueprintId,
		o.name,
		stringSliceOrNull(o.tags),
		o.endpoints[0].SystemId,
		stringPtrOrNull(o.endpoints[0].Interface.Data.IfName),
		intPtrOrNull(o.endpoints[0].Interface.Data.TransformationId),
		cidrOrNull(o.endpoints[0].Interface.Data.Ipv4Address),
		cidrOrNull(o.endpoints[0].Interface.Data.Ipv6Address),
		o.endpoints[1].SystemId,
		stringPtrOrNull(o.endpoints[1].Interface.Data.IfName),
		intPtrOrNull(o.endpoints[1].Interface.Data.TransformationId),
		cidrOrNull(o.endpoints[1].Interface.Data.Ipv4Address),
		cidrOrNull(o.endpoints[1].Interface.Data.Ipv6Address),
	)
}

func (o resourceFreeformLink) testChecks(t testing.TB, rType, rName string, ipAllocationEnabled bool) testChecks {
	result := newTestChecks(rType + "." + rName)

	// required and computed attributes can always be checked
	result.append(t, "TestCheckResourceAttrSet", "id")
	result.append(t, "TestCheckNoResourceAttr", "aggregate_link_id")
	result.append(t, "TestCheckResourceAttr", "blueprint_id", o.blueprintId)
	result.append(t, "TestCheckResourceAttr", "name", o.name)

	if len(o.tags) > 0 {
		result.append(t, "TestCheckResourceAttr", "tags.#", strconv.Itoa(len(o.tags)))
		for _, tag := range o.tags {
			result.append(t, "TestCheckTypeSetElemAttr", "tags.*", tag)
		}
	}

	result.append(t, "TestCheckResourceAttr", "endpoints.%", "2")
	for _, endpoint := range o.endpoints {
		if endpoint.Interface.Data.IfName != nil {
			result.append(t, "TestCheckResourceAttr", "endpoints."+endpoint.SystemId.String()+".interface_name", *endpoint.Interface.Data.IfName)
		} else {
			result.append(t, "TestCheckNoResourceAttr", "endpoints."+endpoint.SystemId.String()+".interface_name")
		}
		if endpoint.Interface.Data.TransformationId != nil {
			result.append(t, "TestCheckResourceAttr", "endpoints."+endpoint.SystemId.String()+".transformation_id", strconv.Itoa(*endpoint.Interface.Data.TransformationId))
		} else {
			result.append(t, "TestCheckNoResourceAttr", "endpoints."+endpoint.SystemId.String()+".transformation_id")
		}
		result.append(t, "TestCheckResourceAttrSet", "endpoints."+endpoint.SystemId.String()+".interface_id")

		if endpoint.Interface.Data.Ipv4Address != nil {
			result.append(t, "TestCheckResourceAttr", "endpoints."+endpoint.SystemId.String()+".ipv4_address", endpoint.Interface.Data.Ipv4Address.String())
		} else {
			if ipAllocationEnabled {
				result.append(t, "TestCheckResourceAttrSet", "endpoints."+endpoint.SystemId.String()+".ipv4_address")
			} else {
				result.append(t, "TestCheckNoResourceAttr", "endpoints."+endpoint.SystemId.String()+".ipv4_address")
			}
		}

		if endpoint.Interface.Data.Ipv6Address != nil {
			result.append(t, "TestCheckResourceAttr", "endpoints."+endpoint.SystemId.String()+".ipv6_address", endpoint.Interface.Data.Ipv6Address.String())
		} else {
			if ipAllocationEnabled {
				result.append(t, "TestCheckResourceAttrSet", "endpoints."+endpoint.SystemId.String()+".ipv6_address")
			} else {
				result.append(t, "TestCheckNoResourceAttr", "endpoints."+endpoint.SystemId.String()+".ipv6_address")
			}
		}
	}

	return result
}

func TestResourceFreeformLinkWithIpAllocationDisabled(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint
	bp, intSysIds, extSysIds := testutils.FfBlueprintB(t, ctx, 2, 2)

	type testStep struct {
		config resourceFreeformLink
	}

	type testCase struct {
		ipAllocationEnabled   bool
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"int_int_start_minimal": {
			steps: []testStep{
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/0"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/0"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/1"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("192.168.10.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::3"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/1"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("192.168.10.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::4"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/3"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/3"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
						},
					},
				},
			},
		},
		"int_int_start_maximal": {
			steps: []testStep{
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/4"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.1.1.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::1"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/4"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.1.1.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::2"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/5"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/5"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/6"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.2.1.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::3"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/6"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.2.1.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::4"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
			},
		},
		"int_ext_start_minimal": {
			steps: []testStep{
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/7"),
										TransformationId: utils.ToPtr(1),
									},
								},
							},
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/7"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("192.168.10.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::3"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("192.168.10.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::4"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/7"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: nil,
										Ipv6Address: nil,
										Tags:        nil,
									},
								},
							},
						},
					},
				},
			},
		},
		"int_ext_start_maximal": {
			steps: []testStep{
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/8"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.1.1.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::1"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("10.1.1.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::2"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/8"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: nil,
										Ipv6Address: nil,
										Tags:        nil,
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/8"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.2.1.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::3"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("10.2.1.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::4"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
			},
		},
		"ext_ext_start_minimal": {
			steps: []testStep{
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{},
								},
							},
							{
								SystemId: extSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("192.168.10.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::3"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("192.168.10.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::4"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{},
								},
							},
							{
								SystemId: extSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{},
								},
							},
						},
					},
				},
			},
		},
		"ext_ext_start_maximal": {
			steps: []testStep{
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("10.1.1.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::1"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: extSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("10.1.1.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::2"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{},
								},
							},
							{
								SystemId: extSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: extSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("10.2.1.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::3"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: extSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										Ipv4Address: &net.IPNet{IP: net.ParseIP("10.2.1.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address: &net.IPNet{IP: net.ParseIP("2001:db8::4"), Mask: net.CIDRMask(64, 128)},
										Tags:        randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformLink)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			t.Parallel()

			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.apiVersionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName, tCase.ipAllocationEnabled)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}

func TestResourceFreeformLinkWithIpAllocationEnabled(t *testing.T) {
	ctx := context.Background()
	client := testutils.GetTestClient(t, ctx)
	apiVersion := version.Must(version.NewVersion(client.ApiVersion()))

	// create a blueprint
	bp, intSysIds, _ := testutils.FfBlueprintB(t, ctx, 2, 2)

	// create an ipv4 allocation group
	ipv4AllocGroupId, err := bp.CreateAllocGroup(ctx, &apstra.FreeformAllocGroupData{
		Name:    acctest.RandString(6),
		Type:    enum.ResourcePoolTypeIpv4,
		PoolIds: []apstra.ObjectId{"Private-10_0_0_0-8"},
	})
	require.NoError(t, err)

	// create an ipv6 allocation group
	ipv6AllocGroupId, err := bp.CreateAllocGroup(ctx, &apstra.FreeformAllocGroupData{
		Name:    acctest.RandString(6),
		Type:    enum.ResourcePoolTypeIpv6,
		PoolIds: []apstra.ObjectId{"Private-fc01-a05-fab-48"},
	})
	require.NoError(t, err)

	// create a resource group
	raGroupId, err := bp.CreateRaGroup(ctx, &apstra.FreeformRaGroupData{
		Label: acctest.RandString(6),
	})
	require.NoError(t, err)

	_, err = bp.CreateResourceGenerator(ctx, &apstra.FreeformResourceGeneratorData{
		ResourceType:    enum.FFResourceTypeIpv4,
		Label:           acctest.RandString(6),
		Scope:           "node(type='link', name='target')",
		AllocatedFrom:   &ipv4AllocGroupId,
		ContainerId:     raGroupId,
		SubnetPrefixLen: utils.ToPtr(31),
	})
	require.NoError(t, err)

	_, err = bp.CreateResourceGenerator(ctx, &apstra.FreeformResourceGeneratorData{
		ResourceType:    enum.FFResourceTypeIpv6,
		Label:           acctest.RandString(6),
		Scope:           "node(type='link', name='target')",
		AllocatedFrom:   &ipv6AllocGroupId,
		ContainerId:     raGroupId,
		SubnetPrefixLen: utils.ToPtr(127),
	})
	require.NoError(t, err)

	type testStep struct {
		config resourceFreeformLink
	}

	type testCase struct {
		ipAllocationEnabled   bool
		apiVersionConstraints version.Constraints
		steps                 []testStep
	}

	testCases := map[string]testCase{
		"int_int_start_minimal_ip_allocation_enabled": {
			ipAllocationEnabled: true,
			steps: []testStep{
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/0"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/0"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/1"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("192.168.10.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::3"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/1"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("192.168.10.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::4"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/3"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/3"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
						},
					},
				},
			},
		},
		"int_int_start_maximal_ip_allocation_enabled": {
			ipAllocationEnabled: true,
			steps: []testStep{
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/4"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.1.1.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::1"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/4"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.1.1.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::2"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/5"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/5"),
										TransformationId: utils.ToPtr(2),
										Ipv4Address:      nil,
										Ipv6Address:      nil,
										Tags:             nil,
									},
								},
							},
						},
					},
				},
				{
					config: resourceFreeformLink{
						blueprintId: bp.Id().String(),
						name:        acctest.RandString(6),
						tags:        randomStrings(rand.Intn(10)+2, 6),
						endpoints: []apstra.FreeformEthernetEndpoint{
							{
								SystemId: intSysIds[0],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/6"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.2.1.1"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::3"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
							{
								SystemId: intSysIds[1],
								Interface: apstra.FreeformInterface{
									Data: &apstra.FreeformInterfaceData{
										IfName:           utils.ToPtr("ge-0/0/6"),
										TransformationId: utils.ToPtr(1),
										Ipv4Address:      &net.IPNet{IP: net.ParseIP("10.2.1.2"), Mask: net.CIDRMask(30, 32)},
										Ipv6Address:      &net.IPNet{IP: net.ParseIP("2001:db8::4"), Mask: net.CIDRMask(64, 128)},
										Tags:             randomStrings(rand.Intn(10)+2, 6),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resourceType := tfapstra.ResourceName(ctx, &tfapstra.ResourceFreeformLink)

	for tName, tCase := range testCases {
		tName, tCase := tName, tCase
		t.Run(tName, func(t *testing.T) {
			// t.Parallel() running these tests in parallel exposes some API races

			if !tCase.apiVersionConstraints.Check(apiVersion) {
				t.Skipf("test case %s requires Apstra %s", tName, tCase.apiVersionConstraints.String())
			}

			steps := make([]resource.TestStep, len(tCase.steps))
			for i, step := range tCase.steps {
				config := step.config.render(resourceType, tName)
				checks := step.config.testChecks(t, resourceType, tName, tCase.ipAllocationEnabled)

				chkLog := checks.string()
				stepName := fmt.Sprintf("test case %q step %d", tName, i+1)

				t.Logf("\n// ------ begin config for %s ------\n%s// -------- end config for %s ------\n\n", stepName, config, stepName)
				t.Logf("\n// ------ begin checks for %s ------\n%s// -------- end checks for %s ------\n\n", stepName, chkLog, stepName)

				steps[i] = resource.TestStep{
					Config: insecureProviderConfigHCL + config,
					Check:  resource.ComposeAggregateTestCheckFunc(checks.checks...),
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
