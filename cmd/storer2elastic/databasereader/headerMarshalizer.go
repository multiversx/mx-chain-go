package databasereader

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

type headerMarshalizer struct {
	marshalizer marshal.Marshalizer
}

// NewHeaderMarshalizer returns a new instance of a headerMarshalizer
func NewHeaderMarshalizer(marshalizer marshal.Marshalizer) (*headerMarshalizer, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}

	return &headerMarshalizer{marshalizer: marshalizer}, nil
}

// UnmarshalShardHeader will unmarshal a shard header from the received bytes
func (hm *headerMarshalizer) UnmarshalShardHeader(headerBytes []byte) (data.HeaderHandler, error) {
	var shardHeader block.Header
	err := hm.marshalizer.Unmarshal(&shardHeader, headerBytes)
	if err != nil {
		return nil, err
	}

	return &shardHeader, nil
}

// UnmarshalMetaBlock will unmarshal a meta block the received bytes
func (hm *headerMarshalizer) UnmarshalMetaBlock(headerBytes []byte) (data.HeaderHandler, error) {
	var metaBlock block.MetaBlock
	err := hm.marshalizer.Unmarshal(&metaBlock, headerBytes)
	if err != nil {
		return nil, err
	}

	return &metaBlock, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hm *headerMarshalizer) IsInterfaceNil() bool {
	return hm == nil
}
