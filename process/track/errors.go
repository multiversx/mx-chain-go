package track

import (
	"github.com/pkg/errors"
)

// ErrNilBlockTrackerHandler signals that a nil block tracker handler has been provided
var ErrNilBlockTrackerHandler = errors.New("nil block tracker handler")

// ErrNilCrossNotarizer signals that a nil block notarizer handler has been provided
var ErrNilCrossNotarizer = errors.New("nil cross notarizer")

// ErrNilSelfNotarizer signals that a nil block notarizer handler has been provided
var ErrNilSelfNotarizer = errors.New("nil self notarizer")

// ErrNilCrossNotarizedHeadersNotifier signals that a nil block notifier handler has been provided
var ErrNilCrossNotarizedHeadersNotifier = errors.New("nil cross notarized header notifier")

// ErrNilSelfNotarizedFromCrossHeadersNotifier signals that a nil block notifier handler has been provided
var ErrNilSelfNotarizedFromCrossHeadersNotifier = errors.New("nil self notarized from cross header notifier")

// ErrNilSelfNotarizedHeadersNotifier signals that a nil block notifier handler has been provided
var ErrNilSelfNotarizedHeadersNotifier = errors.New("nil self notarized header notifier")

// ErrNilFinalMetachainHeadersNotifier signals that a nil block notifier handler has been provided
var ErrNilFinalMetachainHeadersNotifier = errors.New("nil final metachain header notifier")

// ErrNotarizedHeaderOffsetIsOutOfBound signals that a requested offset of the notarized header is out of bound
var ErrNotarizedHeaderOffsetIsOutOfBound = errors.New("requested offset of the notarized header is out of bound")

// ErrNilRoundHandler signals that a nil roundHandler has been provided
var ErrNilRoundHandler = errors.New("nil roundHandler")
