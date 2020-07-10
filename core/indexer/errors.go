package indexer

import (
	"errors"
)

// ErrNilElasticClient is returned when a nil es was passed as an argument
var ErrNilElasticClient = errors.New("nil elastic es provided")

// ErrBackOff -
var ErrBackOff = errors.New("back off")
