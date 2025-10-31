package pointer

func To[A any](a A) *A {
	return &a
}
