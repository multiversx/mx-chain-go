package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// ShardMetaMediatedMock is an implementation of shardMetaMediated
type ShardMetaMediatedMock struct {
	loadPreviousShardHeadersCalled func(header, previousHeader *block.MetaBlock) error
	loadPreviousShardHeadersMetaCalled func(header *block.MetaBlock) error
}

func (smm *ShardMetaMediatedMock) loadPreviousShardHeaders(header, previousHeader *block.MetaBlock) error {
	if smm.loadPreviousShardHeadersCalled != nil {
		return smm.loadPreviousShardHeadersCalled(header, previousHeader)
	}
	return nil
}

func (smm *ShardMetaMediatedMock) loadPreviousShardHeadersMeta(header *block.MetaBlock) error {
	if smm.loadPreviousShardHeadersMetaCalled != nil {
		return smm.loadPreviousShardHeadersMetaCalled(header)
	}
	return nil
}