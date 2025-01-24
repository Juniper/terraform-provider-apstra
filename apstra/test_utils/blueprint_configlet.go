package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra/enum"
	"testing"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/stretchr/testify/require"
)

func CatalogConfigletA(t testing.TB, ctx context.Context, client *apstra.Client) (apstra.ObjectId, *apstra.ConfigletData) {
	t.Helper()

	name := "CatalogConfigletA"
	configlet, err := client.GetConfigletByName(ctx, name)
	if err != nil {
		if !utils.IsApstra404(err) {
			require.NoError(t, err) // we cannot handle non-404 errors
		}

		configletData := apstra.ConfigletData{
			DisplayName: name,
			RefArchs:    []enum.RefDesign{enum.RefDesignDatacenter},
			Generators: []apstra.ConfigletGenerator{{
				ConfigStyle:  enum.ConfigletStyleJunos,
				Section:      enum.ConfigletSectionSystem,
				TemplateText: "interfaces {\n   {% if 'leaf1' in hostname %}\n    xe-0/0/3 {\n      disable;\n    }\n   {% endif %}\n   {% if 'leaf2' in hostname %}\n    xe-0/0/2 {\n      disable;\n    }\n   {% endif %}\n}",
			}},
		}

		id, err := client.CreateConfiglet(context.Background(), &configletData)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, client.DeleteConfiglet(ctx, id)) })

		return id, &configletData
	}

	return configlet.Id, configlet.Data
}

func BlueprintConfigletA(t testing.TB, ctx context.Context, client *apstra.TwoStageL3ClosClient, cid apstra.ObjectId, condition string) apstra.ObjectId {
	t.Helper()

	id, err := client.ImportConfigletById(ctx, cid, condition, "")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.DeleteConfiglet(ctx, id)) })

	return id
}

// TestWidgetsAB instantiates two predefined probes and creates widgets from them,
// returning the widget Object Id and the IbaWidgetData object used for creation
//func TestWidgetsAB(t testing.TB, ctx context.Context, bpClient *apstra.TwoStageL3ClosClient) (apstra.ObjectId, apstra.IbaWidgetData, apstra.ObjectId, apstra.IbaWidgetData) {
//	probeAId, err := bpClient.InstantiateIbaPredefinedProbe(ctx, &apstra.IbaPredefinedProbeRequest{
//		Name: "bgp_session",
//		Data: []byte(`{
//			"Label":     "BGP Session Flapping",
//			"Duration":  300,
//			"Threshold": 40
//		}`),
//	})
//	require.NoError(t, err)
//	t.Cleanup(func() { require.NoError(t, bpClient.DeleteIbaProbe(ctx, probeAId)) })
//
//	probeBId, err := bpClient.InstantiateIbaPredefinedProbe(ctx, &apstra.IbaPredefinedProbeRequest{
//		Name: "drain_node_traffic_anomaly",
//		Data: []byte(`{
//			"Label":     "Drain Traffic Anomaly",
//			"Threshold": 100000
//		}`),
//	})
//	require.NoError(t, err)
//	t.Cleanup(func() { require.NoError(t, bpClient.DeleteIbaProbe(ctx, probeBId)) })
//
//	widgetA := apstra.IbaWidgetData{
//		Type:      enum.IbaWidgetTypeStage,
//		Label:     "BGP Session Flapping",
//		ProbeId:   probeAId,
//		StageName: "BGP Session",
//	}
//	widgetAId, err := bpClient.CreateIbaWidget(ctx, &widgetA)
//	require.NoError(t, err)
//	t.Cleanup(func() { require.NoError(t, bpClient.DeleteIbaWidget(ctx, widgetAId)) })
//
//	widgetB := apstra.IbaWidgetData{
//		Type:      enum.IbaWidgetTypeStage,
//		Label:     "Drain Traffic Anomaly",
//		ProbeId:   probeBId,
//		StageName: "excess_range",
//	}
//	widgetBId, err := bpClient.CreateIbaWidget(ctx, &widgetB)
//	require.NoError(t, err)
//	t.Cleanup(func() { require.NoError(t, bpClient.DeleteIbaWidget(ctx, widgetBId)) })
//
//	return widgetAId, widgetA, widgetBId, widgetB
//}
