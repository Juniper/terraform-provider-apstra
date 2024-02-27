package testutils

import (
	"context"
	"errors"
	"fmt"
	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

const (
	testConfigFile = "../.testconfig.hcl"
	timeout        = 60 * time.Second // probably should be added to env vars and to test hcl file
)

var sharedClient *apstra.Client
var testClientMutex sync.Mutex

type testConfig struct {
	Url      string `hcl:"url,optional"`
	Username string `hcl:"username,optional"`
	Password string `hcl:"password,optional"`
}

func GetTestClient(t testing.TB, ctx context.Context) *apstra.Client {
	t.Helper()

	testClientMutex.Lock()
	defer testClientMutex.Unlock()

	if sharedClient == nil {
		err := testCfgFileToEnv()
		if err != nil {
			t.Fatal(err)
		}

		clientCfg, err := utils.NewClientConfig("")
		if err != nil {
			t.Fatal(err)
		}

		clientCfg.Timeout = timeout
		clientCfg.Experimental = true
		clientCfg.HttpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true

		sharedClient, err = clientCfg.NewClient(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}

	return sharedClient
}

func testCfgFileToEnv() error {
	absPath, err := filepath.Abs(testConfigFile)
	if err != nil {
		return fmt.Errorf("error expanding config file path %s - %w", testConfigFile, err)
	}

	if _, err = os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	testCfg := new(testConfig)
	err = hclsimple.DecodeFile(absPath, nil, testCfg)
	if err != nil {
		return fmt.Errorf("failed to parse configuration from %q - %w", absPath, err)
	}

	if testCfg.Url != "" {
		err = os.Setenv(utils.EnvApstraUrl, testCfg.Url)
		if err != nil {
			return fmt.Errorf("failed setting environment variable %q - %w", utils.EnvApstraUrl, err)
		}
	}

	if testCfg.Username != "" {
		err = os.Setenv(utils.EnvApstraUsername, testCfg.Username)
		if err != nil {
			return fmt.Errorf("failed setting environment variable %q - %w", utils.EnvApstraUsername, err)
		}
	}

	if testCfg.Password != "" {
		err = os.Setenv(utils.EnvApstraPassword, testCfg.Password)
		if err != nil {
			return fmt.Errorf("failed setting environment variable %q - %w", utils.EnvApstraPassword, err)
		}
	}

	return nil
}
