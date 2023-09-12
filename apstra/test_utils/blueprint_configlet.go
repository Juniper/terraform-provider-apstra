package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
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
