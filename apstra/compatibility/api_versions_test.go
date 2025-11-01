package compatibility

import (
	"slices"
	"testing"

	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
)

func TestSupportedApiVersions(t *testing.T) {
	expected := []string{
		apiversions.Apstra420,
		apiversions.Apstra421,
		apiversions.Apstra4211,
		apiversions.Apstra422,
		apiversions.Apstra500,
		apiversions.Apstra501,
		apiversions.Apstra510,
		apiversions.Apstra600,
	}

	result := SupportedApiVersions()

	if !slices.Equal(expected, result) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestSupportedApiVersionsPretty(t *testing.T) {
	expected := apiversions.Apstra420 + ", " +
		apiversions.Apstra421 + ", " +
		apiversions.Apstra4211 + ", " +
		apiversions.Apstra422 + ", " +
		apiversions.Apstra500 + ", " +
		apiversions.Apstra501 + ", " +
		apiversions.Apstra510 + ", and " +
		apiversions.Apstra600

	result := SupportedApiVersionsPretty()
	if expected != result {
		t.Fatalf("expected %q; got %q", expected, result)
	}
}
