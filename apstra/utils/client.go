package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Juniper/apstra-go-sdk/apstra"
	"github.com/Juniper/terraform-provider-apstra/apstra/constants"
)

const (
	envTlsKeyLogFile = "SSLKEYLOGFILE"
	urlEncodeMsg     = `
Note that when the Username or Password fields contain special characters and are
embedded in the URL, they must be URL-encoded by substituting '%%<hex-value>' in
place of each special character. The following table demonstrates some common
substitutions:

%s`
)

func NewClientConfig(apstraUrl, envVarPrefix string) (*apstra.ClientCfg, error) {
	// Populate raw URL string from config or environment.
	if apstraUrl == "" {
		apstraUrl = os.Getenv(envVarPrefix + constants.EnvUrl)
	}

	if apstraUrl == "" {
		return nil, errors.New("missing Apstra URL")
	}

	// Parse the URL.
	parsedUrl, err := url.Parse(apstraUrl)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok &&
			(strings.Contains(urlErr.Error(), "invalid userinfo") ||
				strings.Contains(urlErr.Error(), "invalid port")) {
			// don't print the actual error here because it likely contains a password
			return nil, errors.New(
				"error parsing URL\n" + fmt.Sprintf(urlEncodeMsg, UrlEscapeTable()))
		}

		var urlEE url.EscapeError
		if errors.As(err, &urlEE) {
			return nil, errors.New("error parsing URL - " + fmt.Sprintf(urlEncodeMsg, UrlEscapeTable()))
		}

		return nil, fmt.Errorf("error parsing URL %q - %w", apstraUrl, err)
	}

	// Determine the Apstra username.
	user := parsedUrl.User.Username()
	if user == "" {
		if val, ok := os.LookupEnv(envVarPrefix + constants.EnvUsername); ok {
			user = val
		} else {
			return nil, errors.New("unable to determine apstra username - " + fmt.Sprintf(urlEncodeMsg, UrlEscapeTable()))
		}
	}

	// Determine  the Apstra password.
	pass, found := parsedUrl.User.Password()
	if !found {
		if val, ok := os.LookupEnv(envVarPrefix + constants.EnvPassword); ok {
			pass = val
		} else {
			return nil, errors.New("unable to determine apstra password")
		}
	}

	// Remove credentials from the URL prior to rendering it into ClientCfg.
	parsedUrl.User = nil

	// Set up a logger.
	var logger *log.Logger
	if logFileName, ok := os.LookupEnv(envVarPrefix + constants.EnvLogfile); ok {
		logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		logger = log.New(logFile, "", 0)
	}

	// Set up the TLS session key log.
	var klw io.Writer
	if fileName, ok := os.LookupEnv(envTlsKeyLogFile); ok {
		klw, err = newKeyLogWriter(fileName)
		if err != nil {
			return nil, err
		}
	}

	// Create the client's httpClient
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				KeyLogWriter: klw,
			},
		},
	}

	_, experimental := os.LookupEnv(envVarPrefix + constants.EnvExperimental)

	// Create the clientCfg
	return &apstra.ClientCfg{
		Url:          parsedUrl.String(),
		User:         user,
		Pass:         pass,
		Logger:       logger,
		HttpClient:   httpClient,
		Experimental: experimental,
	}, nil
}
