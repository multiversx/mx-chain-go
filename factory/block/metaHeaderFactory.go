package block

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
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
