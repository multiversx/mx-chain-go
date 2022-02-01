package block

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

type shardHeaderFactory struct {
	headerVersionHandler HeaderVersionGetter
}

// NewShardHeaderFactory creates a shard header factory instance
func NewShardHeaderFactory(headerVersionHandler HeaderVersionGetter) (*shardHeaderFactory, error) {
	if check.IfNil(headerVersionHandler) {
		return nil, ErrNilHeaderVersionHandler
	}

	return &shardHeaderFactory{
		headerVersionHandler: headerVersionHandler,
	}, nil
}

// Create creates a header instance with the correct version and format, according to the epoch
func (shf *shardHeaderFactory) Create(epoch uint32) data.HeaderHandler {
	version := shf.headerVersionHandler.GetVersion(epoch)

	switch version {
	case "2":
		return &block.HeaderV2{
			Header: &block.Header{
				Epoch:           epoch,
				SoftwareVersion: []byte(version),
			},
			ScheduledRootHash: nil,
		}
	default:
		return &block.Header{
			Epoch:           epoch,
			SoftwareVersion: []byte(version),
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (shf *shardHeaderFactory) IsInterfaceNil() bool {
	return shf == nil
}
