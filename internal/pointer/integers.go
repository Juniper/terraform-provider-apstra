package pointer

import "golang.org/x/exp/constraints"

// ConvertInteger converts a pointer to an integer of one type into a pointer to an integer of
// another type.
//
// If src == nil, dst is ignored and the function returns nil.
// If src != nil and dst == nil, the function allocates a new integer of type Dst and returns a pointer to it.
// If src != nil and dst != nil, the function sets the value of the integer pointed to by dst to the value of
// src (after type conversion) and returns a pointer to the integer pointed to by dst. Callers can use either
// the returned value (the same pointer they passed in) or data at dst which now contains the converted value.
//
// Callers who may pass in nil src must not rely on the value at *dst.
func ConvertInteger[Dst, Src constraints.Integer](dst *Dst, src *Src) *Dst {
	if src == nil {
		return nil
	}

	if dst == nil {
		return new(Dst(*src))
	}

	*dst = Dst(*src)
	return dst
}
