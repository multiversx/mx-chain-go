package health

import (
	"errors"
)

var errNilComponent = errors.New("component is nil")
var errNotDiagnosableComponent = errors.New("component is not diagnosable")
