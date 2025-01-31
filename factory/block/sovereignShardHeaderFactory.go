package block

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type sovereignShardHeaderFactory struct {
	headerVersionHandler HeaderVersionGetter
}

// NewSovereignShardHeaderFactory creates a sovereign shard header factory instance
func NewSovereignShardHeaderFactory(headerVersionHandler HeaderVersionGetter) (*sovereignShardHeaderFactory, error) {
	if check.IfNil(headerVersionHandler) {
		return nil, ErrNilHeaderVersionHandler
	}

	return &sovereignShardHeaderFactory{
		headerVersionHandler: headerVersionHandler,
	}, nil
}

// Create creates a sovereign header instance with the correct version and format, according to the epoch
func (shf *sovereignShardHeaderFactory) Create(epoch uint32) data.HeaderHandler {
	version := shf.headerVersionHandler.GetVersion(epoch)

	switch version {
	default:
		return &block.SovereignChainHeader{
			Header: &block.Header{
				Epoch:           epoch,
				SoftwareVersion: []byte(version),
			},
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (shf *sovereignShardHeaderFactory) IsInterfaceNil() bool {
	return shf == nil
}
