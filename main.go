package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	var debug bool
	var printVersion bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.Parse()

	if printVersion {
		p := tfapstra.NewProvider()
		mdr := provider.MetadataResponse{}
		p.Metadata(context.Background(), provider.MetadataRequest{}, &mdr)
		_, err := os.Stdout.WriteString(fmt.Sprintf("terraform-provider-%s %s\n", mdr.TypeName, mdr.Version))
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	err := providerserver.Serve(context.Background(), tfapstra.NewProvider, providerserver.ServeOpts{
		Address: "registry.terraform.io/Juniper/apstra",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}

	//test comment which gofumpt will complain about
}
