package keysManagement

import "errors"

var (
	errDuplicatedKey              = errors.New("duplicated key found")
	errMissingPublicKeyDefinition = errors.New("missing public key definition")
	errNilKeyGenerator            = errors.New("nil key generator")
	errNilP2PIdentityGenerator    = errors.New("nil p2p identity generator")
	errInvalidValue               = errors.New("invalid value")
	errMissingNamedIdentity       = errors.New("missing named identity")
)
