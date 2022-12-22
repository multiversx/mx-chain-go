package firehose

import "errors"

var errNilWriter = errors.New("nil writer provided")

var errNilHeader = errors.New("received nil header")

var errInvalidHeaderType = errors.New("received invalid/unknown header type")

var errCannotCastTransaction = errors.New("cannot cast transaction")

var errCannotCastSCR = errors.New("cannot cast smart contract result")

var errCannotCastReward = errors.New("cannot cast reward transaction")

var errCannotCastReceipt = errors.New("cannot cast receipt transaction")

var errCannotCastEvent = errors.New("cannot cast event")

var errCannotCastBlockBody = errors.New("cannot cast block body")
