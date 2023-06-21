package p2p

import (
	"errors"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = p2p.ErrNilMessage

// ErrNilPreferredPeersHolder signals that a nil preferred peers holder was provided
var ErrNilPreferredPeersHolder = p2p.ErrNilPreferredPeersHolder

// ErrNilStatusHandler signals that a nil status handler has been provided
var ErrNilStatusHandler = errors.New("nil status handler")

// ErrUnknownNetwork signals that an unknown network has been provided
var ErrUnknownNetwork = errors.New("unknown network")
