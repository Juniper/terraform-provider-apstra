package tfapstra

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	dataSourceIpv4PoolTemplateByNameHCL = `
data "apstra_ipv4_pool" "test" {
  name = "%s"
}
`

	dataSourceIpv4PoolTemplateByIdHCL = `
data "apstra_ipv4_pool" "test" {
  id = "%s"
}
`
)

func TestAccDataSourceIpv4Pool_basic(t *testing.T) {
	testAccResourceIpv4PoolCfg1Name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	clientCfg, err := NewClientConfig("")
	if err != nil {
		t.Fatal(err)
	}
	clientCfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
	client, err := clientCfg.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	id, err := client.CreateIp4Pool(context.Background(), &apstra.NewIpPoolRequest{
		DisplayName: testAccResourceIpv4PoolCfg1Name,
		Subnets:     []apstra.NewIpSubnet{{Network: "192.168.0.0/16"}},
	})

	defer func() {
		err := client.DeleteIp4Pool(context.Background(), id)
		if err != nil {
			t.Error(err)
		}
	}()

	resource.Test(t, resource.TestCase{
		//PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID testing
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByIdHCL, string(id)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(id)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", testAccResourceIpv4PoolCfg1Name),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "total", "65536"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used_percentage", "0"),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.#", "1"),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.status", "pool_element_available"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.total", "65536"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.used", "0"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.used_percentage", "0"),
				),
			},
			// Read by Name testing
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByNameHCL, testAccResourceIpv4PoolCfg1Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(id)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", testAccResourceIpv4PoolCfg1Name),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "total", "65536"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used_percentage", "0"),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.#", "1"),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.status", "pool_element_available"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.total", "65536"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.used", "0"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.used_percentage", "0"),
				),
			},
		},
	})
}
