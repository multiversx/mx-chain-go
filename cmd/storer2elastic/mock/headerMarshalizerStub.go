package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// HeaderMarshalizerStub -
type HeaderMarshalizerStub struct {
	UnmarshalShardHeaderCalled func(headerBytes []byte) (*block.Header, error)
	UnmarshalMetaBlockCalled   func(headerBytes []byte) (*block.MetaBlock, error)
}

// UnmarshalShardHeader -
func (h *HeaderMarshalizerStub) UnmarshalShardHeader(headerBytes []byte) (*block.Header, error) {
	if h.UnmarshalShardHeaderCalled != nil {
		return h.UnmarshalShardHeaderCalled(headerBytes)
	}

	return &block.Header{}, nil
}

// UnmarshalMetaBlock -
func (h *HeaderMarshalizerStub) UnmarshalMetaBlock(headerBytes []byte) (*block.MetaBlock, error) {
	if h.UnmarshalMetaBlockCalled != nil {
		return h.UnmarshalMetaBlockCalled(headerBytes)
	}

	return &block.MetaBlock{}, nil
}

// IsInterfaceNil -
func (h *HeaderMarshalizerStub) IsInterfaceNil() bool {
	return h == nil
}
