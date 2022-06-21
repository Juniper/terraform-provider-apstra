package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"log"
	"terraform-provider-apstra/apstra"
)

func main() {
	err := providerserver.Serve(context.Background(), apstra.New, providerserver.ServeOpts{
		Address: "example.com/chrismarget-j/apstra",
	})
	if err != nil {
		log.Fatal(err)
	}
}
