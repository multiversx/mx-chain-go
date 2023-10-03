package keysManagement

import "errors"

// ErrDuplicatedKey signals that a key is already managed by the node
var ErrDuplicatedKey = errors.New("duplicated key found")

// ErrMissingPublicKeyDefinition signals that a public key definition is missing
var ErrMissingPublicKeyDefinition = errors.New("missing public key definition")

// ErrNilKeyGenerator signals that a nil key generator was provided
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrInvalidKey signals that an invalid key was provided
var ErrInvalidKey = errors.New("invalid key")

// ErrNilManagedPeersHolder signals that a nil managed peers holder was provided
var ErrNilManagedPeersHolder = errors.New("nil managed peers holder")

// ErrNilPrivateKey signals that a nil private key was provided
var ErrNilPrivateKey = errors.New("nil private key")

// ErrEmptyPeerID signals that an empty peer ID was provided
var ErrEmptyPeerID = errors.New("empty peer ID")

// ErrNilP2PKeyConverter signals that a nil p2p key converter has been provided
var ErrNilP2PKeyConverter = errors.New("nil p2p key converter")

// ErrNilNodesCoordinator signals that a nil nodes coordinator has been provided
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrNilShardProvider signals that a nil shard provider has been provided
var ErrNilShardProvider = errors.New("nil shard provider")

// ErrNilEpochProvider signals that a nil epoch provider has been provided
var ErrNilEpochProvider = errors.New("nil epoch provider")
