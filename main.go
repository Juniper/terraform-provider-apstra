package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"log"
	"terraform-provider-apstra/apstra"
)

var commit, version string // populated by goreleaser

// NewApstraProvider instantiates the provider in main
func NewApstraProvider() provider.Provider {
	return &apstra.Provider{
		Version: "v" + version,
		Commit:  commit,
	}
}

func main() {
	if commit == "" {
		commit = "devel"
	}
	if version == "" {
		version = "0.0.0"
	}
	log.Printf("version: %s commit: %s", version, commit)
	err := providerserver.Serve(context.Background(), NewApstraProvider, providerserver.ServeOpts{
		Address: "example.com/apstrktr/apstra",
	})
	if err != nil {
		log.Fatal(err)
	}
}
