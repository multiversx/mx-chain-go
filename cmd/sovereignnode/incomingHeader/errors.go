package incomingHeader

import "errors"

var errNilHeadersPool = errors.New("nil headers pool provided")

var errNilTxPool = errors.New("nil tx pool provided")

var errInvalidHeaderType = errors.New("incoming header is not of type HeaderV2")

var errInvalidEventType = errors.New("incoming event is not of type transaction event")

var errInvalidNumTopicsIncomingEvent = errors.New("received invalid number of topics in incoming event")

var errInvalidIncomingEventIdentifier = errors.New("received invalid/unknown incoming event identifier")
