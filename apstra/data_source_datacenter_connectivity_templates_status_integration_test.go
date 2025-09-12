//go:build integration

package tfapstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/apstra-go-sdk/enum"
	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	testutils "github.com/Juniper/terraform-provider-apstra/apstra/test_utils"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

const datasourceDatacneterConnectivityTemplatesStatus = `
data %q %q {
  blueprint_id              = %q
}
`

func TestDatasourceDatacenterConnectivityTemplatesStatus(t *testing.T) {
	ctx := context.Background()

	serverFacingPortIds := func(t *testing.T, ctx context.Context, bp *apstra.TwoStageL3ClosClient) []apstra.ObjectId {
		t.Helper()

		query := new(apstra.PathQuery).
			SetBlueprintId(bp.Id()).
			SetClient(bp.Client()).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "system_type", Value: apstra.QEStringVal("switch")},
			}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeInterface.QEEAttribute(),
				{Key: "name", Value: apstra.QEStringVal("server_interface")},
			}).
			Out([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
			Node([]apstra.QEEAttribute{apstra.NodeTypeLink.QEEAttribute()}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeLink.QEEAttribute()}).
			Node([]apstra.QEEAttribute{apstra.NodeTypeInterface.QEEAttribute()}).
			In([]apstra.QEEAttribute{apstra.RelationshipTypeHostedInterfaces.QEEAttribute()}).
			Node([]apstra.QEEAttribute{
				apstra.NodeTypeSystem.QEEAttribute(),
				{Key: "system_type", Value: apstra.QEStringVal("server")},
			})

		var result struct {
			Items []struct {
				Interface struct {
					Id apstra.ObjectId `json:"id"`
				} `json:"server_interface""`
			} `json:"items"`
		}

		err := query.Do(ctx, &result)
		require.NoError(t, err)

		applicationPointIds := make([]apstra.ObjectId, len(result.Items))
		for i, item := range result.Items {
			applicationPointIds[i] = item.Interface.Id
		}

		return applicationPointIds
	}

	newCt := func(t testing.TB, ctx context.Context, bp *apstra.TwoStageL3ClosClient, tags []string, vlan int, typeName string, assignmentCount int) (apstra.ObjectId, [][]string) {
		t.Helper()

		var vlanId *apstra.Vlan
		if vlan > 0 { // with vlan 0 we send nil pointer to create an invalid CT
			vlanId = utils.ToPtr(apstra.Vlan(vlan))
		}

		// create a security zone unique for each CT
		szName := acctest.RandString(6)
		szId, err := bp.CreateSecurityZone(ctx, &apstra.SecurityZoneData{
			Label:   szName,
			SzType:  apstra.SecurityZoneTypeEVPN,
			VrfName: szName,
			VlanId:  vlanId,
		})
		require.NoError(t, err)

		// create the CT
		ct := apstra.ConnectivityTemplate{
			Label:       acctest.RandString(6),
			Description: acctest.RandString(10),
			Tags:        tags,
			Subpolicies: []*apstra.ConnectivityTemplatePrimitive{
				{
					Attributes: &apstra.ConnectivityTemplatePrimitiveAttributesAttachLogicalLink{
						SecurityZone: &szId,
						Tagged:       true,
						Vlan:         vlanId,
					},
				},
			},
		}
		require.NoError(t, ct.SetIds())
		require.NoError(t, ct.SetUserData())
		require.NoError(t, bp.CreateConnectivityTemplate(ctx, &ct))

		// predict the "status" attribute
		var status enum.EndpointPolicyStatus
		switch {
		case assignmentCount > 0:
			status = enum.EndpointPolicyStatusAssigned
		case vlan == 0:
			status = enum.EndpointPolicyStatusIncomplete
		default:
			status = enum.EndpointPolicyStatusReady
		}

		// return arguments to the test check append function
		var testCheckAppendArgs [][]string
		testCheckAppendArgs = append(testCheckAppendArgs, []string{"TestCheckResourceAttr", fmt.Sprintf("connectivity_templates.%s.id", *ct.Id), ct.Id.String()})
		testCheckAppendArgs = append(testCheckAppendArgs, []string{"TestCheckResourceAttr", fmt.Sprintf("connectivity_templates.%s.name", *ct.Id), ct.Label})
		testCheckAppendArgs = append(testCheckAppendArgs, []string{"TestCheckResourceAttr", fmt.Sprintf("connectivity_templates.%s.description", *ct.Id), ct.Description})
		testCheckAppendArgs = append(testCheckAppendArgs, []string{"TestCheckResourceAttr", fmt.Sprintf("connectivity_templates.%s.assignment_count", *ct.Id), strconv.Itoa(assignmentCount)})
		testCheckAppendArgs = append(testCheckAppendArgs, []string{"TestCheckResourceAttr", fmt.Sprintf("connectivity_templates.%s.status", *ct.Id), status.String()})
		testCheckAppendArgs = append(testCheckAppendArgs, []string{"TestCheckResourceAttr", fmt.Sprintf("connectivity_templates.%s.tags.#", *ct.Id), strconv.Itoa(len(ct.Tags))})
		for _, tag := range ct.Tags {
			testCheckAppendArgs = append(testCheckAppendArgs, []string{"TestCheckTypeSetElemAttr", fmt.Sprintf("connectivity_templates.%s.tags.*", *ct.Id), tag})
		}

		return *ct.Id, testCheckAppendArgs
	}

	// create a blueprint
	bp := testutils.BlueprintA(t, ctx)

	// determine application point IDs
	portIds := serverFacingPortIds(t, ctx, bp)

	// strings we'll need in the terraform config and tests
	datasourceType := tfapstra.DatasourceName(ctx, &tfapstra.DataSourceDatacenterConnectivityTemplatesStatus)
	datasourceName := acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)
	datasourceTypeName := fmt.Sprintf("data.%s.%s", datasourceType, datasourceName)

	// create a test check bundle
	checks := newTestChecks(datasourceTypeName)
	checks.append(t, "TestCheckResourceAttr", "blueprint_id", bp.Id().String())
	checks.append(t, "TestCheckResourceAttr", "connectivity_templates.%", "3")

	// unique vlan IDs to use with valid CTs
	vlanIds := randIntSet(t, 10, 4000, 2)

	// create an invalid CT and add its tests
	_, checkArgs := newCt(t, ctx, bp, randomStrings(rand.Intn(10)+1, 6), 0, datasourceTypeName, 0)
	for _, args := range checkArgs {
		checks.append(t, args[0], args[1:]...)
	}

	// create a valid CT without applying it and add its tests
	_, checkArgs = newCt(t, ctx, bp, randomStrings(rand.Intn(10)+1, 6), vlanIds[0], datasourceTypeName, 0)
	for _, args := range checkArgs {
		checks.append(t, args[0], args[1:]...)
	}

	// create a valid CT and apply it and add its tests
	ctId, checkArgs := newCt(t, ctx, bp, randomStrings(rand.Intn(10)+1, 6), vlanIds[1], datasourceTypeName, len(portIds))
	for _, args := range checkArgs {
		checks.append(t, args[0], args[1:]...)
	}
	assignments := make(map[apstra.ObjectId]map[apstra.ObjectId]bool, len(portIds))
	for _, portId := range portIds {
		assignments[portId] = map[apstra.ObjectId]bool{ctId: true}
	}
	require.NoError(t, bp.SetApplicationPointsConnectivityTemplates(ctx, assignments))

	config := insecureProviderConfigHCL + fmt.Sprintf(datasourceDatacneterConnectivityTemplatesStatus, datasourceType, datasourceName, bp.Id())
	t.Logf("\n// ------ begin config ------\n%s// -------- end config ------\n\n", config)
	t.Logf("\n// ------ begin checks ------\n%s// -------- end checks ------\n\n", checks.string())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeAggregateTestCheckFunc(checks.checks...),
			},
		},
	})
}
