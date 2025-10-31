package errors

import (
	"fmt"
	"reflect"
)

func CreateError(v any) string {
	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return fmt.Sprintf(createFailed, t.Name())
}

func ReadError(v any) string {
	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return fmt.Sprintf(readFailed, t.Name())
}

func UpdateError(v any) string {
	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return fmt.Sprintf(updateFailed, t.Name())
}

func DeleteError(v any) string {
	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return fmt.Sprintf(deleteFailed, t.Name())
}
