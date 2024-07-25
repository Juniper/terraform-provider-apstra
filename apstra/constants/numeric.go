package constants

import "math"

const (
	AsnMin = 1
	AsnMax = math.MaxUint32

	BgpHoldMin = 1
	BgpHoldMax = math.MaxUint16

	BgpKeepaliveMin = 1
	BgpKeepaliveMax = math.MaxUint16

	HoldTimeMin = 3
	HoldTimeMax = math.MaxUint16

	KeepaliveTimeMin = 1
	KeepaliveTimeMax = HoldTimeMax / 3

	L3MtuMin = 1280
	L3MtuMax = 9216

	TtlMin = 1
	TtlMax = math.MaxUint8
)
