package apstravalidator

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseRtValidator(t *testing.T) {
	ctx := context.Background()

	validRTs := []string{
		"0:0",
		"1:1",
		"65535:65535",
		"65536:65535",
		"65535:65536",
		"0.0.0.0:0",
		"255.255.255.255:0",
		"0.0.0.0:65535",
		"255.255.255.255:65535",
		"4294967295:65535",
		"65535:4294967295",
	}

	invalidRTs := []string{
		"",
		":",
		"1:",
		":1",
		"1",
		"4294967296:65535",
		"4294967295:65536",
		"65536:4294967295",
		"65535:4294967296",
		"-1:1",
		"1:-1",
		"256.1.2.3:1",
		"bogus",
	}

	type testCase struct {
		rt        string
		expectErr bool
	}

	var testCases []testCase
	for _, rt := range validRTs { // load valid test cases
		testCases = append(testCases, testCase{rt: rt, expectErr: false})
	}
	for _, rt := range invalidRTs { // load invalid test cases
		testCases = append(testCases, testCase{rt: rt, expectErr: true})
	}

	for _, tCase := range testCases {
		tCase := tCase
		t.Run(tCase.rt, func(t *testing.T) {
			t.Parallel()
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    types.StringValue(tCase.rt),
			}
			response := validator.StringResponse{}

			ParseRtValidator{}.ValidateString(ctx, request, &response)
			if response.Diagnostics.HasError() && !tCase.expectErr {
				t.Fail() // error where none expected
			}
			if !response.Diagnostics.HasError() && tCase.expectErr {
				t.Fail() // expected error not found
			}
		})
	}
}
