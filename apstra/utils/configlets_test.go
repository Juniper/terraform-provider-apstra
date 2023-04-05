package utils

import (
	"testing"
)

func TestValidSectionsAsTable(t *testing.T) {
	expected := `  | **Config Style**  | **Valid Sections** |
  |---|---|
  |cumulus|file, frr, interface, ospf, system|
  |eos|interface, ospf, system, system_top|
  |junos|interface_level_hierarchical, interface_level_delete, interface_level_set, top_level_hierarchical, top_level_set_delete|
  |nxos|system, interface, system_top, ospf|
  |sonic|file, frr, ospf, system|
`
	table := ValidSectionsAsTable()
	if table != expected {
		t.Fatalf("expected this:\n%s\ngot this:\n%s\n", expected, table)
	}
}
