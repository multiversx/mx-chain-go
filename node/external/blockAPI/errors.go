package blockAPI

import "errors"

// ErrInvalidOutputFormat signals that the output format type is not valid
var ErrInvalidOutputFormat = errors.New("the output format type is invalid")

// ErrShardOnlyEndpoint signals that an endpoint was called, but it is only available for shard nodes
var ErrShardOnlyEndpoint = errors.New("the endpoint is only available on shard nodes")

// ErrMetachainOnlyEndpoint signals that an endpoint was called, but it is only available for metachain nodes
var ErrMetachainOnlyEndpoint = errors.New("the endpoint is only available on metachain nodes")

// ErrWrongTypeAssertion signals that an type assertion failed
var ErrWrongTypeAssertion = errors.New("wrong type assertion")
