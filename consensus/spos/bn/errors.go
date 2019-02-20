package bn

import (
	"errors"
)

// ErrRoundCanceled is raised when round is canceled
var ErrRoundCanceled = errors.New("round is canceled")

// ErrInvalidConsensusData is raised when consensus data is invalid
var ErrInvalidConsensusData = errors.New("consensus data is invalid")

// ErrSenderNotOk is raised when sender is invalid
var ErrSenderNotOk = errors.New("sender is invalid")

// ErrMessageForPastRound is raised when message is for past round
var ErrMessageForPastRound = errors.New("message is for past round")

// ErrMessageFromItself is raised when message is from itself
var ErrMessageFromItself = errors.New("message is from itself")

// ErrInvalidSignature is raised when signature is invalid
var ErrInvalidSignature = errors.New("signature is invalid")
