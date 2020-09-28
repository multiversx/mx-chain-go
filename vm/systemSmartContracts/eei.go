package systemSmartContracts

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type vmContext struct {
	blockChainHook      vmcommon.BlockchainHook
	cryptoHook          vmcommon.CryptoHook
	validatorAccountsDB state.AccountsAdapter
	systemContracts     vm.SystemSCContainer
	inputParser         vm.ArgumentsParser
	chanceComputer      sharding.ChanceComputer
	scAddress           []byte

	storageUpdate  map[string]map[string][]byte
	outputAccounts map[string]*vmcommon.OutputAccount
	gasRemaining   uint64

	returnMessage string
	output        [][]byte
}

// NewVMContext creates a context where smart contracts can run and write
func NewVMContext(
	blockChainHook vmcommon.BlockchainHook,
	cryptoHook vmcommon.CryptoHook,
	inputParser vm.ArgumentsParser,
	validatorAccountsDB state.AccountsAdapter,
	chanceComputer sharding.ChanceComputer,
) (*vmContext, error) {
	if check.IfNilReflect(blockChainHook) {
		return nil, vm.ErrNilBlockchainHook
	}
	if check.IfNilReflect(cryptoHook) {
		return nil, vm.ErrNilCryptoHook
	}
	if check.IfNil(inputParser) {
		return nil, vm.ErrNilArgumentsParser
	}
	if check.IfNil(validatorAccountsDB) {
		return nil, vm.ErrNilValidatorAccountsDB
	}
	if check.IfNil(chanceComputer) {
		return nil, vm.ErrNilChanceComputer
	}

	vmc := &vmContext{
		blockChainHook:      blockChainHook,
		cryptoHook:          cryptoHook,
		inputParser:         inputParser,
		validatorAccountsDB: validatorAccountsDB,
		chanceComputer:      chanceComputer,
	}
	vmc.CleanCache()

	return vmc, nil
}

// SetSystemSCContainer sets the existing system smart contracts to the vm context
func (host *vmContext) SetSystemSCContainer(scContainer vm.SystemSCContainer) error {
	if check.IfNil(scContainer) {
		return vm.ErrNilSystemContractsContainer
	}

	host.systemContracts = scContainer
	return nil
}

// GetStorageFromAddress gets the storage from address and key
func (host *vmContext) GetStorageFromAddress(address []byte, key []byte) []byte {
	if storageAdrMap, ok := host.storageUpdate[string(address)]; ok {
		if value, isInMap := storageAdrMap[string(key)]; isInMap {
			return value
		}
	}

	data, err := host.blockChainHook.GetStorageData(address, key)
	if err != nil {
		return nil
	}

	return data
}

// GetStorage gets the values saved for a certain key
func (host *vmContext) GetStorage(key []byte) []byte {
	return host.GetStorageFromAddress(host.scAddress, key)
}

// SetStorageForAddress saves the key value storage under the address
func (host *vmContext) SetStorageForAddress(address []byte, key []byte, value []byte) {
	strAdr := string(address)
	if _, ok := host.storageUpdate[strAdr]; !ok {
		host.storageUpdate[strAdr] = make(map[string][]byte)
	}

	length := len(value)
	host.storageUpdate[strAdr][string(key)] = make([]byte, length)
	copy(host.storageUpdate[strAdr][string(key)][:length], value[:length])
}

// SetStorage saves the key value storage under the address
func (host *vmContext) SetStorage(key []byte, value []byte) {
	host.SetStorageForAddress(host.scAddress, key, value)
}

// GetBalance returns the balance of the given address
func (host *vmContext) GetBalance(addr []byte) *big.Int {
	strAdr := string(addr)
	if outAcc, ok := host.outputAccounts[strAdr]; ok {
		actualBalance := big.NewInt(0).Add(outAcc.Balance, outAcc.BalanceDelta)
		return actualBalance
	}

	account, err := host.blockChainHook.GetUserAccount(addr)
	if err == state.ErrAccNotFound {
		return big.NewInt(0)
	}
	if err != nil {
		return nil
	}

	host.outputAccounts[strAdr] = &vmcommon.OutputAccount{
		Balance:      big.NewInt(0).Set(account.GetBalance()),
		BalanceDelta: big.NewInt(0),
		Address:      addr}

	return account.GetBalance()
}

