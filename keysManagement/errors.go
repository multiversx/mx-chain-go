package keysManagement

import "errors"

var errDuplicatedKey = errors.New("duplicated key found")
var errMissingPublicKeyDefinition = errors.New("missing public key definition")
var errNilKeyGenerator = errors.New("nil key generator")
var errNilP2PIdentityGenerator = errors.New("nil p2p identity generator")
