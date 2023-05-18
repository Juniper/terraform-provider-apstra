package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"strings"
)

const diagnosticWrapperDefaultSeparator = " : "

var _ diag.Diagnostic = diagnosticWrapper{}

// diagnosticWrapper is an implementation of diag.Diagnostic which carries both
// a diag.Diagnostic and a slice of "wrapper" strings which are intended to
// provide additional context about the diagnostic message.
type diagnosticWrapper struct {
	diagnostic   diag.Diagnostic
	wrapMessages []string
	separator    string
}

func (o diagnosticWrapper) Severity() diag.Severity {
	return o.diagnostic.Severity()
}

func (o diagnosticWrapper) Summary() string {
	return o.diagnostic.Summary()
}

// Detail prints the stack of diagnostic messages and the detailed message from
// the underlying diagnostic event separated by diagnosticWrapperDefaultSeparator
func (o diagnosticWrapper) Detail() string {
	details := make([]string, len(o.wrapMessages))
	for i := range o.wrapMessages {
		details[i] = o.wrapMessages[i]
	}
	Reverse(details)
	return strings.Join(append(details, o.diagnostic.Detail()), o.separator)
}

// Equal compares the embedded diag.Diagnostic against 'in'. Note that
// "equality" now is no longer commutative. a.Equal(b) != b.Equal(a).
// ...
// yeah.
func (o diagnosticWrapper) Equal(in diag.Diagnostic) bool {
	return o.diagnostic.Equal(in)
}

func (o *diagnosticWrapper) SetSeparator(in string) {
	o.separator = in
}

func WrapDiagnostic(in diag.Diagnostic, detail string) diag.Diagnostic {
	if in == nil {
		return nil
	}

	if inWrapper, ok := in.(diagnosticWrapper); ok {
		inWrapper.wrapMessages = append(inWrapper.wrapMessages, detail)
		return inWrapper
	}

	return diagnosticWrapper{
		diagnostic:   in,
		wrapMessages: []string{detail},
		separator:    diagnosticWrapperDefaultSeparator,
	}
}

func WrapEachDiagnostic(in diag.Diagnostics, prefix string) diag.Diagnostics {
	for i, d := range in {
		in[i] = WrapDiagnostic(d, prefix)
	}
	return in
}
