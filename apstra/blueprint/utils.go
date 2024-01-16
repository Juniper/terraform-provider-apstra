package blueprint

import (
	"context"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func NodeTags(ctx context.Context, id string, client apstra.TwoStageL3ClosClient, diags *diag.Diagnostics) []string {
	query := new(apstra.PathQuery).
		SetBlueprintId(client.Id()).
		SetBlueprintType(apstra.BlueprintTypeStaging).
		SetClient(client.Client()).
		Node([]apstra.QEEAttribute{{Key: "id", Value: apstra.QEStringVal(id)}}).
		In([]apstra.QEEAttribute{apstra.RelationshipTypeTag.QEEAttribute()}).
		Node([]apstra.QEEAttribute{
			apstra.NodeTypeTag.QEEAttribute(),
			{Key: "name", Value: apstra.QEStringVal("n_tag")},
		})

	var queryResponse struct {
		Items []struct {
			Tag struct {
				Label string `json:"label"`
			} `json:"tag"`
		} `json:"items"`
	}

	err := query.Do(ctx, &queryResponse)
	if err != nil {
		diags.AddError(fmt.Sprintf("failed querying for node %q tags", id), err.Error())
		return nil
	}

	result := make([]string, len(queryResponse.Items))
	for i, item := range queryResponse.Items {
		result[i] = item.Tag.Label
	}

	return result
}

func friendlyPolicyRuleProtocols() []string {
	enums := apstra.PolicyRuleProtocols.Members()
	friendlyStrings := make([]string, len(enums))

	for i, enum := range enums {
		friendlyStrings[i] = utils.StringersToFriendlyString(enum)
	}

	return friendlyStrings
}
