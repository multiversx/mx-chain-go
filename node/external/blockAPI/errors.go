package blockAPI

import "errors"

// ErrInvalidOutputFormat signals that the output format type is not valid
var ErrInvalidOutputFormat = errors.New("the output format type is invalid")

// ErrShardOnlyEndpoint signals that an endpoint was called, but it is only available for shard nodes
var ErrShardOnlyEndpoint = errors.New("the endpoint is only available on shard nodes")

// ErrMetachainOnlyEndpoint signals that an endpoint was called, but it is only available for metachain nodes
var ErrMetachainOnlyEndpoint = errors.New("the endpoint is only available on metachain nodes")

var errCannotLoadMiniblocks = errors.New("cannot load miniblock(s)")
var errCannotUnmarshalMiniblocks = errors.New("cannot unmarshal miniblock(s)")
var errCannotLoadTransactions = errors.New("cannot load transaction(s)")
var errCannotUnmarshalTransactions = errors.New("cannot unmarshal transaction(s)")
var errCannotLoadReceipts = errors.New("cannot load receipt(s)")
var errCannotUnmarshalReceipts = errors.New("cannot unmarshal receipt(s)")
var errUnknownBlockRequestType = errors.New("unknown block request type")
