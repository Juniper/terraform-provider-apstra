package errors

type ResourceNotFound string

func (e ResourceNotFound) Error() string {
	return string(e)
}
