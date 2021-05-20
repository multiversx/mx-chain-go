package block

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
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
	}

	return &block.Header{
		Epoch:           epoch,
		SoftwareVersion: []byte(version),
	}
}
