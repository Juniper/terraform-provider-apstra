package sysredundancycache

var (
	// These maps are exported/public only while running tests, never in production code.
	BPToGroupToSystems = bpToGroupToSystems
	BPToSystemToGroup  = bpToSystemToGroup
)
