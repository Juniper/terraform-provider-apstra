package apstra

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"path/filepath"
	"strings"
)

func sliceTfStringToSliceString(in []types.String) []string {
	//goland:noinspection GoPreferNilSlice
	out := []string{}
	for _, t := range in {
		out = append(out, t.Value)
	}
	return out
}

func sliceStringToSliceTfString(in []string) []types.String {
	var out []types.String
	for _, t := range in {
		out = append(out, types.String{Value: t})
	}
	return out
}

func keyLogWriterFromEnv(keyLogEnv string) (*os.File, error) {
	fileName, foundKeyLogFile := os.LookupEnv(keyLogEnv)
	if !foundKeyLogFile {
		return nil, nil
	}

	// expand ~ style home directory
	if strings.HasPrefix(fileName, "~/") {
		dirname, _ := os.UserHomeDir()
		fileName = filepath.Join(dirname, fileName[2:])
	}

	err := os.MkdirAll(filepath.Dir(fileName), os.FileMode(0600))
	if err != nil {
		return nil, err
	}
	return os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
}
