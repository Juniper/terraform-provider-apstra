package compatibility

import (
	apiversions "github.com/Juniper/terraform-provider-apstra/apstra/api_versions"
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"testing"
)

func TestSupportedApiVersions(t *testing.T) {
	expected := []string{
		apiversions.Apstra410,
		apiversions.Apstra411,
		apiversions.Apstra412,
		apiversions.Apstra420,
	}

	result := SupportedApiVersions()

	if !utils.SlicesMatch(expected, result) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestSupportedApiVersionsPretty(t *testing.T) {
	expected := apiversions.Apstra410 + ", " + apiversions.Apstra411 + ", " + apiversions.Apstra412 + ", and " + apiversions.Apstra420
	result := SupportedApiVersionsPretty()
	if expected != result {
		t.Fatalf("expected %q; got %q", expected, result)
	}
}
