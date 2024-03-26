package genericMocks

import (
	factorySovereign "github.com/multiversx/mx-chain-go/factory/sovereign"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	sovereignMock "github.com/multiversx/mx-chain-go/testscommon/sovereign"
)

// DataCodecFactoryMock -
type DataCodecFactoryMock struct {
	CreateDataCodecCalled func() factorySovereign.DataDecoderCreator
}

// CreateDataCodec -
func (dc *DataCodecFactoryMock) CreateDataCodec() sovereign.DataDecoderHandler {
	if dc.CreateDataCodecCalled != nil {
		return dc.CreateDataCodec()
	}
	return &sovereignMock.DataCodecMock{}
}

// IsInterfaceNil -
func (dc *DataCodecFactoryMock) IsInterfaceNil() bool {
	return dc == nil
}
