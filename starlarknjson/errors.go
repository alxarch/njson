package starlarknjson

import (
	"fmt"
	"go.starlark.net/starlark"
)

type ValueError struct {
	err error
}
func newValueError(err error) error {
	return &ValueError{
		err: err,
	}
}
func (e *ValueError) Error() string {
	return "ValueError: " + e.err.Error()
}
func (e *ValueError) Unwrap() error {
	return e.err
}

type KeyError string

func (e KeyError) Error() string {
	return "KeyError: " + string(e)
}

func methodError(method *starlark.Builtin, format string, args ...interface{}) string {
	if recv := method.Receiver(); recv != nil {
		args = append([]interface{}{recv, method.Name()}, args...)
		format = "%s.%s() " + format
		return fmt.Sprintf(format, args...)
	}
	args = append([]interface{}{method.Name()}, args...)
	format = "%s()" + format
	return fmt.Sprintf(format, args...)
}

type TypeError string

func (e TypeError) Error() string {
	return "TypeError: " + string(e)
}

