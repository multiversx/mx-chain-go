package incomingEventsProc

import (
	"errors"
)

var errNilIncomingEventHandler = errors.New("nil incoming event handler provided")
