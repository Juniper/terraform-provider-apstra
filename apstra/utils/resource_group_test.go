package utils

import (
	"log"
	"testing"
)

func TestAllResourceGroupNameStrings(t *testing.T) {
	argns := AllResourceGroupNameStrings()
	for i := range argns {
		if argns[i] == "" {
			t.Fatal("AllResourceGroupNameStrings() returned an empty string")
		}
	}
	expectedRgnCount := 20
	if len(argns) != expectedRgnCount {
		t.Fatalf("expected %d resource group names, got %d", expectedRgnCount, len(argns))
	}
	log.Println(argns)
}
