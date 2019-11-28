package external

import "github.com/pkg/errors"

// ErrWrongTypeAssertion signals that a wrong type assertion occurred
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilBlockChain signals that an operation has been attempted to or with a nil blockchain
var ErrNilBlockChain = errors.New("nil block chain")

// ErrNilShardCoordinator signals that an operation has been attempted to or with a nil shard coordinator
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilStore signals that the provided storage service is nil
var ErrNilStore = errors.New("nil data storage service")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNilSCQueryService signals that a nil SC query service has been provided
var ErrNilSCQueryService = errors.New("nil SC query service")

// ErrNilStatusMetrics signals that a nil status metrics was provided
var ErrNilStatusMetrics = errors.New("nil status metrics handler")
