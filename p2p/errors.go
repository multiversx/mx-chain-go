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
<<<<<<< HEAD

// ErrMessageUnmarshalError signals that an invalid message was received from a peer. There is no way to communicate
// with such a peer as it does not respect the protocol
var ErrMessageUnmarshalError = errors.New("message unmarshal error")

// ErrUnsupportedFields signals that unsupported fields are provided
var ErrUnsupportedFields = errors.New("unsupported fields")

// ErrUnsupportedMessageVersion signals that an unsupported message version was detected
var ErrUnsupportedMessageVersion = errors.New("unsupported message version")

// ErrNilSyncTimer signals that a nil sync timer was provided
var ErrNilSyncTimer = errors.New("nil sync timer")

// ErrNilPreferredPeersHolder signals that a nil preferred peers holder was provided
var ErrNilPreferredPeersHolder = errors.New("nil peers holder")

// ErrInvalidSeedersReconnectionInterval signals that an invalid seeders reconnection interval error occurred
var ErrInvalidSeedersReconnectionInterval = errors.New("invalid seeders reconnection interval")

// ErrMessageProcessorAlreadyDefined signals that a message processor was already defined on the provided topic and identifier
var ErrMessageProcessorAlreadyDefined = errors.New("message processor already defined")

// ErrMessageProcessorDoesNotExists signals that a message processor does not exist on the provided topic and identifier
var ErrMessageProcessorDoesNotExists = errors.New("message processor does not exists")

// ErrWrongTypeAssertions signals that a wrong type assertion occurred
var ErrWrongTypeAssertions = errors.New("wrong type assertion")

// ErrNilConnectionsWatcher signals that a nil connections watcher has been provided
var ErrNilConnectionsWatcher = errors.New("nil connections watcher")

// ErrNilPeersRatingHandler signals that a nil peers rating handler has been provided
var ErrNilPeersRatingHandler = errors.New("nil peers rating handler")

// ErrNilCacher signals that a nil cacher has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilP2PSigner signals that a nil p2p signer has been provided
var ErrNilP2PSigner = errors.New("nil p2p signer")
=======
>>>>>>> rc/v1.4.0
