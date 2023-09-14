package ent

import "fmt"

func CustomValidationError(name string, format string, a ...any) error {
	return &ValidationError{
		Name: name,
		err:  fmt.Errorf(format, a...),
	}
}
