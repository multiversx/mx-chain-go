package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type InterceptedTxBlockBody struct {
	*block.TxBlockBody
	hash    []byte
	shardID uint32
}

// NewInterceptedTxBlockBody creates a new instance of InterceptedTxBlockBody struct
func NewInterceptedTxBlockBody() *InterceptedTxBlockBody {
	return &InterceptedTxBlockBody{}
}

func (inTxBlkBdy *InterceptedTxBlockBody) Check() bool {
	return true
}

func (inTxBlkBdy *InterceptedTxBlockBody) SetHash(hash []byte) {
	inTxBlkBdy.hash = hash
}

func (inTxBlkBdy *InterceptedTxBlockBody) Hash() []byte {
	return inTxBlkBdy.hash
}

func (inTxBlkBdy *InterceptedTxBlockBody) New() p2p.Newer {
	return NewInterceptedTxBlockBody()
}

func (inTxBlkBdy *InterceptedTxBlockBody) ID() string {
	return string(inTxBlkBdy.hash)
}

func (inTxBlkBdy *InterceptedTxBlockBody) GetTxBlockBody() *block.TxBlockBody {
	return inTxBlkBdy.TxBlockBody
}

func (inTxBlkBdy *InterceptedTxBlockBody) Shard() uint32 {
	return inTxBlkBdy.shardID
}
