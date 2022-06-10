package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"terraform-provider-apstra/apstra"
)

func main() {
	tfsdk.Serve(context.Background(), apstra.New, tfsdk.ServeOpts{
		Name: "apstra",
	})
}
