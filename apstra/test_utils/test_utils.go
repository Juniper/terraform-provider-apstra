package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"net/http"
	"sync"
)

var sharedClient *apstra.Client
var testClientMutex sync.Mutex

func GetTestClient(ctx context.Context) (*apstra.Client, error) {
	testClientMutex.Lock()
	defer testClientMutex.Unlock()

	if sharedClient == nil {
		clientCfg, err := utils.NewClientConfig("")
		if err != nil {
			return nil, err
		}
		clientCfg.Experimental = true
		clientCfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true

		sharedClient, err = clientCfg.NewClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return sharedClient, nil
}
