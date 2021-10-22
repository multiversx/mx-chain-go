package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// RoundDetectorCacheStub -
type RoundDetectorCacheStub struct {
	AddCalled            func(round uint64, pubKey []byte, header *slash.HeaderInfo) error
	GetDataCalled        func(round uint64, pubKey []byte) slash.HeaderInfoList
	GetPubKeysCalled     func(round uint64) [][]byte
	IsInterfaceNilCalled func() bool
}

// Add -
func (rdc *RoundDetectorCacheStub) Add(round uint64, pubKey []byte, headerInfo *slash.HeaderInfo) error {
	if rdc.AddCalled != nil {
		return rdc.AddCalled(round, pubKey, headerInfo)
	}
	return nil
}

// GetData -
func (rdc *RoundDetectorCacheStub) GetData(round uint64, pubKey []byte) slash.HeaderInfoList {
	if rdc.GetDataCalled != nil {
		return rdc.GetDataCalled(round, pubKey)
	}
	return nil
}

// GetPubKeys -
func (rdc *RoundDetectorCacheStub) GetPubKeys(round uint64) [][]byte {
	if rdc.GetPubKeysCalled != nil {
		return rdc.GetPubKeysCalled(round)
	}
	return nil
}

// IsInterfaceNil -
func (rdc *RoundDetectorCacheStub) IsInterfaceNil() bool {
	return rdc == nil
}
