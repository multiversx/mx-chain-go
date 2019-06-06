package indexer

import (
	"errors"
)

// ErrCannotCreateIndex signals that we could not creeate an elasticsearch index
var ErrCannotCreateIndex = errors.New("cannot create elasitc index")
