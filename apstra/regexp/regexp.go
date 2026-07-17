package apstraregexp

import "regexp"

const (
	alphaNumW2HLConstraintReString = "^[" + alphaNumW2HLConstraintChars + "]+$"
	alphaNumW2HLConstraintChars    = "a-zA-Z0-9_-"
	AlphaNumW2HLConstraintMsg      = "value must consist only of the following characters: `" + alphaNumW2HLConstraintChars + "`."

	freeformHostnameConstraintReString = "^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$"
	FreeformHostnameConstraintMsg      = "value must match regexp: `" + freeformHostnameConstraintReString + "`."

	hostNameConstraintReString = "^[" + hostNameConstraintChars + "]+$"
	hostNameConstraintChars    = "a-zA-Z0-9.-"
	HostNameConstraintMsg      = "value must consist only of the following characters: `" + hostNameConstraintChars + "`."

	stdNameConstraintReString = "^[" + stdNameConstraintChars + "]+$"
	stdNameConstraintChars    = "a-zA-Z0-9._-"
	StdNameConstraintMsg      = "value must consist only of the following characters: `" + stdNameConstraintChars + "`."

	alphaCharsRequiredConstraintReString = "^.*[a-zA-Z]+.*$"
	AlphaCharsRequiredConstraintMsg      = "value must contain at least one letter"

	renderableDescriptionReString = `^([!#-;=@-[\]-~]([ !#-;=@-[\]-~]*[!#-;=@-[\]-~])?)?$`
	RenderableDescriptionMsg      = "value must contain only printable ASCII characters other than `\"`, `<`, `>`, `?`, and `\\` with no leading or trailing spaces"
)

var (
	AlphaNumW2HLConstraint       = regexp.MustCompile(alphaNumW2HLConstraintReString)
	FreeformHostnameConstraint   = regexp.MustCompile(freeformHostnameConstraintReString)
	HostNameConstraint           = regexp.MustCompile(hostNameConstraintReString)
	StdNameConstraint            = regexp.MustCompile(stdNameConstraintReString)
	AlphaCharsRequiredConstraint = regexp.MustCompile(alphaCharsRequiredConstraintReString)
	RenderableDescription        = regexp.MustCompile(renderableDescriptionReString)
)
