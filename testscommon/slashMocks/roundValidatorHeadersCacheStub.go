package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// RoundValidatorHeadersCacheStub -
type RoundValidatorHeadersCacheStub struct {
	AddCalled            func(round uint64, pubKey []byte, headerInfo data.HeaderInfoHandler) error
	GetHeadersCalled     func(round uint64, pubKey []byte) []data.HeaderInfoHandler
	GetPubKeysCalled     func(round uint64) [][]byte
	IsInterfaceNilCalled func() bool
}

// Add -
func (rvhc *RoundValidatorHeadersCacheStub) Add(round uint64, pubKey []byte, headerInfo data.HeaderInfoHandler) error {
	if rvhc.AddCalled != nil {
		return rvhc.AddCalled(round, pubKey, headerInfo)
	}
	return nil
}

// GetHeaders -
func (rvhc *RoundValidatorHeadersCacheStub) GetHeaders(round uint64, pubKey []byte) []data.HeaderInfoHandler {
	if rvhc.GetHeadersCalled != nil {
		return rvhc.GetHeadersCalled(round, pubKey)
	}
	return nil
}

// GetPubKeys -
func (rvhc *RoundValidatorHeadersCacheStub) GetPubKeys(round uint64) [][]byte {
	if rvhc.GetPubKeysCalled != nil {
		return rvhc.GetPubKeysCalled(round)
	}
	return nil
}

// IsInterfaceNil -
func (rvhc *RoundValidatorHeadersCacheStub) IsInterfaceNil() bool {
	return rvhc == nil
}
