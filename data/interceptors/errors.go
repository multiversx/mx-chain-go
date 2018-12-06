package interceptors

import (
	"github.com/pkg/errors"
)

// ErrNilMessenger signals that a nil Messenger object was provided
var ErrNilMessenger = errors.New("nil Messenger")

// ErrNilNewer signals that a nil Newer object was provided
var ErrNilNewer = errors.New("nil Newer")

// ErrRegisteringValidator signals that a registration validator occur
var ErrRegisteringValidator = errors.New("error while registering validator")

// ErrNilInterceptor signals that a nil interceptor has been provided
var ErrNilInterceptor = errors.New("nil interceptor")

// ErrNilBlockchain signals that a nil blockchain has been provided
var ErrNilBlockchain = errors.New("nil blockchain")
