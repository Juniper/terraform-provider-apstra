package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"log"
	"terraform-provider-apstra/apstra"
)

const (
	DefaultVersion = "0.0.0"
	DefaultCommit  = "devel"
)

var commit, version string // populated by goreleaser

// NewApstraProvider instantiates the provider in main
func NewApstraProvider() provider.Provider {
	l := len(commit)
	switch {
	case l == 0:
		commit = DefaultCommit
	case l > 7:
		commit = commit[:8]
	}

	if len(version) == 0 {
		version = DefaultVersion
	}

	return &tfapstra.Provider{
		Version: version,
		Commit:  commit,
	}
}

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), NewApstraProvider, providerserver.ServeOpts{
		Address: "example.com/apstrktr/apstra",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
