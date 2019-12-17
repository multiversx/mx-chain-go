package peer

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

type shardMediator struct {
	vs shardMetaMediated
}

type metaMediator struct {
	vs shardMetaMediated
}

func (lh *shardMediator) loadPreviousShardHeaders(header, previousHeader *block.MetaBlock) error {
	if lh.vs == nil {
		return process.ErrNilMediator
	}

	return lh.vs.loadPreviousShardHeaders(header, previousHeader)
}

func (lh *metaMediator) loadPreviousShardHeaders(header, previousHeader *block.MetaBlock) error {
	if lh.vs == nil {
		return process.ErrNilMediator
	}

	return lh.vs.loadPreviousShardHeadersMeta(header)
}
