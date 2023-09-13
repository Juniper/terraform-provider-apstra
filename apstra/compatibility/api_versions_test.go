package compatibility

import (
	"github.com/Juniper/terraform-provider-apstra/apstra/utils"
	"testing"
)

func TestSupportedApiVersions(t *testing.T) {
	expected := []string{
		"4.1.0",
		"4.1.1",
		"4.1.2",
	}

	result := SupportedApiVersions()

	if !utils.SlicesMatch(expected, result) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestSupportedApiVersionsPretty(t *testing.T) {
	expected := "4.1.0, 4.1.1, and 4.1.2"
	result := SupportedApiVersionsPretty()
	if expected != result {
		t.Fatalf("expected %q; got %q", expected, result)
	}
}
