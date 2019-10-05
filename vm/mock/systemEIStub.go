package mock

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type SystemEIStub struct {
	TransferCalled func (destination []byte, sender []byte, value *big.Int, input []byte) error
	GetBalanceCalled func(addr []byte) *big.Int
	SetStorageCalled func(addr []byte, key []byte, value []byte)
	GetStorageCalled func(addr []byte, key []byte) []byte
	SelfDestructCalled func(addr []byte, beneficiary []byte)
	CreateVMOutputCalled func() *vmcommon.VMOutput
	CleanCacheCalled func()
}

func (s *SystemEIStub) Transfer(destination []byte, sender []byte, value *big.Int, input []byte) error {
	if s.TransferCalled != nil {
		return s.TransferCalled(destination, sender, value, input)
	}
	return nil
}

func (s *SystemEIStub) GetBalance(addr []byte) *big.Int {
	if s.GetBalanceCalled != nil {
		return s.GetBalanceCalled(addr)
	}
	return big.NewInt(0)
}

func (s *SystemEIStub) SetStorage(addr []byte, key []byte, value []byte) {
	if s.SetStorageCalled != nil {
		s.SetStorageCalled(addr, key, value)
	}
}

func (s *SystemEIStub) GetStorage(addr []byte, key []byte) []byte {
	if s.GetStorageCalled != nil {
		return s.GetStorageCalled(addr, key)
	}
	return nil
}

func (s *SystemEIStub) SelfDestruct(addr []byte, beneficiary []byte) {
	if s.SelfDestructCalled != nil {
		s.SelfDestructCalled(addr, beneficiary)
	}
	return
}

func (s *SystemEIStub) CreateVMOutput() *vmcommon.VMOutput {
	if s.CreateVMOutputCalled != nil {
		return s.CreateVMOutputCalled()
	}

	return &vmcommon.VMOutput{}
}

func (s *SystemEIStub) CleanCache() {
	if s.CleanCacheCalled != nil {
		s.CleanCacheCalled()
	}
	return
}

func (s *SystemEIStub) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}

