package sovereign

import (
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type sovereignDataCodecFactory struct {
	dataCodec sovereign.DataDecoderHandler
}

// NewSovereignDataCodecFactory creates a new sovereign data codec factory
func NewSovereignDataCodecFactory(dc sovereign.DataDecoderHandler) (*sovereignDataCodecFactory, error) {
	if check.IfNil(dc) {
		return nil, errors.ErrNilDataCodec
	}

	return &sovereignDataCodecFactory{
		dataCodec: dc,
	}, nil
}

// CreateDataCodec creates a new data codec for the chain run type sovereign
func (sdcf *sovereignDataCodecFactory) CreateDataCodec() sovereign.DataDecoderHandler {
	return sdcf.dataCodec
}

// IsInterfaceNil returns true if there is no value under the interface
func (sdcf *sovereignDataCodecFactory) IsInterfaceNil() bool {
	return sdcf == nil
}
