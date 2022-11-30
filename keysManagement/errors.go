package keysManagement

import "errors"

var (
	errDuplicatedKey              = errors.New("duplicated key found")
	errMissingPublicKeyDefinition = errors.New("missing public key definition")
	errNilKeyGenerator            = errors.New("nil key generator")
	errNilP2PIdentityGenerator    = errors.New("nil p2p identity generator")
	errInvalidValue               = errors.New("invalid value")
	errInvalidKey                 = errors.New("invalid key")
	errNilManagedPeersHolder      = errors.New("nil managed peers holder")
	errNilPrivateKey              = errors.New("nil private key")
	errEmptyPeerID                = errors.New("empty peer ID")
)
