package sync

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
)

type metaHeaderRequester struct {
	requestHandler process.RequestHandler
}

func NewMetaHeaderRequester(requestHandler process.RequestHandler) (*metaHeaderRequester, error) {
	return &metaHeaderRequester{
		requestHandler: requestHandler,
	}, nil
}

func (mhr *metaHeaderRequester) ShouldRequestHeader(shardId uint32) bool {
	return shardId == core.MetachainShardId
}

func (mhr *metaHeaderRequester) RequestHeader(hash []byte) {
	mhr.requestHandler.RequestMetaHeader(hash)
}

func (mhr *metaHeaderRequester) IsInterfaceNil() bool {
	return mhr == nil
}
