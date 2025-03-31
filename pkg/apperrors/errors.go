package apperrors

import (
	"errors"
	"fmt"
)

type Code string

const (
	Conflict      Code = "CONFLICT"
	Internal      Code = "INTERNAL"
	NotFound      Code = "NOT_FOUND"
	BadRequest    Code = "BAD_REQUEST"
	AlreadyExists Code = "ALREADY_EXISTS"
)

type AppError struct {
	Code    Code
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) Is(target error) bool {
	var t *AppError
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func IsCode(err error, code Code) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code == code
	}
	return false
}

func GetCode(err error) Code {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return Internal
}

type ErrorBuilder struct {
	code    Code
	message string
	err     error
}

func NewBadRequest() *ErrorBuilder {
	return &ErrorBuilder{code: BadRequest}
}

func NewNotFound() *ErrorBuilder {
	return &ErrorBuilder{code: NotFound}
}

func NewConflict() *ErrorBuilder {
	return &ErrorBuilder{code: Conflict}
}

func NewInternal() *ErrorBuilder {
	return &ErrorBuilder{code: Internal}
}

func NewAlreadyExists() *ErrorBuilder { return &ErrorBuilder{code: AlreadyExists} }

func (b *ErrorBuilder) WithDescription(desc string) *AppError {
	b.message = desc
	return &AppError{
		Code:    b.code,
		Message: b.message,
		Err:     b.err,
	}
}

func (b *ErrorBuilder) WithCause(err error) *AppError {
	b.err = err
	return &AppError{
		Code:    b.code,
		Message: b.message,
		Err:     b.err,
	}
}

func (b *ErrorBuilder) WithDescriptionAndCause(desc string, cause error) *AppError {
	b.message = desc
	b.err = cause
	return &AppError{
		Code:    b.code,
		Message: b.message,
		Err:     b.err,
	}
}
