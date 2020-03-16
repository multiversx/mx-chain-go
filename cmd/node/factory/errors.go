package factory

import "errors"

// errPublicKeyMismatch signals that the read public key mismatch the one read
var errPublicKeyMismatch = errors.New("public key mismatch between the computed and the one read from the file")
