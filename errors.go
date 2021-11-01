package gridon

import "errors"

var (
	ErrUnknown        = errors.New("unknown")
	ErrNilArgument    = errors.New("nil argument")
	ErrOrderCondition = errors.New("orders condition")
)
