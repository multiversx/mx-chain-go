package incomingEventsProc

import (
	"errors"
)

var errNilIncomingEventHandler = errors.New("nil incoming event handler provided")

var errNilEventProcDepositTokens = errors.New("nil event processor for deposit tokens provided")

var errNilEventProcConfirmExecutedOp = errors.New("nil event processor for confirmed executed operations provided")
