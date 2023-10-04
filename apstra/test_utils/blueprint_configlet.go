package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"testing"
)

func CatalogConfigletA(ctx context.Context, client *apstra.Client) (apstra.ObjectId, *apstra.ConfigletData,
	func(context.Context, apstra.ObjectId) error, error) {
	deleteFunc := func(ctx context.Context, in apstra.ObjectId) error { return client.DeleteConfiglet(ctx, in) }
	name := "CatalogConfigletA"

	c, err := client.GetConfigletByName(ctx, name)
	if utils.IsApstra404(err) {
		var cg []apstra.ConfigletGenerator
		cg = append(cg, apstra.ConfigletGenerator{
			ConfigStyle:  apstra.PlatformOSJunos,
			Section:      apstra.ConfigletSectionSystem,
			TemplateText: "interfaces {\n   {% if 'leaf1' in hostname %}\n    xe-0/0/3 {\n      disable;\n    }\n   {% endif %}\n   {% if 'leaf2' in hostname %}\n    xe-0/0/2 {\n      disable;\n    }\n   {% endif %}\n}",
		})

		var refarchs []apstra.RefDesign
		refarchs = append(refarchs, apstra.RefDesignTwoStageL3Clos)
		data := apstra.ConfigletData{
			DisplayName: name,
			RefArchs:    refarchs,
			Generators:  cg,
		}

		id, err := client.CreateConfiglet(context.Background(), &data)
		if err != nil {
			return "", nil, nil, nil
		}

		return id, &data, deleteFunc, nil
	}
	return c.Id, c.Data, deleteFunc, nil
}

func BlueprintConfigletA(ctx context.Context, client *apstra.TwoStageL3ClosClient, cid apstra.ObjectId, condition string) (apstra.ObjectId,
	func(context.Context, apstra.ObjectId) error, error) {
	deleteFunc := func(ctx context.Context, in apstra.ObjectId) error { return client.DeleteConfiglet(ctx, in) }

	id, err := client.ImportConfigletById(ctx, cid, condition, "")
	if err != nil {
		return "", nil, err
	}
	return id, deleteFunc, nil
}

// testWidgetsAB instantiates two predefined probes and creates widgets from them,
// returning the widget Object Id and the IbaWidgetData object used for creation

func TestWidgetsAB(ctx context.Context, t *testing.T, bpClient *apstra.TwoStageL3ClosClient) (apstra.ObjectId,
	apstra.IbaWidgetData, apstra.ObjectId, apstra.IbaWidgetData, func() error) {
	probeAId, err := bpClient.InstantiateIbaPredefinedProbe(ctx, &apstra.IbaPredefinedProbeRequest{
		Name: "bgp_session",
		Data: []byte(`{
			"Label":     "BGP Session Flapping",
			"Duration":  300,
			"Threshold": 40
		}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	probeBId, err := bpClient.InstantiateIbaPredefinedProbe(ctx, &apstra.IbaPredefinedProbeRequest{
		Name: "drain_node_traffic_anomaly",
		Data: []byte(`{
			"Label":     "Drain Traffic Anomaly",
			"Threshold": 100000
		}`),
	})

	if err != nil {
		t.Fatal(err)
	}

	widgetA := apstra.IbaWidgetData{
		Type:      apstra.IbaWidgetTypeStage,
		Label:     "BGP Session Flapping",
		ProbeId:   probeAId,
		StageName: "BGP Session",
	}
	widgetAId, err := bpClient.CreateIbaWidget(ctx, &widgetA)
	if err != nil {
		t.Fatal(err)
	}

	widgetB := apstra.IbaWidgetData{
		Type:      apstra.IbaWidgetTypeStage,
		Label:     "Drain Traffic Anomaly",
		ProbeId:   probeBId,
		StageName: "excess_range",
	}
	widgetBId, err := bpClient.CreateIbaWidget(ctx, &widgetB)
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() error {
		err = bpClient.DeleteIbaWidget(ctx, widgetAId)
		if err != nil {
			return err
		}
		err = bpClient.DeleteIbaWidget(ctx, widgetBId)
		if err != nil {
			return err
		}
		err = bpClient.DeleteIbaProbe(ctx, probeAId)
		if err != nil {
			return err
		}
		err = bpClient.DeleteIbaProbe(ctx, probeBId)
		return err
	}

	return widgetAId, widgetA, widgetBId, widgetB, cleanup
}
