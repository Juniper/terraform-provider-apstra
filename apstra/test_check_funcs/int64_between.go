package testcheck

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"strconv"
	"strings"
)

// TestCheckResourceInt64AttrBetween is based on
// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1121C1-L1130C2
func TestCheckResourceInt64AttrBetween(name, key string, min, max int64) resource.TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		return testCheckResourceAttr(is, name, key, min, max)
	})
}

// testCheckResourceAttr is based on
// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1154C1-L1198C2
func testCheckResourceAttr(is *terraform.InstanceState, name, key string, min, max int64) error {
	v, ok := is.Attributes[key]

	if !ok {
		return fmt.Errorf("%s: Attribute '%s' not found", name, key)
	}

	vi64, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fmt.Errorf("%s: Attribute %q, value %q does not parse to int64 - %w", name, key, v, err)
	}

	if vi64 < min || vi64 > max {
		return fmt.Errorf("%s: Attribute %q must fall between %d and %d, got %d", name, key, min, max, vi64)
	}

	return nil
}

// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1645C1-L1657C2
func modulePrimaryInstanceState(ms *terraform.ModuleState, name string) (*terraform.InstanceState, error) {
	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("not found: %s in %s", name, ms.Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("no primary instance: %s in %s", name, ms.Path)
	}

	return is, nil
}

// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1672C1-L1675C2
func primaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	ms := s.RootModule() //nolint:staticcheck // legacy usage
	return modulePrimaryInstanceState(ms, name)
}

// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1680C1-L1687C2
func indexesIntoTypeSet(key string) bool {
	for _, part := range strings.Split(key, ".") {
		if i, err := strconv.Atoi(part); err == nil && i > 100 {
			return true
		}
	}
	return false
}

// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1689C1-L1697C2
func checkIfIndexesIntoTypeSet(key string, f resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := f(s)
		if err != nil && indexesIntoTypeSet(key) {
			return fmt.Errorf("error in test check: %s\nTest check address %q likely indexes into TypeSet\nThis is currently not possible in the SDK", err, key)
		}
		return err
	}
}
