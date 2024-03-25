package sovereign

import (
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
)

type dataCodecFactory struct {
}

// NewDataCodecFactory creates a new data codec factory
func NewDataCodecFactory() (*dataCodecFactory, error) {
	return &dataCodecFactory{}, nil
}

// CreateDataCodec creates a new data codec for the chain run type normal
func (dcf *dataCodecFactory) CreateDataCodec() sovereign.DataDecoderHandler {
	return disabled.NewDisabledDataCodec()
}

// IsInterfaceNil returns true if there is no value under the interface
func (dcf *dataCodecFactory) IsInterfaceNil() bool {
	return dcf == nil
}
