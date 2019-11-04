package systemSmartContracts

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type vmContext struct {
	blockChainHook vmcommon.BlockchainHook
	cryptoHook     vmcommon.CryptoHook
	scAddress      []byte

	storageUpdate  map[string]map[string][]byte
	outputAccounts map[string]*vmcommon.OutputAccount

	output []byte

	selfDestruct map[string][]byte
}

// NewVMContext creates a context where smart contracts can run and write
func NewVMContext(blockChainHook vmcommon.BlockchainHook, cryptoHook vmcommon.CryptoHook) (*vmContext, error) {
	if blockChainHook == nil {
		return nil, vm.ErrNilBlockchainHook
	}
	if cryptoHook == nil {
		return nil, vm.ErrNilCryptoHook
	}

	vmc := &vmContext{blockChainHook: blockChainHook, cryptoHook: cryptoHook}
	vmc.CleanCache()

	return vmc, nil
}

// SelfDestruct destroy the current smart contract, setting the beneficiary for the liberated storage fees
func (host *vmContext) SelfDestruct(beneficiary []byte) {
	host.selfDestruct[string(host.scAddress)] = beneficiary
}

// GetStorage get the values saved for a certain key
func (host *vmContext) GetStorage(key []byte) []byte {
	strAdr := string(host.scAddress)
	if _, ok := host.storageUpdate[strAdr]; ok {
		if value, ok := host.storageUpdate[strAdr][string(key)]; ok {
			return value
		}
	}

	data, err := host.blockChainHook.GetStorageData(host.scAddress, key)
	if err != nil {
		return nil
	}

	return data
}

// SetStorage saves the key value storage under the address
func (host *vmContext) SetStorage(key []byte, value []byte) {
	strAdr := string(host.scAddress)

	if _, ok := host.storageUpdate[strAdr]; !ok {
		host.storageUpdate[strAdr] = make(map[string][]byte, 0)
	}

	length := len(value)
	host.storageUpdate[strAdr][string(key)] = make([]byte, length)
	copy(host.storageUpdate[strAdr][string(key)][:length], value[:length])
}

// GetBalance returns the balance of the given address
func (host *vmContext) GetBalance(addr []byte) *big.Int {
	strAdr := string(addr)
	if outAcc, ok := host.outputAccounts[strAdr]; ok {
		actualBalance := big.NewInt(0).Add(outAcc.Balance, outAcc.BalanceDelta)
		return actualBalance
	}

	balance, err := host.blockChainHook.GetBalance(addr)
	if err != nil {
		return nil
	}

	host.outputAccounts[strAdr] = &vmcommon.OutputAccount{
		Balance:      big.NewInt(0).Set(balance),
		BalanceDelta: big.NewInt(0),
		Address:      addr}

	return balance
}

// Transfer handles any necessary value transfer required and takes
// the necessary steps to create accounts
func (host *vmContext) Transfer(
	destination []byte,
	sender []byte,
	value *big.Int,
	input []byte,
) error {

	senderAcc, ok := host.outputAccounts[string(sender)]
	if !ok {
		senderAcc = &vmcommon.OutputAccount{
			Address:      sender,
			BalanceDelta: big.NewInt(0),
			Balance:      big.NewInt(0),
		}
		host.outputAccounts[string(senderAcc.Address)] = senderAcc
	}

	destAcc, ok := host.outputAccounts[string(destination)]
	if !ok {
		destAcc = &vmcommon.OutputAccount{
			Address:      destination,
			BalanceDelta: big.NewInt(0),
			Balance:      big.NewInt(0),
		}
		host.outputAccounts[string(destAcc.Address)] = destAcc
	}

	_ = senderAcc.BalanceDelta.Sub(senderAcc.BalanceDelta, value)
	_ = destAcc.BalanceDelta.Add(destAcc.BalanceDelta, value)

	return nil
}

