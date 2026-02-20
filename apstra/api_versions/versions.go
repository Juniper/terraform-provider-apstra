package apiversions

// the idea with these constants is that all references to an Apstra release
// track back to this file. When support for an old release is dropped, these
// constants should help track down code relevant to those versions.
const (
	Apstra420  = "4.2.0"
	Apstra421  = "4.2.1"
	Apstra4211 = "4.2.1.1"
	Apstra422  = "4.2.2"
	Apstra500  = "5.0.0"
	Apstra501  = "5.0.1"
	Apstra510  = "5.1.0"
	Apstra600  = "6.0.0"
	Apstra610  = "6.1.0"

	GeApstra421 = ">=" + Apstra421
	GeApstra500 = ">=" + Apstra500
	GeApstra610 = ">=" + Apstra610

	GtApstra422 = ">" + Apstra422

	LeApstra422 = "<=" + Apstra422
	LeApstra600 = "<=" + Apstra600

	LtApstra500 = "<" + Apstra500
	LtApstra610 = "<" + Apstra610
)
