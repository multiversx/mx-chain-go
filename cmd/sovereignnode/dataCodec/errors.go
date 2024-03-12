package dataCodec

import "errors"

var errEmptyData = errors.New("empty bytes to deserialize event data")

var errEmptyTokenData = errors.New("empty bytes to deserialize token data")
