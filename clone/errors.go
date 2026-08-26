package clone

import "github.com/Nigel2392/errors"

const (
	CodePointerError errors.GoCode = "PointerError"
	CodeNoSteps      errors.GoCode = "NoSteps"
	CodeInvalid      errors.GoCode = "Invalid"
)

var (
	ErrNotPointer = errors.New(CodePointerError, "value is not a pointer")
	ErrNoSteps    = errors.New(CodeNoSteps, "no step found")
	ErrInvalid    = errors.New(CodeInvalid, "invalid value")
)
