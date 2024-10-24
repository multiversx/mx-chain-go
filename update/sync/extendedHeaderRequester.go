package sync

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

type extendedHeaderRequester struct {
	requestHandler ExtendedShardHeaderRequestHandler
}

func NewExtendedHeaderRequester(requestHandler ExtendedShardHeaderRequestHandler) (*extendedHeaderRequester, error) {
	return &extendedHeaderRequester{
		requestHandler: requestHandler,
	}, nil
}

func (ehr *extendedHeaderRequester) ShouldRequestHeader(shardId uint32) bool {
	return shardId == core.MainChainShardId
}

func (ehr *extendedHeaderRequester) RequestHeader(hash []byte) {
	ehr.requestHandler.RequestExtendedShardHeader(hash)
}

func (ehr *extendedHeaderRequester) IsInterfaceNil() bool {
	return ehr == nil
}
