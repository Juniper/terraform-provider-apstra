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
	log.Println(argns)
}
