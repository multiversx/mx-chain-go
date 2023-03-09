package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// SystemEIStub -
type SystemEIStub struct {
	TransferCalled                      func(destination []byte, sender []byte, value *big.Int, input []byte) error
	GetBalanceCalled                    func(addr []byte) *big.Int
	SetStorageCalled                    func(key []byte, value []byte)
	AddReturnMessageCalled              func(msg string)
	GetStorageCalled                    func(key []byte) []byte
	SelfDestructCalled                  func(beneficiary []byte)
	CreateVMOutputCalled                func() *vmcommon.VMOutput
	CleanCacheCalled                    func()
	FinishCalled                        func(value []byte)
	AddCodeCalled                       func(addr []byte, code []byte)
	AddTxValueToSmartContractCalled     func(value *big.Int, scAddress []byte)
	BlockChainHookCalled                func() vm.BlockchainHook
	CryptoHookCalled                    func() vmcommon.CryptoHook
	UseGasCalled                        func(gas uint64) error
	IsValidatorCalled                   func(blsKey []byte) bool
	StatusFromValidatorStatisticsCalled func(blsKey []byte) string
	ExecuteOnDestContextCalled          func(destination, sender []byte, value *big.Int, input []byte) (*vmcommon.VMOutput, error)
	DeploySystemSCCalled                func(baseContract []byte, newAddress []byte, caller []byte, initFunction string, value *big.Int, args [][]byte) (vmcommon.ReturnCode, error)
	GetStorageFromAddressCalled         func(address []byte, key []byte) []byte
	SetStorageForAddressCalled          func(address []byte, key []byte, value []byte)
	CanUnJailCalled                     func(blsKey []byte) bool
	IsBadRatingCalled                   func(blsKey []byte) bool
	SendGlobalSettingToAllCalled        func(sender []byte, input []byte)
	GetContractCalled                   func(address []byte) (vm.SystemSmartContract, error)
	GasLeftCalled                       func() uint64
	CleanStorageUpdatesCalled           func()
	ReturnMessage                       string
	AddLogEntryCalled                   func(entry *vmcommon.LogEntry)
}

// AddLogEntry -
func (s *SystemEIStub) AddLogEntry(entry *vmcommon.LogEntry) {
	if s.AddLogEntryCalled != nil {
		s.AddLogEntryCalled(entry)
	}
}

// GasLeft -
func (s *SystemEIStub) GasLeft() uint64 {
	if s.GasLeftCalled != nil {
		return s.GasLeftCalled()
	}

	return 0
}

// GetContract -
func (s *SystemEIStub) GetContract(address []byte) (vm.SystemSmartContract, error) {
	if s.GetContractCalled != nil {
		return s.GetContractCalled(address)
	}
	return &SystemSCStub{}, nil
}

// CanUnJail -
func (s *SystemEIStub) CanUnJail(blsKey []byte) bool {
	if s.CanUnJailCalled != nil {
		return s.CanUnJailCalled(blsKey)
	}
	return false
}

// IsBadRating -
func (s *SystemEIStub) IsBadRating(blsKey []byte) bool {
	if s.IsBadRatingCalled != nil {
		return s.IsBadRatingCalled(blsKey)
	}
	return false
}

// IsValidator -
func (s *SystemEIStub) IsValidator(blsKey []byte) bool {
	if s.IsValidatorCalled != nil {
		return s.IsValidatorCalled(blsKey)
	}
	return false
}

// StatusFromValidatorStatistics -
func (s *SystemEIStub) StatusFromValidatorStatistics(blsKey []byte) string {
	if s.StatusFromValidatorStatisticsCalled != nil {
		return s.StatusFromValidatorStatisticsCalled(blsKey)
	}
	return ""
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

// DeploySystemSC -
func (s *SystemEIStub) DeploySystemSC(
	baseContract []byte,
	newAddress []byte,
	ownerAddress []byte,
	initFunction string,
	value *big.Int,
	input [][]byte,
) (vmcommon.ReturnCode, error) {
	if s.DeploySystemSCCalled != nil {
		return s.DeploySystemSCCalled(baseContract, newAddress, ownerAddress, initFunction, value, input)
	}
	return vmcommon.Ok, nil
}

// SetSystemSCContainer -
func (s *SystemEIStub) SetSystemSCContainer(_ vm.SystemSCContainer) error {
	return nil
}

// BlockChainHook -
func (s *SystemEIStub) BlockChainHook() vm.BlockchainHook {
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

// SendGlobalSettingToAll -
func (s *SystemEIStub) SendGlobalSettingToAll(sender []byte, input []byte) {
	if s.SendGlobalSettingToAllCalled != nil {
		s.SendGlobalSettingToAllCalled(sender, input)
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
	} else {
		s.ReturnMessage = msg
	}
}

// GetReturnMessage -
func (s *SystemEIStub) GetReturnMessage() string {
	return s.ReturnMessage
}

// GetStorage -
func (s *SystemEIStub) GetStorage(key []byte) []byte {
	if s.GetStorageCalled != nil {
		return s.GetStorageCalled(key)
	}
	return nil
}

// GetStorageFromAddress -
func (s *SystemEIStub) GetStorageFromAddress(address []byte, key []byte) []byte {
	if s.GetStorageFromAddressCalled != nil {
		return s.GetStorageFromAddressCalled(address, key)
	}
	return nil
}

// SetStorageForAddress -
func (s *SystemEIStub) SetStorageForAddress(address []byte, key []byte, value []byte) {
	if s.SetStorageForAddressCalled != nil {
		s.SetStorageForAddressCalled(address, key, value)
	}
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

// CleanStorageUpdates -
func (s *SystemEIStub) CleanStorageUpdates() {
	if s.CleanStorageUpdatesCalled != nil {
		s.CleanStorageUpdatesCalled()
	}
}

// IsInterfaceNil -
func (s *SystemEIStub) IsInterfaceNil() bool {
	return s == nil
}
