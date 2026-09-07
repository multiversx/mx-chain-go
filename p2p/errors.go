package p2p

import (
	"errors"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = p2p.ErrNilMessage

// ErrMessageShouldBeIgnored signals that a valid message should not be processed or propagated
var ErrMessageShouldBeIgnored = p2p.ErrMessageShouldBeIgnored

// ErrNilPreferredPeersHolder signals that a nil preferred peers holder was provided
var ErrNilPreferredPeersHolder = p2p.ErrNilPreferredPeersHolder

// ErrNilStatusHandler signals that a nil status handler has been provided
var ErrNilStatusHandler = errors.New("nil status handler")
