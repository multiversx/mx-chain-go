package mock

import "github.com/ElrondNetwork/elrond-go/process"

// RoundDetectorCacheStub -
type RoundDetectorCacheStub struct {
	AddCalled            func(round uint64, pubKey []byte, data process.InterceptedData) error
	GetDataCalled        func(round uint64, pubKey []byte) []process.InterceptedData
	GetPubKeysCalled     func(round uint64) [][]byte
	IsInterfaceNilCalled func() bool
}

// Add -
func (rdc *RoundDetectorCacheStub) Add(round uint64, pubKey []byte, data process.InterceptedData) error {
	if rdc.AddCalled != nil {
		return rdc.AddCalled(round, pubKey, data)
	}
	return nil
}

// GetData -
func (rdc *RoundDetectorCacheStub) GetData(round uint64, pubKey []byte) []process.InterceptedData {
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
