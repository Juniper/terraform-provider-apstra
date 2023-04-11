package utils

import (
	"fmt"
	"strings"
)

const (
	mustEscapeUrlChars = "\"#%/<>?[\\]^"
)

func UrlEscapeTable() string {
	sb := strings.Builder{}
	for i, b := range mustEscapeUrlChars {
		if i%4 == 0 {
			if i == 0 {
				sb.WriteString("\t")
			} else {
				sb.WriteString("\n\t")
			}
		} else {
			sb.WriteString("\t")
		}
		sb.WriteString(fmt.Sprintf("%s => %%%X", string(b), b))
	}

	return sb.String()
}
