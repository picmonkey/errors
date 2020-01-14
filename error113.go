// +build go1.13

package errors

import "errors"

var (
	As     = errors.As
	Unwrap = errors.Unwrap
	is     = errors.Is
)
