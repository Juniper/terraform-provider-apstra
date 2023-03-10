package utils

import (
	"bitbucket.org/apstrktr/goapstra"
	"errors"
)

func IsApstra404(err error) bool {
	var ace goapstra.ApstraClientErr
	if errors.As(err, &ace) && ace.Type() == goapstra.ErrNotfound {
		return true
	}
	return false
}
