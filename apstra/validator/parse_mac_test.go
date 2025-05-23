package apstravalidator

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseMacValidator(t *testing.T) {
	ctx := context.Background()

	validMacs := []string{
		"00:01:02:03:04:05",
		"aa:aa:aa:aa:aa:aa",
		"AA:AA:AA:AA:AA:AA",
	}

	invalidMacs := []string{
		"",
		"00:01:02:03:04:05:06",
		"0200.5e10.0000.0001", // net.ParseMAC() is okay with this, but we require 48-bit addresses
	}

	type testCase struct {
		mac       string
		expectErr bool
	}

	var testCases []testCase
	for _, mac := range validMacs { // load valid test cases
		testCases = append(testCases, testCase{mac: mac, expectErr: false})
	}
	for _, rt := range invalidMacs { // load invalid test cases
		testCases = append(testCases, testCase{mac: rt, expectErr: true})
	}

	for _, tCase := range testCases {
		tCase := tCase
		t.Run(tCase.mac, func(t *testing.T) {
			t.Parallel()
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    types.StringValue(tCase.mac),
			}
			response := validator.StringResponse{}

			ParseMacValidator{}.ValidateString(ctx, request, &response)
			if response.Diagnostics.HasError() && !tCase.expectErr {
				t.Fail() // error where none expected
			}
			if !response.Diagnostics.HasError() && tCase.expectErr {
				t.Fail() // expected error not found
			}
		})
	}
}
