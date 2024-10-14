package apiversions

// the idea with these constants is that all references to an Apstra release
// track back to this file. When support for an old release is dropped, these
// constants should help track down code relevant to those versions.
const (
	Apstra410  = "4.1.0"
	Apstra411  = "4.1.1"
	Apstra412  = "4.1.2"
	Apstra420  = "4.2.0"
	Apstra421  = "4.2.1"
	Apstra4211 = "4.2.1.1"
	Apstra422  = "4.2.2"
	Apstra500  = "5.0.0"

	LtApstra500 = "<" + Apstra500
)
