package registry

import "github.com/Nigel2392/errors"

const (
	CodeNotFound errors.GoCode = "NotFound"
)

var (
	ErrNotFound errors.AbstractError[errors.Error] = errors.New(CodeNotFound, "not found")
)
