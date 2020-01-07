package track

import (
	"github.com/pkg/errors"
)

// ErrNilBlockTrackerHandler signals that a nil block tracker handler has been provided
var ErrNilBlockTrackerHandler = errors.New("nil block tracker handler")

// ErrNilCrossNotarizer signals that a nil block notarizer handler has been provided
var ErrNilCrossNotarizer = errors.New("nil cross notarizer")

// ErrCrossNotarizedHeadersNotifier signals that a nil block notifier handler has been provided
var ErrCrossNotarizedHeadersNotifier = errors.New("nil cross notarized header notifier")

// ErrSelfNotarizedHeadersNotifier signals that a nil block notifier handler has been provided
var ErrSelfNotarizedHeadersNotifier = errors.New("nil self notarized header notifier")

// ErrNotarizedHeaderOffsetIsOutOfBound signals that a requested offset of the notarized header is out of bound
var ErrNotarizedHeaderOffsetIsOutOfBound = errors.New("requested offset of the notarized header is out of bound")
