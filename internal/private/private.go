package private

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// SetterLoader is the interface implemented by types which can store and retrieve their
// own data in a State (a private state).
type SetterLoader interface {
	SetPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics)
	LoadPrivateState(ctx context.Context, ps State, diags *diag.Diagnostics)
}

// RoundTrip extracts the private state from ps into an instance of a SetterLoader,
// then stores it back into ps. It is a convenience function which facilitates
// updates to the key under which a SetterLoader stores its data (schema changes).
// By round-tripping the data, the private state will be upgraded.
//
// For SetterLoaders which use a single, consistent key, this function is effectively
// a no-op.
//
// For SetterLoaders which are migrating to a new key string, this function facilitates
// migration by ensuring that data is persisted to private state under both keys.
//
// The caller specifies the value type T (not a pointer) whose pointer implements
// SetterLoader, e.g. RoundTrip[ResourceDatacenterTag](ctx, ps, diags). No example
// value is required.
func RoundTrip[T any, PT interface {
	*T
	SetterLoader
}](ctx context.Context, ps State, diags *diag.Diagnostics) {
	var v T
	sl := PT(&v)

	// Load the struct from private state. If it was not stored, sl will be left unchanged (zero value).
	var d diag.Diagnostics
	sl.LoadPrivateState(ctx, ps, &d)
	diags.Append(d...)
	if d.HasError() {
		return
	}

	sl.SetPrivateState(ctx, ps, diags)
}

// loadPrivateState unmarshals the private state stored in ps into target. Private state data
// is keyed by the name of type T, the type target points to. If ps holds no data for that key,
// target is left unchanged and no error is reported. T must not be a pointer type (target must
// not be a pointer to a pointer): privateStateKey panics in that case. A nil target, or data
// which cannot be unmarshaled, is reported via diags.
func loadPrivateState[T any](ctx context.Context, ps State, target *T, diags *diag.Diagnostics, legacyKey ...string) {
	if target == nil {
		diags.AddError("failed to load private state", "target is nil")
		return
	}

	// We prefer to read legacy key, if any, for backward compatibility.
	switch len(legacyKey) {
	case 0: // No legacy key specified. Move on to using the key derived from the type name.
	case 1: // Try loading the data using the legacy key.
		b, d := ps.GetKey(ctx, legacyKey[0])
		diags.Append(d...)
		if d.HasError() {
			return
		}

		// neither zero-length nor nil data can be persisted to private state, so
		// there are only two conditions: Key has data, or key has no data. nil
		// is not a condition we need to handle.
		if len(b) > 0 {
			if err := json.Unmarshal(b, target); err != nil {
				diags.AddError("failed to unmarshal private state", err.Error())
			}
			return
		}
	default: // Programmer error.
		panic(fmt.Sprintf("internal error: loadPrivateState called with more than one legacy key: %v", legacyKey))
	}

	// Read using the key derived from the type name. This is the preferred key for new data.
	b, d := ps.GetKey(ctx, privateStateKey[T]())
	diags.Append(d...)
	if d.HasError() {
		return
	}

	if len(b) == 0 {
		return
	}

	if err := json.Unmarshal(b, target); err != nil {
		diags.AddError("failed to unmarshal private state", err.Error())
		return
	}
}

// setPrivateState marshals source to JSON and stores it in ps. Private state data is keyed by
// the name of type T, which loadPrivateState uses to retrieve it. A caller can specify a second,
// legacy key by passing a single value via the `legacyKey` variadic argument. source must be a
// value, not a pointer: privateStateKey panics if T is a pointer type. Marshaling or storage
// errors are reported via diags.
//
// A zero-value source is never persisted to private state. Instead, any private state data
// currently stored under the key (and legacy key, if any) is deleted. This approach minimizes
// the footprint of private state data, and avoids creating data which was never actually set by
// the caller.
func setPrivateState[T any](ctx context.Context, ps State, source T, diags *diag.Diagnostics, legacyKey ...string) {
	if len(legacyKey) > 1 {
		panic(fmt.Sprintf("internal error: setPrivateState called with more than one legacy key: %v", legacyKey))
	}

	if reflect.ValueOf(&source).Elem().IsZero() {
		// For efficiency, we don't store zero-value data. This is to avoid creating data which was never actually set by the caller.
		diags.Append(ps.SetKey(ctx, privateStateKey[T](), nil)...)
		if len(legacyKey) == 1 {
			diags.Append(ps.SetKey(ctx, legacyKey[0], nil)...)
		}
		return
	}

	b, err := json.Marshal(source)
	if err != nil {
		diags.AddError("failed to marshal private state", err.Error())
		return
	}

	// Save the data under the key derived from the type name. This is the preferred key for new data.
	diags.Append(ps.SetKey(ctx, privateStateKey[T](), b)...)

	// Save the data under the legacy key, if any. This is for backward compatibility.
	if len(legacyKey) == 1 {
		diags.Append(ps.SetKey(ctx, legacyKey[0], b)...)
	}
}

// privateStateKey returns the standard key under which a value of type T should be stored in
// private state. Both loadPrivateState and setPrivateState must use it so that the keys always
// match. The key is the type name (a string) as determined by the reflect package, for example
// "private.ResourceDatacenterTag". Renaming a type or this package orphans previously stored
// state. Fortunately, it will also cause tests to fail. Just don't rename things since we're
// encoding the package and type names into the state file.
//
// T must not be a pointer type. Passing a pointer is a programming error, and will panic.
func privateStateKey[T any]() string {
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Pointer {
		panic(fmt.Sprintf("internal error: privateStateKey called with pointer type %s; use a value type instead", t.String()))
	}
	return t.String()
}
