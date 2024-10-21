package apstravalidator

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = mustHaveNOfValidator{}

type mustHaveNOfValidator struct {
	n          int
	attributes []string
	atLeast    bool
	atMost     bool
}

func (o mustHaveNOfValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("ensure that the object has at least %d of the following attributes configured: ['%s']",
		o.n, strings.Join(o.attributes, "', '"))
}

func (o mustHaveNOfValidator) MarkdownDescription(ctx context.Context) string {
	return o.Description(ctx)
}

func (o mustHaveNOfValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return // can't validate null or unknown objects
	}

	// extract the attributes from the configured object
	attributeMap := req.ConfigValue.Attributes()

	// when no attributes are enumerated during validator creation, that's a signal
	// to consider all attributes. Rewrite the attribute slice with all attributes.
	if len(o.attributes) == 0 {
		o.attributes = make([]string, len(attributeMap))
		i := 0
		for attributeName := range attributeMap {
			o.attributes[i] = attributeName
			i++
		}

	}

	// n can never be larger than the number of known attributes
	if o.n > len(o.attributes) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid validator for element value",
			"While performing schema-based validation, an unexpected error occurred. "+
				"A schema validator which validates specific attributes has been configured for this object, "+
				"but the validator has been configured to check more objects than it knows about.\n\n"+
				fmt.Sprintf("Count of attributes to check: %d\n Attributes known to the validator: ['%s']",
					o.n, strings.Join(o.attributes, "', '"),
				),
		)

		return
	}

	foundValueCount := 0 // we'll compare this to 'n' later

	// loop over the attributes the caller asked us to count
	for _, requiredAttribute := range o.attributes {
		var foundAttribute attr.Value
		var ok bool

		// make sure the specified attributes exist
		if foundAttribute, ok = attributeMap[requiredAttribute]; !ok {
			attributeSlice := make([]string, len(attributeMap))
			var i int
			for s := range attributeMap {
				attributeSlice[i] = s
			}
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid validator for element value",
				"While performing schema-based validation, an unexpected error occurred. "+
					"A schema validator which validates specific attributes has been configured for this object, "+
					"but the available attributes don't include all of the attributes requested for validation.\n\n"+
					fmt.Sprintf("Available attributes: ['%s']\n Attributes to be validated: ['%s']",
						strings.Join(attributeSlice, "', '"),
						strings.Join(o.attributes, "', '"),
					),
			)
			return
		}

		// Can't validate with unknown values
		if foundAttribute.IsUnknown() {
			return
		}

		// increment the counter if the attribute has a value
		if !foundAttribute.IsNull() {
			foundValueCount++
		}
	}

	if o.atLeast && foundValueCount < o.n {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Insufficient attribute configuration",
			fmt.Sprintf("At least %d values from: ['%s'] must be configured.",
				o.n,
				strings.Join(o.attributes, "', '")),
		)
		return
	}

	if o.atMost && foundValueCount > o.n {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Too many attributes configured",
			fmt.Sprintf("At most %d values from: ['%s'] must be configured.",
				o.n,
				strings.Join(o.attributes, "', '")),
		)
		return
	}

	if !o.atLeast && !o.atMost && foundValueCount != o.n {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Wrong number of attributes configured",
			fmt.Sprintf("exactly %d values from: ['%s'] must be configured.",
				o.n,
				strings.Join(o.attributes, "', '")),
		)
		return
	}
}

func AtLeastNAttributes(n int, attributes ...string) validator.Object {
	return &mustHaveNOfValidator{
		n:          n,
		attributes: attributes,
		atLeast:    true,
	}
}

func AtMostNAttributes(n int, attributes ...string) validator.Object {
	return &mustHaveNOfValidator{
		n:          n,
		attributes: attributes,
		atMost:     true,
	}
}

func ExactlyNAttributes(n int, attributes ...string) validator.Object {
	return &mustHaveNOfValidator{
		n:          n,
		attributes: attributes,
	}
}
