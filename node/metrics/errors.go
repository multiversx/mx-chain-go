package metrics

import "fmt"

// ErrNilAppStatusHandler signals that a nil app status handler instance has been provided
var ErrNilAppStatusHandler = fmt.Errorf("nil app status handler when initializing metrics")
