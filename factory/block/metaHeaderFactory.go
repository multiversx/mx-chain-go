package block

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// HeaderVersionGetter can get the header version based on epoch
type HeaderVersionGetter interface {
	GetVersion(epoch uint32) string
	IsInterfaceNil() bool
}

type metaHeaderVersionHandler struct {
	headerVersionHandler HeaderVersionGetter
}

// NewMetaHeaderFactory creates a meta header factory instance
func NewMetaHeaderFactory(headerVersionHandler HeaderVersionGetter) (*metaHeaderVersionHandler, error) {
	if check.IfNil(headerVersionHandler) {
		return nil, ErrNilHeaderVersionHandler
	}
	return &metaHeaderVersionHandler{
		headerVersionHandler: headerVersionHandler,
	}, nil
}

// Create creates a metaBlock instance with the correct version and format, according to the epoch
func (mhv *metaHeaderVersionHandler) Create(epoch uint32) data.HeaderHandler {
	version := mhv.headerVersionHandler.GetVersion(epoch)

	return &block.MetaBlock{
		Epoch:           epoch,
		SoftwareVersion: []byte(version),
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mhv *metaHeaderVersionHandler) IsInterfaceNil() bool {
	return mhv == nil
}
