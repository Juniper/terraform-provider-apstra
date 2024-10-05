package testutils

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

const (
	testConfigFile = "../.testconfig.hcl"
	timeout        = 60 * time.Second // probably should be added to env vars and to test hcl file
)

var (
	sharedClient    *apstra.Client
	testClientMutex sync.Mutex
	testCfg         *testConfig
	testCfgMutex    *sync.Mutex = new(sync.Mutex)
)

type testConfig struct {
	Url                   string `hcl:"url,optional"`
	Username              string `hcl:"username,optional"`
	Password              string `hcl:"password,optional"`
	TlsValidationDisabled bool   `hcl:"tls_validation_disabled,optional"`
}

func GetTestClient(t testing.TB, ctx context.Context) *apstra.Client {
	t.Helper()

	testClientMutex.Lock()
	defer testClientMutex.Unlock()

	TestCfgFileToEnv(t)

	if sharedClient == nil {
		clientCfg, err := utils.NewClientConfig("", "")
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

func TestCfgFileToEnv(t testing.TB) {
	t.Helper()

	testCfgMutex.Lock()
	defer testCfgMutex.Unlock()

	if testCfg == nil {
		testCfg = new(testConfig)

		absPath, err := filepath.Abs(testConfigFile)
		if err != nil {
			t.Fatalf("while expanding config file path %s - %s", testConfigFile, err)
		}

		if _, err = os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
			return
		}

		err = hclsimple.DecodeFile(absPath, nil, testCfg)
		if err != nil {
			t.Fatalf("while parsing configuration from %q - %s", absPath, err)
		}
	}

	if testCfg.Url != "" {
		t.Setenv(constants.EnvUrl, testCfg.Url)
	}

	if testCfg.Username != "" {
		t.Setenv(constants.EnvUsername, testCfg.Username)
	}

	if testCfg.Password != "" {
		t.Setenv(constants.EnvPassword, testCfg.Password)
	}

	t.Setenv(constants.EnvTlsNoVerify, strconv.FormatBool(testCfg.TlsValidationDisabled))
}
