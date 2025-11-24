package errors

import "errors"

var (
	ErrDB       = errors.New("database")
	ErrNotFound = errors.New("resource not found")
)
