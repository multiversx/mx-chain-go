package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SystemEIStub -
type SystemEIStub struct {
	TransferCalled                  func(destination []byte, sender []byte, value *big.Int, input []byte) error
	GetBalanceCalled                func(addr []byte) *big.Int
	SetStorageCalled                func(key []byte, value []byte)
	AddReturnMessageCalled          func(msg string)
	GetStorageCalled                func(key []byte) []byte
	SelfDestructCalled              func(beneficiary []byte)
	CreateVMOutputCalled            func() *vmcommon.VMOutput
	CleanCacheCalled                func()
	FinishCalled                    func(value []byte)
	AddCodeCalled                   func(addr []byte, code []byte)
	AddTxValueToSmartContractCalled func(value *big.Int, scAddress []byte)
	BlockChainHookCalled            func() vmcommon.BlockchainHook
	CryptoHookCalled                func() vmcommon.CryptoHook
	UseGasCalled                    func(gas uint64) error
	IsValidatorCalled               func(blsKey []byte) bool
	ExecuteOnDestContextCalled      func(destination, sender []byte, value *big.Int, input []byte) (*vmcommon.VMOutput, error)
}

// IsValidator -
func (s *SystemEIStub) IsValidator(blsKey []byte) bool {
	if s.IsValidatorCalled != nil {
		return s.IsValidatorCalled(blsKey)
	}
	return false
}

// UseGas -
func (s *SystemEIStub) UseGas(gas uint64) error {
	if s.UseGasCalled != nil {
		return s.UseGasCalled(gas)
	}
	return nil
}

// SetGasProvided -
func (s *SystemEIStub) SetGasProvided(_ uint64) {
}

// ExecuteOnDestContext -
func (s *SystemEIStub) ExecuteOnDestContext(
	destination []byte,
	sender []byte,
	value *big.Int,
	input []byte,
) (*vmcommon.VMOutput, error) {
	if s.ExecuteOnDestContextCalled != nil {
		return s.ExecuteOnDestContextCalled(destination, sender, value, input)
	}

	return &vmcommon.VMOutput{}, nil
}

// SetSystemSCContainer -
func (s *SystemEIStub) SetSystemSCContainer(_ vm.SystemSCContainer) error {
	return nil
}

// BlockChainHook -
func (s *SystemEIStub) BlockChainHook() vmcommon.BlockchainHook {
	if s.BlockChainHookCalled != nil {
		return s.BlockChainHookCalled()
	}
	return &BlockChainHookStub{}
}

// CryptoHook -
func (s *SystemEIStub) CryptoHook() vmcommon.CryptoHook {
	if s.CryptoHookCalled != nil {
		return s.CryptoHookCalled()
	}
	return hooks.NewVMCryptoHook()
}

// AddCode -
func (s *SystemEIStub) AddCode(addr []byte, code []byte) {
	if s.AddCodeCalled != nil {
		s.AddCodeCalled(addr, code)
	}
}

// AddTxValueToSmartContract -
func (s *SystemEIStub) AddTxValueToSmartContract(value *big.Int, scAddress []byte) {
	if s.AddTxValueToSmartContractCalled != nil {
		s.AddTxValueToSmartContractCalled(value, scAddress)
	}
}

// SetSCAddress -
func (s *SystemEIStub) SetSCAddress(_ []byte) {
}

// Finish -
func (s *SystemEIStub) Finish(value []byte) {
	if s.FinishCalled != nil {
		s.FinishCalled(value)
	}
}

// Transfer -
func (s *SystemEIStub) Transfer(destination []byte, sender []byte, value *big.Int, input []byte, _ uint64) error {
	if s.TransferCalled != nil {
		return s.TransferCalled(destination, sender, value, input)
	}
	return nil
}

// GetBalance -
func (s *SystemEIStub) GetBalance(addr []byte) *big.Int {
	if s.GetBalanceCalled != nil {
		return s.GetBalanceCalled(addr)
	}
	return big.NewInt(0)
}

// SetStorage -
func (s *SystemEIStub) SetStorage(key []byte, value []byte) {
	if s.SetStorageCalled != nil {
		s.SetStorageCalled(key, value)
	}
}

// AddReturnMessage -
func (s *SystemEIStub) AddReturnMessage(msg string) {
	if s.AddReturnMessageCalled != nil {
		s.AddReturnMessageCalled(msg)
	}
}

// GetStorage -
func (s *SystemEIStub) GetStorage(key []byte) []byte {
	if s.GetStorageCalled != nil {
		return s.GetStorageCalled(key)
	}
	return nil
}

// SelfDestruct -
func (s *SystemEIStub) SelfDestruct(beneficiary []byte) {
	if s.SelfDestructCalled != nil {
		s.SelfDestructCalled(beneficiary)
	}
}

// CreateVMOutput -
func (s *SystemEIStub) CreateVMOutput() *vmcommon.VMOutput {
	if s.CreateVMOutputCalled != nil {
		return s.CreateVMOutputCalled()
	}

	return &vmcommon.VMOutput{}
}

// CleanCache -
func (s *SystemEIStub) CleanCache() {
	if s.CleanCacheCalled != nil {
		s.CleanCacheCalled()
	}
}

// IsInterfaceNil -
func (s *SystemEIStub) IsInterfaceNil() bool {
	return s == nil
}
