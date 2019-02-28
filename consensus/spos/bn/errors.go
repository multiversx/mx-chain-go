package bn

import (
	"errors"
)

// ErrRoundCanceled is raised when round is canceled
var ErrRoundCanceled = errors.New("round is canceled")

// ErrSenderNotOk is raised when sender is invalid
var ErrSenderNotOk = errors.New("sender is invalid")

// ErrMessageForPastRound is raised when message is for past round
var ErrMessageForPastRound = errors.New("message is for past round")

// ErrInvalidSignature is raised when signature is invalid
var ErrInvalidSignature = errors.New("signature is invalid")

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")
