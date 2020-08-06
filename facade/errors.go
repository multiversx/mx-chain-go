package facade

import "github.com/pkg/errors"

// ErrHeartbeatsNotActive signals that the heartbeat system is not active
var ErrHeartbeatsNotActive = errors.New("heartbeat system not active")

// ErrNilNode signals that a nil node instance has been provided
var ErrNilNode = errors.New("nil node")

// ErrNilApiResolver signals that a nil api resolver instance has been provided
var ErrNilApiResolver = errors.New("nil api resolver")

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNoApiRoutesConfig signals that no configuration was found for API routes
var ErrNoApiRoutesConfig = errors.New("no configuration found for API routes")

// ErrNilPeerState signals that a nil peer state has been provided
var ErrNilPeerState = errors.New("nil peer state")

// ErrNilAccountState signals that a nil account state has been provided
var ErrNilAccountState = errors.New("nil account state")
