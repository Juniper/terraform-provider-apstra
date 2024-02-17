package utils

type Stringer struct {
	s string
}

func (o Stringer) String() string {
	return o.s
}

func NewStringer(s string) Stringer {
	return Stringer{s: s}
}

func ToPtr[A any](a A) *A {
	return &a
}
