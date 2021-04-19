package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// HeaderMarshalizerStub -
type HeaderMarshalizerStub struct {
	UnmarshalShardHeaderCalled func(headerBytes []byte) (data.ShardHeaderHandler, error)
	UnmarshalMetaBlockCalled   func(headerBytes []byte) (data.MetaHeaderHandler, error)
}

// UnmarshalShardHeader -
func (h *HeaderMarshalizerStub) UnmarshalShardHeader(headerBytes []byte) (data.ShardHeaderHandler, error) {
	if h.UnmarshalShardHeaderCalled != nil {
		return h.UnmarshalShardHeaderCalled(headerBytes)
	}

	return &block.Header{}, nil
}

// UnmarshalMetaBlock -
func (h *HeaderMarshalizerStub) UnmarshalMetaBlock(headerBytes []byte) (data.MetaHeaderHandler, error) {
	if h.UnmarshalMetaBlockCalled != nil {
		return h.UnmarshalMetaBlockCalled(headerBytes)
	}

	return &block.MetaBlock{}, nil
}

// IsInterfaceNil -
func (h *HeaderMarshalizerStub) IsInterfaceNil() bool {
	return h == nil
}
