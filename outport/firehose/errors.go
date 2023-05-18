package firehose

import "errors"

var errNilWriter = errors.New("nil writer provided")

var errNilBlockCreator = errors.New("nil block creator provided")

var errNilOutportBlock = errors.New("received nil outport block")
