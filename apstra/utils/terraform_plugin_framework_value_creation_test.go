package utils

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"log"
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

func TestInt64ValueOrNull(t *testing.T) {
	ctx := context.Background()
	diags := diag.Diagnostics{}
	rand.Seed(time.Now().UnixNano())

	r := rand.Intn(127) + 1 // 1 - 128 to avoid zero and fit in a int8
	log.Println(r)

	testCases := []any{
		r, int8(r), int16(r), int32(r), int64(r), uint(r), uint8(r), uint16(r), uint32(r), uint64(r),
		(*int)(unsafe.Pointer(&r)),
		(*int8)(unsafe.Pointer(&r)),
		(*int16)(unsafe.Pointer(&r)),
		(*int32)(unsafe.Pointer(&r)),
		(*int64)(unsafe.Pointer(&r)),
		(*uint)(unsafe.Pointer(&r)),
		(*uint8)(unsafe.Pointer(&r)),
		(*uint16)(unsafe.Pointer(&r)),
		(*uint32)(unsafe.Pointer(&r)),
		(*uint64)(unsafe.Pointer(&r)),
	}
	for _, tc := range testCases {
		attrVal := Int64ValueOrNull(ctx, tc, &diags)
		if attrVal.ValueInt64() != int64(r) {
			t.Fatal()
		}
		if diags.HasError() {
			t.Fatal()
		}
	}

	var intPtrNil *int
	var int8PtrNil *int8
	var int16PtrNil *int16
	var int32PtrNil *int32
	var int64PtrNil *int64
	var uintPtrNil *uint
	var uint8PtrNil *uint8
	var uint16PtrNil *uint16
	var uint32PtrNil *uint32
	var uint64PtrNil *uint64
	nilCases := []any{
		intPtrNil,
		int8PtrNil,
		int16PtrNil,
		int32PtrNil,
		int64PtrNil,
		uintPtrNil,
		uint8PtrNil,
		uint16PtrNil,
		uint32PtrNil,
		uint64PtrNil,
	}
	for _, tc := range nilCases {
		attrVal := Int64ValueOrNull(ctx, tc, &diags)
		if !attrVal.IsNull() {
			t.Fatal()
		}
		if diags.HasError() {
			t.Fatal()
		}
	}

	attrVal := Int64ValueOrNull(ctx, nil, &diags)
	if !attrVal.IsNull() {
		t.Fatal()
	}

	failCases := []any{"foo"}

	for _, tc := range failCases {
		failDiags := diag.Diagnostics{}
		attrVal := Int64ValueOrNull(ctx, tc, &failDiags)
		if !attrVal.IsNull() {
			t.Fatalf("%s", tc)
		}
		if !failDiags.HasError() {
			t.Fatalf("%s", tc)
		}
	}
}