// Finish append the value to the final output
func (host *vmContext) Finish(value []byte) {
	host.output = append(host.output, value...)
}

// CleanCache cleans the current vmContext
func (host *vmContext) CleanCache() {
	host.storageUpdate = make(map[string]map[string][]byte, 0)
	host.selfDestruct = make(map[string][]byte)
	host.outputAccounts = make(map[string]*vmcommon.OutputAccount, 0)
	host.output = make([]byte, 0)
}

// CreateVMOutput adapts vm output and all saved data from sc run into VM Output
func (host *vmContext) CreateVMOutput() *vmcommon.VMOutput {
	vmOutput := &vmcommon.VMOutput{}
	// save storage updates
	outAccs := make(map[string]*vmcommon.OutputAccount, 0)
	for addr, updates := range host.storageUpdate {
		if _, ok := outAccs[addr]; !ok {
			outAccs[addr] = &vmcommon.OutputAccount{Address: []byte(addr)}
		}

		for key, value := range updates {
			storageUpdate := &vmcommon.StorageUpdate{
				Offset: []byte(key),
				Data:   value,
			}

			outAccs[addr].StorageUpdates = append(outAccs[addr].StorageUpdates, storageUpdate)
		}
	}

	// add balances
	for addr, outAcc := range host.outputAccounts {
		if _, ok := outAccs[addr]; !ok {
			outAccs[addr] = &vmcommon.OutputAccount{}
		}

		outAccs[addr].Address = outAcc.Address
		outAccs[addr].BalanceDelta = outAcc.BalanceDelta

		if len(outAcc.Code) > 0 {
			outAccs[addr].Code = outAcc.Code
		}
		if outAcc.Nonce > 0 {
			outAccs[addr].Nonce = outAcc.Nonce
		} else {
			outAccs[addr].Nonce, _ = host.blockChainHook.GetNonce(outAcc.Address)
		}
	}

	// add self destructed contracts
	for addr := range host.selfDestruct {
		vmOutput.DeletedAccounts = append(vmOutput.DeletedAccounts, []byte(addr))
	}

	// save to the output finally
	for _, outAcc := range outAccs {
		vmOutput.OutputAccounts = append(vmOutput.OutputAccounts, outAcc)
	}

	vmOutput.GasRemaining = big.NewInt(0)
	vmOutput.GasRefund = big.NewInt(0)

	if len(host.output) > 0 {
		vmOutput.ReturnData = append(vmOutput.ReturnData, big.NewInt(0).SetBytes(host.output))
	}

	return vmOutput
}

// SetSCAddress sets the smart contract address
func (host *vmContext) SetSCAddress(addr []byte) {
	host.scAddress = addr
}

// AddCode adds the input code to the address
func (host *vmContext) AddCode(address []byte, code []byte) {
	newSCAcc, ok := host.outputAccounts[string(address)]
	if !ok {
		host.outputAccounts[string(address)] = &vmcommon.OutputAccount{
			Address:        address,
			Nonce:          0,
			BalanceDelta:   big.NewInt(0),
			StorageUpdates: nil,
			Code:           code,
		}
	} else {
		newSCAcc.Code = code
	}
}

// AddTxValueToSmartContract adds the input transaction value to the smart contract address
func (host *vmContext) AddTxValueToSmartContract(value *big.Int, scAddress []byte) {
	destAcc, ok := host.outputAccounts[string(scAddress)]
	if !ok {
		destAcc = &vmcommon.OutputAccount{
			Address:      scAddress,
			BalanceDelta: big.NewInt(0),
		}
		host.outputAccounts[string(destAcc.Address)] = destAcc
	}

	destAcc.BalanceDelta = big.NewInt(0).Add(destAcc.BalanceDelta, value)
}

// IsInterfaceNil returns if the underlying implementation is nil
func (host *vmContext) IsInterfaceNil() bool {
	if host == nil {
		return true
	}
	return false
}
