package apstravalidator

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func OneOfStringers[S fmt.Stringer](s []S) validator.String {
	strings := make([]string, len(s))
	for i := range s {
		strings[i] = s[i].String()
	}
	return stringvalidator.OneOf(strings...)
}
