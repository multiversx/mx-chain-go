package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type FinalityAttesterStub struct {
	GetFinalHeaderCalled func(shardID uint32) (data.HeaderHandler, []byte, error)
}

func (fas *FinalityAttesterStub) GetFinalHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if fas.GetFinalHeaderCalled != nil {
		return fas.GetFinalHeaderCalled(shardID)
	}

	hdr := &block.Header{
		Nonce:   0,
		ShardId: shardID,
	}

	return hdr, make([]byte, 0), nil
}

func (fas *FinalityAttesterStub) IsInterfaceNil() bool {
	return fas == nil
}
