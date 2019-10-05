package systemSmartContracts

import (
	"errors"
	"fmt"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type SCFunctionHandler func(value *big.Int, args []*big.Int)

type vmContext struct {
	blockChainHook vmcommon.BlockchainHook
	cryptoHook     vmcommon.CryptoHook

	storageUpdate  map[string](map[string][]byte)
	outputAccounts map[string]*vmcommon.OutputAccount

	output []byte

	selfDestruct map[string][]byte
}

func NewVMContext(blockChainHook vmcommon.BlockchainHook, cryptoHook vmcommon.CryptoHook) (*vmContext, error) {
	if blockChainHook == nil {
		return nil, errors.New("nil blockchain hook")
	}
	if cryptoHook == nil {
		return nil, errors.New("nil cryptohook")
	}

	return &vmContext{blockChainHook: blockChainHook,
		cryptoHook: cryptoHook}, nil
}

// SelfDestruct destroy the current smart contract, setting the beneficiary for the liberated storage fees
func (host *vmContext) SelfDestruct(addr []byte, beneficiary []byte) {
	host.selfDestruct[string(addr)] = beneficiary
}

// GetStorage
func (host *vmContext) GetStorage(addr []byte, key []byte) []byte {
	strAdr := string(addr)
	if _, ok := host.storageUpdate[strAdr]; ok {
		if value, ok := host.storageUpdate[strAdr][string(key)]; ok {
			return value
		}
	}

	data, err := host.blockChainHook.GetStorageData(addr, key)
	if err != nil {
		fmt.Printf("GetStorage returned with error %s \n", err.Error())
	}

	return data
}

// SetStorage saves the key value storage under the address
func (host *vmContext) SetStorage(addr []byte, key []byte, value []byte) {
	strAdr := string(addr)

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
		fmt.Printf("GetBalance returned with error %s \n", err.Error())
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
func (host *vmContext) Transfer(destination []byte, sender []byte, value *big.Int, input []byte,
) error {

	senderAcc, ok := host.outputAccounts[string(sender)]
	if !ok {
		senderAcc = &vmcommon.OutputAccount{
			Address:      sender,
			BalanceDelta: big.NewInt(0),
		}
		host.outputAccounts[string(senderAcc.Address)] = senderAcc
	}

	destAcc, ok := host.outputAccounts[string(destination)]
	if !ok {
		destAcc = &vmcommon.OutputAccount{
			Address:      destination,
			BalanceDelta: big.NewInt(0),
		}
		host.outputAccounts[string(destAcc.Address)] = destAcc
	}

	senderAcc.BalanceDelta = big.NewInt(0).Sub(senderAcc.BalanceDelta, value)
	destAcc.BalanceDelta = big.NewInt(0).Add(destAcc.BalanceDelta, value)

	return nil
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
		}
	}

	// save to the output finally
	for _, outAcc := range outAccs {
		vmOutput.OutputAccounts = append(vmOutput.OutputAccounts, outAcc)
	}

	vmOutput.GasRemaining = big.NewInt(0)
	vmOutput.GasRefund = big.NewInt(0)

	return vmOutput
}

// IsInterfaceNil returns if the underlying implementation is nil
func (host *vmContext) IsInterfaceNil() bool {
	if host == nil {
		return true
	}
	return false
}
