package tfapstra

import (
	"context"
	"fmt"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
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

func TestAccDataSourceIpv4Pool_A(t *testing.T) {
	ctx := context.Background()
	pool, deleteFunc, err := testutils.Ipv4PoolA(ctx)
	if err != nil {
		t.Error(err)
		t.Fatal(deleteFunc(ctx))
	}
	defer func() {
		err := deleteFunc(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	resource.Test(t, resource.TestCase{
		//PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByIdHCL, string(pool.Id)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(pool.Id)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", pool.DisplayName),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "status", pool.Status.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "total", pool.Total.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used", pool.Used.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used_percentage", fmt.Sprintf("%0.f", pool.UsedPercentage)),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.#", fmt.Sprintf("%d", len(pool.Subnets))),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.status", pool.Subnets[0].Status),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.total", pool.Subnets[0].Total.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.used", pool.Subnets[0].Used.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.used_percentage", fmt.Sprintf("%0.f", pool.Subnets[0].UsedPercentage)),
				),
			},
			// Read by Name
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByNameHCL, pool.DisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(pool.Id)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", pool.DisplayName),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "status", pool.Status.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "total", pool.Total.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used", pool.Used.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used_percentage", fmt.Sprintf("%0.f", pool.UsedPercentage)),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.#", fmt.Sprintf("%d", len(pool.Subnets))),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.status", pool.Subnets[0].Status),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.total", pool.Subnets[0].Total.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.used", pool.Subnets[0].Used.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.0.used_percentage", fmt.Sprintf("%0.f", pool.Subnets[0].UsedPercentage)),
				),
			},
		},
	})
}

func TestAccDataSourceIpv4Pool_B(t *testing.T) {
	ctx := context.Background()
	pool, deleteFunc, err := testutils.Ipv4PoolB(ctx)
	if err != nil {
		t.Error(err)
		t.Fatal(deleteFunc(ctx))
	}
	defer func() {
		err := deleteFunc(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	resource.Test(t, resource.TestCase{
		//PreCheck:                 setup,
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByIdHCL, string(pool.Id)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(pool.Id)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", pool.DisplayName),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "status", pool.Status.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "total", pool.Total.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used", pool.Used.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used_percentage", fmt.Sprintf("%0.f", pool.UsedPercentage)),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.#", fmt.Sprintf("%d", len(pool.Subnets))),
				),
			},
			// Read by Name
			{
				Config: insecureProviderConfigHCL + fmt.Sprintf(dataSourceIpv4PoolTemplateByNameHCL, pool.DisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source/resource fields match
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "id", string(pool.Id)),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "name", pool.DisplayName),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "status", pool.Status.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "total", pool.Total.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used", pool.Used.String()),
					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "used_percentage", fmt.Sprintf("%0.f", pool.UsedPercentage)),

					resource.TestCheckResourceAttr("data.apstra_ipv4_pool.test", "subnets.#", fmt.Sprintf("%d", len(pool.Subnets))),
				),
			},
		},
	})
}
