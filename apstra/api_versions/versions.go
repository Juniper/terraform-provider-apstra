package apiversions

// the idea with these constants is that all references to an Apstra release
// track back to this file. When support for an old release is dropped, these
// constants should help track down code relevant to those versions.
const (
	Apstra500 = "5.0.0"
	Apstra501 = "5.0.1"
	Apstra510 = "5.1.0"
	Apstra600 = "6.0.0"
	Apstra610 = "6.1.0"
	Apstra611 = "6.1.1"
	Apstra612 = "6.1.2"
	Apstra620 = "6.2.0"

	GeApstra610 = ">=" + Apstra610
	GeApstra620 = ">=" + Apstra620

	LeApstra600 = "<=" + Apstra600

	LtApstra610 = "<" + Apstra610
	LtApstra620 = "<" + Apstra620
)
