package utils

import (
	"errors"
	"github.com/Juniper/apstra-go-sdk/apstra"
)

func IsApstra404(err error) bool {
	var ace apstra.ClientErr
	if errors.As(err, &ace) && ace.Type() == apstra.ErrNotfound {
		return true
	}
	return false
}
