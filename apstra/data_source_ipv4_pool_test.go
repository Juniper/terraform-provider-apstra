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

func testClient(t *testing.T) *apstra.Client {
	clientCfg, err := NewClientConfig("")
	if err != nil {
		t.Fatal(err)
	}
	clientCfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true

	client, err := clientCfg.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func TestAccDataSourceIpv4Pool_ById(t *testing.T) {
	client := testClient(t)

	name1 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	name2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	id1, err := client.CreateIp4Pool(context.Background(), &apstra.NewIpPoolRequest{
		DisplayName: name1,
		Subnets: []apstra.NewIpSubnet{
			{Network: "192.168.0.0/16"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	id2, err := client.CreateIp4Pool(context.Background(), &apstra.NewIpPoolRequest{
		DisplayName: name2,
		Subnets: []apstra.NewIpSubnet{
			{Network: "192.168.0.0/24"},
			{Network: "192.168.1.0/24"},
			{Network: "192.168.2.0/23"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = client.DeleteIp4Pool(context.Background(), id1)
		if err != nil {
			t.Error(err)
		}
		err = client.DeleteIp4Pool(context.Background(), id2)
		if err != nil {
			t.Error(err)
		}
	}()

	resource.Test(t, resource.TestCase{
		//PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID testing 1
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByIdHCL, string(id1)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(id1)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", name1),
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
			// Read by ID testing 2
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByIdHCL, id2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(id2)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", name2),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "total", "1024"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used_percentage", "0"),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.#", "3"),
				),
			},
			// Read by Name testing 1
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByNameHCL, name1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(id1)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", name1),
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
			// Read by Name testing 2
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByNameHCL, name2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(id2)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", name2),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "status", "not_in_use"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "total", "1024"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used", "0"),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used_percentage", "0"),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.#", "3"),
				),
			},
		},
	})
}
