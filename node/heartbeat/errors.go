package heartbeat

import "errors"

// ErrEmptyPublicKeyList signals that a nil or empty public key list has been provided
var ErrEmptyPublicKeyList = errors.New("nil or empty public key list")

// ErrNilMessenger signals that a nil p2p messenger has been provided
var ErrNilMessenger = errors.New("nil P2P Messenger")

// ErrNilSingleSigner signals that a nil single signer has been provided
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("nil private key")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilKeyGenerator is raised when a valid key generator is expected but nil used
var ErrNilKeyGenerator = errors.New("key generator is nil")

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")

// ErrInvalidMaxDurationPeerUnresponsive signals that the duration provided is invalid
var ErrInvalidMaxDurationPeerUnresponsive = errors.New("invalid max duration to declare the peer unresponsive")
