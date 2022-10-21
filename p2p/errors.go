package p2p

import (
	"errors"

	p2p "github.com/ElrondNetwork/elrond-go-p2p"
)

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = p2p.ErrNilMessage

// ErrNilPreferredPeersHolder signals that a nil preferred peers holder was provided
var ErrNilPreferredPeersHolder = p2p.ErrNilPreferredPeersHolder

// ErrNilStatusHandler signals that a nil status handler has been provided
var ErrNilStatusHandler = errors.New("nil status handler")
