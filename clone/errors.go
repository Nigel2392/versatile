package clone

import "github.com/Nigel2392/errors"

const (
	CodePointerError errors.GoCode = "PointerError"
	CodeNoSteps      errors.GoCode = "NoSteps"
	CodeInvalid      errors.GoCode = "Invalid"
	CodeObjectCached errors.GoCode = "ObjectCached"
)

var (
	ErrNotPointer   = errors.New(CodePointerError, "value is not a pointer")
	ErrNoSteps      = errors.New(CodeNoSteps, "no step found")
	ErrInvalid      = errors.New(CodeInvalid, "invalid value")
	ErrObjectCached = errors.New(CodeObjectCached, "no action, object was cached")
)
