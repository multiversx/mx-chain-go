package genericMocks

import (
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	sovereign2 "github.com/multiversx/mx-chain-go/testscommon/sovereign"
)

// DataCodecFactoryMock -
type DataCodecFactoryMock struct {
	CreateDataCodecCalled func() sovereign.DataDecoderCreator
}

// CreateDataCodec -
func (dc *DataCodecFactoryMock) CreateDataCodec() sovereign.DataDecoderHandler {
	if dc.CreateDataCodecCalled != nil {
		return dc.CreateDataCodec()
	}
	return &sovereign2.DataCodecMock{}
}

// IsInterfaceNil -
func (dc *DataCodecFactoryMock) IsInterfaceNil() bool {
	return dc == nil
}