// Transfer handles any necessary value transfer required and takes
// the necessary steps to create accounts
func (host *vmContext) Transfer(destination []byte, sender []byte, value *big.Int, input []byte, gasLimit uint64) error {

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
	destAcc.Data = append(destAcc.Data, input...)
	destAcc.GasLimit += gasLimit

	return nil
}

func (host *vmContext) copyToNewContext() *vmContext {
	newContext := vmContext{
		storageUpdate:  host.storageUpdate,
		outputAccounts: host.outputAccounts,
		output:         host.output,
		scAddress:      host.scAddress,
	}

	return &newContext
}

func (host *vmContext) copyFromContext(currContext *vmContext) {
	host.output = append(host.output, currContext.output...)
	host.AddReturnMessage(currContext.returnMessage)

	for key, storageUpdate := range currContext.storageUpdate {
		if _, ok := host.storageUpdate[key]; !ok {
			host.storageUpdate[key] = storageUpdate
			continue
		}

		for internKey, internStore := range storageUpdate {
			host.storageUpdate[key][internKey] = internStore
		}
	}

	host.outputAccounts = currContext.outputAccounts
	host.scAddress = currContext.scAddress
}

func (host *vmContext) createContractCallInput(destination []byte, sender []byte, value *big.Int, data []byte) (*vmcommon.ContractCallInput, error) {
	function, arguments, err := host.inputParser.ParseData(string(data))
	if err != nil {
		return nil, err
	}

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   arguments,
			CallValue:   value,
			GasPrice:    0,
			GasProvided: 0,
		},
		RecipientAddr: destination,
		Function:      function,
	}

	return input, nil
}

