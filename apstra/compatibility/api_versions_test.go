package compatibility

import (
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"testing"
)

func TestSupportedApiVersions(t *testing.T) {
	expected := []string{
		apiversions.Apstra420,
		apiversions.Apstra421,
		apiversions.Apstra4211,
		apiversions.Apstra422,
		apiversions.Apstra500,
	}

	result := SupportedApiVersions()

	if !utils.SlicesMatch(expected, result) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestSupportedApiVersionsPretty(t *testing.T) {
	expected :=
		apiversions.Apstra420 + ", " +
			apiversions.Apstra421 + ", " +
			apiversions.Apstra4211 + ", " +
			apiversions.Apstra422 + ", and " +
			apiversions.Apstra500

	result := SupportedApiVersionsPretty()
	if expected != result {
		t.Fatalf("expected %q; got %q", expected, result)
	}
}
