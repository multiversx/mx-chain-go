package systemSmartContracts

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	vmData "github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type vmContext struct {
	blockChainHook      vm.BlockchainHook
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
	blockChainHook vm.BlockchainHook,
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

func (host *vmContext) getCodeFromAddress(address []byte) []byte {
	userAcc, err := host.blockChainHook.GetUserAccount(address)
	if err != nil {
		// backward compatibility
		return address
	}

	code := host.blockChainHook.GetCode(userAcc)
	if len(code) == 0 {
		return address
	}

	return code
}

// GetContract gets the actual system contract from address and code
func (host *vmContext) GetContract(address []byte) (vm.SystemSmartContract, error) {
	code := host.getCodeFromAddress(address)
	contract, err := host.systemContracts.Get(code)
	if err != nil {
		return nil, err
	}

	if !contract.CanUseContract() {
		// backward compatibility
		return nil, vm.ErrUnknownSystemSmartContract
	}

	return contract, nil
}

// GetStorageFromAddress gets the storage from address and key
func (host *vmContext) GetStorageFromAddress(address []byte, key []byte) []byte {
	storageAdrMap, exists := host.storageUpdate[string(address)]
	if exists {
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
	_, exists := host.storageUpdate[strAdr]
	if !exists {
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
	outAcc, exists := host.outputAccounts[strAdr]
	if exists {
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

// SendGlobalSettingToAll handles sending the information to all the shards
func (host *vmContext) SendGlobalSettingToAll(_ []byte, input []byte) {
	outputTransfer := vmcommon.OutputTransfer{
		Value:    big.NewInt(0),
		Data:     input,
		CallType: vmData.DirectCall,
	}

	for i := uint8(0); i < uint8(host.blockChainHook.NumberOfShards()); i++ {
		systemAddress := make([]byte, len(core.SystemAccountAddress))
		copy(systemAddress, core.SystemAccountAddress)
		systemAddress[len(core.SystemAccountAddress)-1] = i

		globalOutAcc, exists := host.outputAccounts[string(systemAddress)]
		if !exists {
			globalOutAcc = &vmcommon.OutputAccount{
				Address:      systemAddress,
				Balance:      big.NewInt(0),
				BalanceDelta: big.NewInt(0),
			}
		}
		globalOutAcc.OutputTransfers = append(globalOutAcc.OutputTransfers, outputTransfer)
		host.outputAccounts[string(systemAddress)] = globalOutAcc
	}
}

// Transfer handles any necessary value transfer required and takes
// the necessary steps to create accounts
func (host *vmContext) Transfer(
	destination []byte,
	sender []byte,
	value *big.Int,
	input []byte,
	gasLimit uint64,
) error {

	senderAcc, exists := host.outputAccounts[string(sender)]
	if !exists {
		senderAcc = &vmcommon.OutputAccount{
			Address:      sender,
			BalanceDelta: big.NewInt(0),
			Balance:      big.NewInt(0),
		}
		host.outputAccounts[string(senderAcc.Address)] = senderAcc
	}

	destAcc, exists := host.outputAccounts[string(destination)]
	if !exists {
		destAcc = &vmcommon.OutputAccount{
			Address:      destination,
			BalanceDelta: big.NewInt(0),
			Balance:      big.NewInt(0),
		}
		host.outputAccounts[string(destAcc.Address)] = destAcc
	}

	_ = senderAcc.BalanceDelta.Sub(senderAcc.BalanceDelta, value)
	_ = destAcc.BalanceDelta.Add(destAcc.BalanceDelta, value)

	outputTransfer := vmcommon.OutputTransfer{
		Value:    big.NewInt(0).Set(value),
		GasLimit: gasLimit,
		Data:     input,
		CallType: vmData.DirectCall,
	}
	destAcc.OutputTransfers = append(destAcc.OutputTransfers, outputTransfer)

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

func (host *vmContext) mergeContext(currContext *vmContext) {
	host.output = append(host.output, currContext.output...)
	host.AddReturnMessage(currContext.returnMessage)

	for key, storageUpdate := range currContext.storageUpdate {
		_, exists := host.storageUpdate[key]
		if !exists {
			host.storageUpdate[key] = storageUpdate
			continue
		}

		for internKey, internStore := range storageUpdate {
			host.storageUpdate[key][internKey] = internStore
		}
	}

	for _, rightAccount := range currContext.outputAccounts {
		leftAccount, exist := host.outputAccounts[string(rightAccount.Address)]
		if !exist {
			leftAccount = &vmcommon.OutputAccount{}
			host.outputAccounts[string(rightAccount.Address)] = leftAccount
		}
		leftAccount.MergeOutputAccounts(rightAccount)
	}
	host.scAddress = currContext.scAddress
}

func (host *vmContext) createContractCallInput(
	destination []byte,
	sender []byte,
	value *big.Int,
	data []byte,
) (*vmcommon.ContractCallInput, error) {
	function, arguments, err := host.inputParser.ParseData(string(data))
	if err != nil {
		return nil, err
	}

	return createDirectCallInput(destination, sender, value, function, arguments), nil
}

func createDirectCallInput(
	destination []byte,
	sender []byte,
	value *big.Int,
	function string,
	arguments [][]byte,
) *vmcommon.ContractCallInput {
	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: sender,
			Arguments:  arguments,
			CallValue:  value,
		},
		RecipientAddr: destination,
		Function:      function,
	}
	return input
}

// DeploySystemSC will deploy a smart contract according to the input
// will call the init function and merge the vmOutputs
// will add to the system smart contracts container the new address
func (host *vmContext) DeploySystemSC(
	baseContract []byte,
	newAddress []byte,
	ownerAddress []byte,
	initFunction string,
	value *big.Int,
	input [][]byte,
) (vmcommon.ReturnCode, error) {

	if check.IfNil(host.systemContracts) {
		return vmcommon.ExecutionFailed, vm.ErrUnknownSystemSmartContract
	}

	callInput := createDirectCallInput(newAddress, ownerAddress, value, initFunction, input)
	err := host.Transfer(callInput.RecipientAddr, host.scAddress, callInput.CallValue, nil, 0)
	if err != nil {
		return vmcommon.ExecutionFailed, err
	}

	contract, err := host.systemContracts.Get(baseContract)
	if err != nil {
		return vmcommon.ExecutionFailed, err
	}

	oldSCAddress := host.scAddress
	host.SetSCAddress(callInput.RecipientAddr)
	returnCode := contract.Execute(callInput)
	host.SetSCAddress(oldSCAddress)
	host.addContractDeployToOutput(newAddress, ownerAddress, baseContract)

	return returnCode, nil
}

func (host *vmContext) addContractDeployToOutput(
	contractAddress []byte,
	ownerAddress []byte,
	code []byte,
) {
	codeMetaData := &vmcommon.CodeMetadata{
		Upgradeable: false,
		Payable:     false,
		Readable:    true,
	}

	outAcc, exists := host.outputAccounts[string(contractAddress)]
	if !exists {
		outAcc = &vmcommon.OutputAccount{
			Address:      contractAddress,
			BalanceDelta: big.NewInt(0),
			Balance:      big.NewInt(0),
		}
		host.outputAccounts[string(outAcc.Address)] = outAcc
	}

	outAcc.CodeMetadata = codeMetaData.ToBytes()
	outAcc.Code = code
	outAcc.CodeDeployerAddress = ownerAddress
	host.outputAccounts[string(outAcc.Address)] = outAcc
}

// ExecuteOnDestContext executes the input data in the destinations context
func (host *vmContext) ExecuteOnDestContext(destination []byte, sender []byte, value *big.Int, input []byte) (*vmcommon.VMOutput, error) {
	if check.IfNil(host.systemContracts) {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	callInput, err := host.createContractCallInput(destination, sender, value, input)
	if err != nil {
		return nil, err
	}

	err = host.Transfer(callInput.RecipientAddr, callInput.CallerAddr, callInput.CallValue, nil, 0)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{}
	currContext := host.copyToNewContext()
	defer func() {
		host.output = make([][]byte, 0)
		host.mergeContext(currContext)
	}()

	host.softCleanCache()
	host.SetSCAddress(callInput.RecipientAddr)

	contract, err := host.GetContract(callInput.RecipientAddr)
	if err != nil {
		return nil, err
	}

	if callInput.Function == core.SCDeployInitFunctionName {
		return &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: "cannot call smart contract init function",
		}, nil
	}

	returnCode := contract.Execute(callInput)

	if returnCode == vmcommon.Ok {
		vmOutput = host.CreateVMOutput()
	} else {
		// all changes must be deleted
		host.outputAccounts = make(map[string]*vmcommon.OutputAccount)
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
func (host *vmContext) BlockChainHook() vm.BlockchainHook {
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

// GasLeft returns the remaining gas
func (host *vmContext) GasLeft() uint64 {
	return host.gasRemaining
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
		_, exists := outAccs[addr]
		if !exists {
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
		_, exists := outAccs[addr]
		if !exists {
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

		// backward compatibility - genesis was done without this
		if host.blockChainHook.CurrentNonce() > 0 {
			if len(outAcc.CodeDeployerAddress) > 0 {
				outAccs[addr].CodeDeployerAddress = outAcc.CodeDeployerAddress
			}

			if len(outAcc.CodeMetadata) > 0 {
				outAccs[addr].CodeMetadata = outAcc.CodeMetadata
			}
		}

		outAccs[addr].OutputTransfers = outAcc.OutputTransfers
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
	newSCAcc, exists := host.outputAccounts[string(address)]
	if !exists {
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
	destAcc, exists := host.outputAccounts[string(scAddress)]
	if !exists {
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
		return false
	}

	validatorAccount, castOk := acc.(state.PeerAccountHandler)
	if !castOk {
		return false
	}

	// TODO: rename GetList from validator account
	isValidator := validatorAccount.GetList() == string(common.EligibleList) ||
		validatorAccount.GetList() == string(common.WaitingList) || validatorAccount.GetList() == string(common.LeavingList)
	return isValidator
}

// StatusFromValidatorStatistics returns the list in which the validator is present
func (host *vmContext) StatusFromValidatorStatistics(blsKey []byte) string {
	acc, err := host.validatorAccountsDB.GetExistingAccount(blsKey)
	if err != nil {
		return string(common.InactiveList)
	}

	validatorAccount, castOk := acc.(state.PeerAccountHandler)
	if !castOk {
		return string(common.InactiveList)
	}

	return validatorAccount.GetList()
}

// CanUnJail returns true if the validator is jailed in the validator statistics
func (host *vmContext) CanUnJail(blsKey []byte) bool {
	acc, err := host.validatorAccountsDB.GetExistingAccount(blsKey)
	if err != nil {
		return false
	}

	validatorAccount, castOk := acc.(state.PeerAccountHandler)
	if !castOk {
		return false
	}

	return validatorAccount.GetList() == string(common.JailedList)
}

// IsBadRating returns true if the validators temp rating is under jailed limit
func (host *vmContext) IsBadRating(blsKey []byte) bool {
	acc, err := host.validatorAccountsDB.GetExistingAccount(blsKey)
	if err != nil {
		return false
	}

	validatorAccount, castOk := acc.(state.PeerAccountHandler)
	if !castOk {
		return false
	}

	minChance := host.chanceComputer.GetChance(0)
	return host.chanceComputer.GetChance(validatorAccount.GetTempRating()) < minChance
}

// CleanStorageUpdates deletes all the storage updates, used especially to delete data which was only read not modified
func (host *vmContext) CleanStorageUpdates() {
	host.storageUpdate = make(map[string]map[string][]byte)
}

// IsInterfaceNil returns if the underlying implementation is nil
func (host *vmContext) IsInterfaceNil() bool {
	return host == nil
}
