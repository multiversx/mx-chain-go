package factory

import "errors"

// errPublicKeyMismatch signals that the read public key mismatch the one read
var errPublicKeyMismatch = errors.New("public key mismatch between the computed and the one read from the file")

// errWrongTypeAssertion signals that a wrong type assertion occurred
var errWrongTypeAssertion = errors.New("wrong type assertion")
