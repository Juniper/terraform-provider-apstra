package utils

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"reflect"
)

// JSONEqual takes 2 json strings in types.String variables and matches them
func JSONEqual(m1, m2 types.String, d *diag.Diagnostics) bool {
	var map1 interface{}
	var map2 interface{}

	var err error
	err = json.Unmarshal([]byte(m1.ValueString()), &map1)
	if err != nil {
		d.AddError("error unmarshalling string", err.Error())
		return false
	}
	err = json.Unmarshal([]byte(m2.ValueString()), &map2)
	if err != nil {
		d.AddError("error unmarshalling string", err.Error())
		return false
	}
	return reflect.DeepEqual(map1, map2)
}

// IsJSON takes a string and returns true if json, false if not

func IsJSON(str types.String) bool {
	var m interface{}
	err := json.Unmarshal([]byte(str.ValueString()), &m)
	if err != nil {
		return false
	}
	return true
}
