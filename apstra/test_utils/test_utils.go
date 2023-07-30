package testutils

import (
	"context"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"net/http"
	"sync"
	"terraform-provider-apstra/apstra/utils"
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
		clientCfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true

		sharedClient, err = clientCfg.NewClient(ctx)
		if err != nil {
			return nil, err
		}
	}

	return sharedClient, nil
}
