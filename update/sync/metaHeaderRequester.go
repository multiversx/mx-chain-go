package sync

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type metaHeaderRequester struct {
	requestHandler process.RequestHandler
}

// NewMetaHeaderRequester creates a new meta header requester wrapper
func NewMetaHeaderRequester(requestHandler process.RequestHandler) (*metaHeaderRequester, error) {
	if check.IfNil(requestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	return &metaHeaderRequester{
		requestHandler: requestHandler,
	}, nil
}

// ShouldRequestHeader returns true if the shard id is metachain
func (mhr *metaHeaderRequester) ShouldRequestHeader(shardId uint32) bool {
	return shardId == core.MetachainShardId
}

// RequestHeader requests meta header by hash
func (mhr *metaHeaderRequester) RequestHeader(hash []byte) {
	mhr.requestHandler.RequestMetaHeader(hash)
}

// IsInterfaceNil checks if underlying pointer is nil
func (mhr *metaHeaderRequester) IsInterfaceNil() bool {
	return mhr == nil
}
