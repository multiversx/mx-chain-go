package mock

import "github.com/ElrondNetwork/elrond-go/process"

// RoundDetectorCacheStub -
type RoundDetectorCacheStub struct {
	AddCalled            func(round uint64, pubKey []byte, data process.InterceptedData)
	ContainsCalled       func(round uint64, pubKey []byte, data process.InterceptedData) bool
	GetDataCalled        func(round uint64, pubKey []byte) []process.InterceptedData
	GetPubKeysCalled     func(round uint64) [][]byte
	IsInterfaceNilCalled func() bool
}

func (rdc *RoundDetectorCacheStub) Add(round uint64, pubKey []byte, data process.InterceptedData) {
	if rdc.AddCalled != nil {
		rdc.AddCalled(round, pubKey, data)
	}
}

func (rdc *RoundDetectorCacheStub) Contains(round uint64, pubKey []byte, data process.InterceptedData) bool {
	if rdc.ContainsCalled != nil {
		return rdc.ContainsCalled(round, pubKey, data)
	}
	return false
}

func (rdc *RoundDetectorCacheStub) GetData(round uint64, pubKey []byte) []process.InterceptedData {
	if rdc.GetDataCalled != nil {
		return rdc.GetDataCalled(round, pubKey)
	}
	return nil
}

func (rdc *RoundDetectorCacheStub) GetPubKeys(round uint64) [][]byte {
	if rdc.GetPubKeysCalled != nil {
		return rdc.GetPubKeysCalled(round)
	}
	return nil
}

func (rdc *RoundDetectorCacheStub) IsInterfaceNil() bool {
	return rdc == nil
}
