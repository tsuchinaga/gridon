package gridon

import (
	"errors"
)

var (
	ErrUnknown                 = errors.New("unknown")
	ErrNilArgument             = errors.New("nil argument")
	ErrNoData                  = errors.New("no data")
	ErrDuplicateData           = errors.New("duplicate data")
	ErrOrderCondition          = errors.New("order condition")
	ErrCancelCondition         = errors.New("cancel condition")
	ErrNotEnoughCash           = errors.New("not enough cash")
	ErrNotEnoughPosition       = errors.New("not enough position")
	ErrUndecidableValue        = errors.New("undecidable value")
	ErrAlreadyExists           = errors.New("already exists")
	ErrNotFound                = errors.New("not found")
	ErrNotExistsTimeRange      = errors.New("not exists time range")
	ErrCannotGetBasePrice      = errors.New("can not get base price")
	ErrZeroGridWidth           = errors.New("zero grid width")
	ErrShortSellingRestriction = errors.New("short selling restriction")
)