// ExecuteOnDestContext executes the input data in the destinations context
func (host *vmContext) ExecuteOnDestContext(destination []byte, sender []byte, value *big.Int, data []byte) (*vmcommon.VMOutput, error) {
	if check.IfNil(host.systemContracts) {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	input, err := host.createContractCallInput(destination, sender, value, data)
	if err != nil {
		return nil, err
	}

	err = host.Transfer(input.RecipientAddr, input.CallerAddr, input.CallValue, nil, 0)
	if err != nil {
		return nil, err
	}

	currContext := host.copyToNewContext()
	defer func() {
		host.output = make([][]byte, 0)
		host.copyFromContext(currContext)
	}()

	host.softCleanCache()
	host.SetSCAddress(input.RecipientAddr)

	contract, err := host.systemContracts.Get(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	if input.Function == core.SCDeployInitFunctionName {
		return &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: "cannot call smart contract init function",
		}, nil
	}

	returnCode := contract.Execute(input)

	vmOutput := &vmcommon.VMOutput{}
	if returnCode == vmcommon.Ok {
		vmOutput = host.CreateVMOutput()
	}
	vmOutput.ReturnCode = returnCode
	vmOutput.ReturnMessage = host.returnMessage

	return vmOutput, nil
}

// Finish append the value to the final output
func (host *vmContext) Finish(value []byte) {
	host.output = append(host.output, value)
}

// AddReturnMessage will set the return message
func (host *vmContext) AddReturnMessage(message string) {
	if message == "" {
		return
	}

	if host.returnMessage == "" {
		host.returnMessage = message
		return
	}

	host.returnMessage += "@" + message
}

// BlockChainHook returns the blockchain hook
func (host *vmContext) BlockChainHook() vmcommon.BlockchainHook {
	return host.blockChainHook
}

// CryptoHook returns the cryptoHook
func (host *vmContext) CryptoHook() vmcommon.CryptoHook {
	return host.cryptoHook
}

// CleanCache cleans the current vmContext
func (host *vmContext) CleanCache() {
	host.storageUpdate = make(map[string]map[string][]byte)
	host.outputAccounts = make(map[string]*vmcommon.OutputAccount)
	host.output = make([][]byte, 0)
	host.returnMessage = ""
	host.gasRemaining = 0
}

// SetGasProvided sets the provided gas
func (host *vmContext) SetGasProvided(gasProvided uint64) {
	host.gasRemaining = gasProvided
}

// UseGas subs from the provided gas the value to consume
func (host *vmContext) UseGas(gasToConsume uint64) error {
	if host.gasRemaining < gasToConsume {
		return vm.ErrNotEnoughGas
	}
	host.gasRemaining = host.gasRemaining - gasToConsume
	return nil
}

func (host *vmContext) softCleanCache() {
	host.outputAccounts = make(map[string]*vmcommon.OutputAccount)
	host.output = make([][]byte, 0)
	host.returnMessage = ""
}

// CreateVMOutput adapts vm output and all saved data from sc run into VM Output
func (host *vmContext) CreateVMOutput() *vmcommon.VMOutput {
	vmOutput := &vmcommon.VMOutput{}

	outAccs := make(map[string]*vmcommon.OutputAccount)
	for addr, updates := range host.storageUpdate {
		if _, ok := outAccs[addr]; !ok {
			outAccs[addr] = &vmcommon.OutputAccount{
				Address:        []byte(addr),
				StorageUpdates: make(map[string]*vmcommon.StorageUpdate),
			}
		}

		for key, value := range updates {
			storageUpdate := &vmcommon.StorageUpdate{
				Offset: []byte(key),
				Data:   value,
			}

			outAccs[addr].StorageUpdates[key] = storageUpdate
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
		if len(outAcc.Data) > 0 {
			outAccs[addr].Data = outAcc.Data
		}

		outAccs[addr].GasLimit = outAcc.GasLimit
	}

	vmOutput.OutputAccounts = outAccs

	vmOutput.GasRemaining = host.gasRemaining
	vmOutput.GasRefund = big.NewInt(0)

	vmOutput.ReturnMessage = host.returnMessage

	if len(host.output) > 0 {
		vmOutput.ReturnData = append(vmOutput.ReturnData, host.output...)
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

// IsValidator returns true if the validator is in eligible or waiting list
func (host *vmContext) IsValidator(blsKey []byte) bool {
	acc, err := host.validatorAccountsDB.GetExistingAccount(blsKey)
	if err != nil {
		//TODO remove this
		log.Warn("vmContext.IsValidator", "return", false, "err", err)
		return false
	}

	validatorAccount, ok := acc.(state.PeerAccountHandler)
	if !ok {
		//TODO remove this
		log.Warn("vmContext.IsValidator", "return", false, "err", "invalid cast")
		return false
	}

	// TODO: rename GetList from validator account
	isValidator := validatorAccount.GetList() == string(core.EligibleList) ||
		validatorAccount.GetList() == string(core.WaitingList) || validatorAccount.GetList() == string(core.LeavingList)
	//TODO remove this
	log.Warn("vmContext.IsValidator", "list", validatorAccount.GetList(), "isValidator", isValidator)
	return isValidator
}

// CanUnJail returns true if the validator is jailed in the validator statistics
func (host *vmContext) CanUnJail(blsKey []byte) bool {
	acc, err := host.validatorAccountsDB.GetExistingAccount(blsKey)
	if err != nil {
		return false
	}

	validatorAccount, ok := acc.(state.PeerAccountHandler)
	if !ok {
		return false
	}

	return validatorAccount.GetList() == string(core.JailedList)
}

// IsBadRating returns true if the validators temp rating is under jailed limit
func (host *vmContext) IsBadRating(blsKey []byte) bool {
	acc, err := host.validatorAccountsDB.GetExistingAccount(blsKey)
	if err != nil {
		return false
	}

	validatorAccount, ok := acc.(state.PeerAccountHandler)
	if !ok {
		return false
	}

	minChance := host.chanceComputer.GetChance(0)
	return host.chanceComputer.GetChance(validatorAccount.GetTempRating()) < minChance
}

// IsInterfaceNil returns if the underlying implementation is nil
func (host *vmContext) IsInterfaceNil() bool {
	return host == nil
}
