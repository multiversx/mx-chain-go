package metrics

import "fmt"

// ErrNilStatusHandlerUtils signals that a nil status handler utils instance has been provided
var ErrNilStatusHandlerUtils = fmt.Errorf("nil StatusHandlerUtils when initializing metrics")
