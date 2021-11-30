package app

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Validate a struct against the given validation.FieldRules. This wrapper handles
// the internal errors that can be returned from validation. Panics if an internal error
// occurs or the error type is unknown.
func Validate(structPtr interface{}, fields ...*validation.FieldRules) validation.Errors {
	err := validation.ValidateStruct(structPtr, fields...)

	if err == nil {
		return nil
	}

	var e validation.InternalError
	if errors.As(err, &e) {
		// Panic here as if the internal validator fails we want to know
		// straight away (this should be picked up in development) as the
		// errors are usually down to invalid arguments etc.
		panic(e)
	}

	var errs validation.Errors
	if !errors.As(err, &errs) {
		panic("validator should know the type at this point")
	}

	return errs
}
