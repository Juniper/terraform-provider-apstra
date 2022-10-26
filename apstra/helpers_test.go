package apstra

import (
	"bitbucket.org/apstrktr/goapstra"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestSliceAttrValueToSliceString(t *testing.T) {
	test := []attr.Value{
		types.String{Value: "foo"},
		types.Int64{Value: 6},
	}
	expected := []string{
		"foo",
		"6",
	}
	result := sliceAttrValueToSliceString(test)
	if len(expected) != len(result) {
		t.Fatalf("expected %d results, got %d results", len(expected), len(result))
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != result[i] {
			t.Fatalf("expected '%s', got '%s'", expected[i], result[i])
		}
	}
}

func TestSliceAttrValueToSliceObjectId(t *testing.T) {
	test := []attr.Value{
		types.String{Value: "foo"},
		types.Int64{Value: 6},
	}
	expected := []goapstra.ObjectId{
		"foo",
		"6",
	}
	result := sliceAttrValueToSliceObjectId(test)
	if len(expected) != len(result) {
		t.Fatalf("expected %d results, got %d results", len(expected), len(result))
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != result[i] {
			t.Fatalf("expected '%s', got '%s'", expected[i], result[i])
		}
	}
}
