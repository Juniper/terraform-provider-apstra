package importvalidation_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	importvalidation "github.com/Juniper/terraform-provider-apstra/internal/import_validation"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// testIdentity is the single struct-based identity type used by the test cases.
// The schema it reports is set per-instance via attrs, so each case can mark
// different fields as required (or reference fields which don't exist).
type testIdentity struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Opt    types.String `tfsdk:"opt,omitempty"`
	Plain  string       `tfsdk:"plain"`  // does not implement attr.Value
	hidden types.String `tfsdk:"hidden"` // unexported: cannot be inspected

	attrs map[string]identityschema.Attribute
}

func (t testIdentity) Attributes() map[string]identityschema.Attribute {
	return t.attrs
}

// notStruct is a non-struct identity.
type notStruct map[string]string

func (notStruct) Attributes() map[string]identityschema.Attribute {
	return map[string]identityschema.Attribute{"id": identityschema.StringAttribute{RequiredForImport: true}}
}

func TestCheckRequiredFields(t *testing.T) {
	t.Parallel()

	required := identityschema.StringAttribute{RequiredForImport: true}
	optional := identityschema.StringAttribute{OptionalForImport: true}

	// wantErr describes an expected error diagnostic: the summary must match
	// exactly, and the detail must begin with detailPrefix.
	type wantErr struct {
		summary      string
		detailPrefix string
	}

	var (
		errValue   = wantErr{importvalidation.FailSummaryMissingValue, importvalidation.FailDetailNullOrUnknown}
		errField   = wantErr{importvalidation.FailSummaryMissingField, importvalidation.FailDetailNoTag}
		errInvalid = wantErr{importvalidation.FailSummaryInvalidField, importvalidation.FailDetailCannotInspect}
		errType    = wantErr{importvalidation.FailSummaryInvalidType, importvalidation.FailDetailNotAttrValue}
		errNil     = wantErr{importvalidation.FailSummaryGeneral, importvalidation.FailDetailNilIdentity}
		errStruct  = wantErr{importvalidation.FailSummaryGeneral, importvalidation.FailDetailNotStruct}
	)

	tests := map[string]struct {
		identity       importvalidation.IdentityWithAttributes
		expectedErrors []wantErr
	}{
		"all_required_present": {
			identity: testIdentity{
				ID:   types.StringValue("a"),
				Name: types.StringValue("b"),
				Opt:  types.StringNull(),
				attrs: map[string]identityschema.Attribute{
					"id":   required,
					"name": required,
					"opt":  optional,
				},
			},
		},
		"pointer_all_present": {
			identity: &testIdentity{
				ID:   types.StringValue("a"),
				Name: types.StringValue("b"),
				attrs: map[string]identityschema.Attribute{
					"id":   required,
					"name": required,
					"opt":  optional,
				},
			},
		},
		"optional_unknown_ignored": {
			identity: testIdentity{
				ID:   types.StringValue("a"),
				Name: types.StringValue("b"),
				Opt:  types.StringUnknown(),
				attrs: map[string]identityschema.Attribute{
					"id":   required,
					"name": required,
					"opt":  optional,
				},
			},
		},
		"empty_string_is_present": {
			identity: testIdentity{
				ID:   types.StringValue(""),
				Name: types.StringValue("b"),
				attrs: map[string]identityschema.Attribute{
					"id":   required,
					"name": required,
					"opt":  optional,
				},
			},
		},
		"one_null": {
			expectedErrors: []wantErr{errValue},
			identity: testIdentity{
				ID:   types.StringValue("a"),
				Name: types.StringNull(),
				attrs: map[string]identityschema.Attribute{
					"id":   required,
					"name": required,
					"opt":  optional,
				},
			},
		},
		"one_unknown": {
			expectedErrors: []wantErr{errValue},
			identity: testIdentity{
				ID:   types.StringUnknown(),
				Name: types.StringValue("b"),
				attrs: map[string]identityschema.Attribute{
					"id":   required,
					"name": required,
					"opt":  optional,
				},
			},
		},
		"all_null": {
			expectedErrors: []wantErr{errValue, errValue},
			identity: testIdentity{
				attrs: map[string]identityschema.Attribute{
					"id":   required,
					"name": required,
					"opt":  optional,
				},
			},
		},
		"nil_pointer": {
			expectedErrors: []wantErr{errNil},
			identity:       (*testIdentity)(nil),
		},
		"not_a_struct": {
			expectedErrors: []wantErr{errStruct},
			identity:       notStruct{},
		},
		"schema_field_missing_from_struct": {
			expectedErrors: []wantErr{errField},
			identity: testIdentity{
				ID: types.StringValue("a"),
				attrs: map[string]identityschema.Attribute{
					"id":      required,
					"missing": required,
				},
			},
		},
		"field_not_attr_value": {
			expectedErrors: []wantErr{errType},
			identity: testIdentity{
				Plain: "a",
				attrs: map[string]identityschema.Attribute{
					"plain": required,
				},
			},
		},
		"unexported_field": {
			expectedErrors: []wantErr{errInvalid},
			identity: testIdentity{
				hidden: types.StringValue("a"),
				attrs: map[string]identityschema.Attribute{
					"hidden": required,
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			importvalidation.CheckRequiredFields(context.Background(), tt.identity, &diags)

			actual := diags.Errors()
			require.Len(t, actual, len(tt.expectedErrors))

			// match each actual error to a distinct expected error
			remaining := slices.Clone(tt.expectedErrors)
			for _, d := range actual {
				i := slices.IndexFunc(remaining, func(w wantErr) bool {
					return w.summary == d.Summary() && strings.HasPrefix(d.Detail(), w.detailPrefix)
				})
				require.GreaterOrEqual(t, i, 0, "unexpected error: %q / %q", d.Summary(), d.Detail())
				remaining = slices.Delete(remaining, i, i+1)
			}
		})
	}
}
