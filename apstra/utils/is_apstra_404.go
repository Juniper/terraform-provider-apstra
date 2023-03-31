package utils

import (
	"github.com/Juniper/apstra-go-sdk/apstra"
	"errors"
)

func IsApstra404(err error) bool {
	var ace apstra.ApstraClientErr
	if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
		return true
	}
	return false
}
