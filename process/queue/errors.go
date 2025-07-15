package queue

import "errors"

// ErrNoHeaderForProcessing indicates that there are no headers available in the queue for processing.
var ErrNoHeaderForProcessing = errors.New("no header for processing")
