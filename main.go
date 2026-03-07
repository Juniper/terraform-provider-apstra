package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	tfapstra "github.com/Juniper/terraform-provider-apstra/apstra"
	imperativeserver "github.com/chrismarget/imperative-terraform/server"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	debug := flag.Bool("debug", false, "set to true to run the provider with support for debuggers like delve")
	version := flag.Bool("version", false, "print version and exit")
	imperativeServerMode := flag.Bool("imperative-server-mode", false, "run in imperative server mode")
	flag.Parse()

	var err error
	switch {
	case *imperativeServerMode:
		server := imperativeserver.New(
			imperativeserver.WithProvider(tfapstra.NewProvider()),
			imperativeserver.WithDataSources([]string{"apstra_tag"}),
			imperativeserver.WithResources([]string{}),
			imperativeserver.WithTimeouts(5*time.Minute, 5*time.Second),
		)
		err = server.Serve(context.Background())
	case *version:
		p := tfapstra.NewProvider()
		mdr := provider.MetadataResponse{}
		p.Metadata(context.Background(), provider.MetadataRequest{}, &mdr)
		_, err = os.Stdout.WriteString(fmt.Sprintf("terraform-provider-%s %s\n", mdr.TypeName, mdr.Version))
	default:
		err = providerserver.Serve(context.Background(), tfapstra.NewProvider, providerserver.ServeOpts{
			Address: "registry.terraform.io/Juniper/apstra",
			Debug:   *debug,
		})
	}
	if err != nil {
		log.Fatal(err)
	}
}
