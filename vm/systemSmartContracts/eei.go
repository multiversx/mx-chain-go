package systemSmartContracts

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const transferValueOnly = "transferValueOnly"

type vmContext struct {
	blockChainHook      vm.BlockchainHook
	cryptoHook          vmcommon.CryptoHook
	validatorAccountsDB state.AccountsAdapter
	userAccountsDB      state.AccountsAdapter
	systemContracts     vm.SystemSCContainer
	inputParser         vm.ArgumentsParser
	chanceComputer      nodesCoordinator.ChanceComputer
	shardCoordinator    sharding.Coordinator
	scAddress           []byte

	storageUpdate  map[string]map[string][]byte
	outputAccounts map[string]*vmcommon.OutputAccount
	gasRemaining   uint64

	returnMessage string
	output        [][]byte
	logs          []*vmcommon.LogEntry

	enableEpochsHandler common.EnableEpochsHandler
	crtTransferIndex    uint32
}

// VMContextArgs holds the arguments needed to create a new vmContext
type VMContextArgs struct {
	BlockChainHook      vm.BlockchainHook
	CryptoHook          vmcommon.CryptoHook
	InputParser         vm.ArgumentsParser
	ValidatorAccountsDB state.AccountsAdapter
	UserAccountsDB      state.AccountsAdapter
	ChanceComputer      nodesCoordinator.ChanceComputer
	EnableEpochsHandler common.EnableEpochsHandler
	ShardCoordinator    sharding.Coordinator
}

// NewVMContext creates a context where smart contracts can run and write
func NewVMContext(args VMContextArgs) (*vmContext, error) {
	if check.IfNilReflect(args.BlockChainHook) {
		return nil, vm.ErrNilBlockchainHook
	}
	if check.IfNilReflect(args.CryptoHook) {
		return nil, vm.ErrNilCryptoHook
	}
	if check.IfNil(args.InputParser) {
		return nil, vm.ErrNilArgumentsParser
	}
	if check.IfNil(args.ValidatorAccountsDB) {
		return nil, vm.ErrNilValidatorAccountsDB
	}
	if check.IfNil(args.UserAccountsDB) {
		return nil, vm.ErrNilUserAccountsDB
	}
	if check.IfNil(args.ChanceComputer) {
		return nil, vm.ErrNilChanceComputer
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, vm.ErrNilShardCoordinator
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.MultiClaimOnDelegationFlag,
		common.SetSenderInEeiOutputTransferFlag,
		common.AlwaysMergeContextsInEEIFlag,
	})
	if err != nil {
		return nil, err
	}

	vmc := &vmContext{
		blockChainHook:      args.BlockChainHook,
		cryptoHook:          args.CryptoHook,
		inputParser:         args.InputParser,
		validatorAccountsDB: args.ValidatorAccountsDB,
		userAccountsDB:      args.UserAccountsDB,
		chanceComputer:      args.ChanceComputer,
		enableEpochsHandler: args.EnableEpochsHandler,
		shardCoordinator:    args.ShardCoordinator,
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

	data, _, err := host.blockChainHook.GetStorageData(address, key)
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
	outAcc, exists := host.outputAccounts[string(addr)]
	if exists {
		actualBalance := big.NewInt(0).Add(outAcc.Balance, outAcc.BalanceDelta)
		return actualBalance
	}

	account, err := host.blockChainHook.GetUserAccount(addr)
	if err != nil {
		return big.NewInt(0)
	}

	return account.GetBalance()
}

// SendGlobalSettingToAll handles sending the information to all the shards
func (host *vmContext) SendGlobalSettingToAll(sender []byte, input []byte) error {
	if host.shardCoordinator.SameShard(sender, core.SystemAccountAddress) {
		return host.ProcessBuiltInFunction(core.SystemAccountAddress, sender, big.NewInt(0), input, 0)
	}

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
		outputTransfer.Index = host.NextOutputTransferIndex()
		globalOutAcc.OutputTransfers = append(globalOutAcc.OutputTransfers, outputTransfer)
		host.outputAccounts[string(systemAddress)] = globalOutAcc
	}

	return nil
}

func (host *vmContext) transferValueOnly(
	destination []byte,
	sender []byte,
	value *big.Int,
) {
	senderAcc, destAcc := host.getSenderDestination(sender, destination)

	_ = senderAcc.BalanceDelta.Sub(senderAcc.BalanceDelta, value)
	_ = destAcc.BalanceDelta.Add(destAcc.BalanceDelta, value)
}

func (host *vmContext) getSenderDestination(sender, destination []byte) (*vmcommon.OutputAccount, *vmcommon.OutputAccount) {
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

	return senderAcc, destAcc
}

// Transfer handles any necessary value transfer required and takes
// the necessary steps to create accounts
func (host *vmContext) Transfer(
	destination []byte,
	sender []byte,
	value *big.Int,
	input []byte,
	gasLimit uint64,
) {
	host.transferValueOnly(destination, sender, value)
	senderAcc, destAcc := host.getSenderDestination(sender, destination)
	outputTransfer := vmcommon.OutputTransfer{
		Index:    host.NextOutputTransferIndex(),
		Value:    big.NewInt(0).Set(value),
		GasLimit: gasLimit,
		Data:     input,
		CallType: vmData.DirectCall,
	}

	if host.enableEpochsHandler.IsFlagEnabled(common.SetSenderInEeiOutputTransferFlag) {
		outputTransfer.SenderAddress = senderAcc.Address
	}
	destAcc.OutputTransfers = append(destAcc.OutputTransfers, outputTransfer)
}

// ProcessBuiltInFunction will execute process if sender and destination is same shard/sovereign
func (host *vmContext) ProcessBuiltInFunction(
	destination []byte,
	sender []byte,
	value *big.Int,
	input []byte,
	gasLimit uint64,
) error {
	host.Transfer(destination, sender, value, input, gasLimit)
	if !host.shardCoordinator.SameShard(sender, destination) {
		return nil
	}

	vmInput, err := host.createVMInputForBuiltInFunctionCall(destination, sender, value, input, gasLimit)
	if err != nil {
		return err
	}

	vmOutput, err := host.blockChainHook.ProcessBuiltInFunction(vmInput)
	if err != nil {
		return err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return errors.New(vmOutput.ReturnMessage)
	}

	for _, logEntry := range vmOutput.Logs {
		host.AddLogEntry(logEntry)
	}

	return nil
}

func (host *vmContext) createVMInputForBuiltInFunctionCall(
	destination []byte,
	sender []byte,
	value *big.Int,
	input []byte,
	gasLimit uint64,
) (*vmcommon.ContractCallInput, error) {
	function, arguments, err := host.inputParser.ParseData(string(input))
	if err != nil {
		return nil, err
	}
	if !host.blockChainHook.IsBuiltinFunctionName(function) {
		return nil, err
	}

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   arguments,
			CallValue:   value,
			GasProvided: gasLimit,
			CallType:    vmData.ESDTTransferAndExecute,
		},
		RecipientAddr: destination,
		Function:      function,
	}

	return vmInput, nil
}

// GetLogs returns the logs
func (host *vmContext) GetLogs() []*vmcommon.LogEntry {
	return host.logs
}

// GetTotalSentToUser returns the total sent to the specified address
func (host *vmContext) GetTotalSentToUser(dest []byte) *big.Int {
	destination, exists := host.outputAccounts[string(dest)]
	if !exists {
		return big.NewInt(0)
	}

	return destination.BalanceDelta
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

func (host *vmContext) properMergeContexts(parentContext *vmContext, returnCode vmcommon.ReturnCode) {
	if !host.enableEpochsHandler.IsFlagEnabled(common.MultiClaimOnDelegationFlag) {
		host.mergeContext(parentContext)
		return
	}

	host.scAddress = parentContext.scAddress
	host.AddReturnMessage(parentContext.returnMessage)

	// merge contexts if the return code is OK or the fix flag is activated because it was wrong not to merge them if the call failed
	shouldMergeContexts := returnCode == vmcommon.Ok || host.enableEpochsHandler.IsFlagEnabled(common.AlwaysMergeContextsInEEIFlag)
	if !shouldMergeContexts {
		// backwards compatibility
		return
	}

	host.output = append(host.output, parentContext.output...)
	for _, rightAccount := range parentContext.outputAccounts {
		leftAccount, exist := host.outputAccounts[string(rightAccount.Address)]
		if !exist {
			leftAccount = &vmcommon.OutputAccount{
				Balance:      big.NewInt(0),
				BalanceDelta: big.NewInt(0),
				Address:      rightAccount.Address,
			}
			host.outputAccounts[string(rightAccount.Address)] = leftAccount
		}
		addOutputAccounts(leftAccount, rightAccount)
	}
}

func addOutputAccounts(
	destination *vmcommon.OutputAccount,
	rightAccount *vmcommon.OutputAccount,
) {
	if len(rightAccount.Address) != 0 {
		destination.Address = rightAccount.Address
	}
	if rightAccount.Balance != nil {
		destination.Balance = rightAccount.Balance
	}
	if destination.BalanceDelta == nil {
		destination.BalanceDelta = big.NewInt(0)
	}
	if rightAccount.BalanceDelta != nil {
		destination.BalanceDelta.Add(destination.BalanceDelta, rightAccount.BalanceDelta)
	}
	if len(rightAccount.Code) > 0 {
		destination.Code = rightAccount.Code
	}
	if len(rightAccount.CodeMetadata) > 0 {
		destination.CodeMetadata = rightAccount.CodeMetadata
	}
	if rightAccount.Nonce > destination.Nonce {
		destination.Nonce = rightAccount.Nonce
	}

	destination.GasUsed += rightAccount.GasUsed

	if rightAccount.CodeDeployerAddress != nil {
		destination.CodeDeployerAddress = rightAccount.CodeDeployerAddress
	}

	destination.BytesAddedToStorage += rightAccount.BytesAddedToStorage
	destination.BytesDeletedFromStorage += rightAccount.BytesDeletedFromStorage
	destination.OutputTransfers = append(destination.OutputTransfers, rightAccount.OutputTransfers...)
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

func (host *vmContext) transferBeforeSCToSCExec(callInput *vmcommon.ContractCallInput, sender []byte, callType string) {
	if !host.enableEpochsHandler.IsFlagEnabled(common.MultiClaimOnDelegationFlag) {
		host.Transfer(callInput.RecipientAddr, sender, callInput.CallValue, nil, 0)
		return
	}
	host.transferValueOnly(callInput.RecipientAddr, sender, callInput.CallValue)

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(transferValueOnly),
		Address:    sender,
		Topics:     [][]byte{callInput.CallValue.Bytes(), callInput.RecipientAddr},
		Data:       vmcommon.FormatLogDataForCall(callType, callInput.Function, callInput.Arguments),
	}
	host.AddLogEntry(logEntry)
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

	host.transferBeforeSCToSCExec(callInput, host.scAddress, "DeploySmartContract")

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

	host.transferBeforeSCToSCExec(callInput, sender, "ExecuteOnDestContext")

	vmOutput := &vmcommon.VMOutput{ReturnCode: vmcommon.UserError}
	currContext := host.copyToNewContext()
	defer func() {
		// we need to reset here the output since it was already transferred in the vmOutput (host.CreateVMOutput() function)
		// and we do not want to duplicate them
		host.output = make([][]byte, 0)
		host.properMergeContexts(currContext, vmOutput.ReturnCode)
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
		host.storageUpdate = currContext.storageUpdate
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

// GetReturnMessage will return the accumulated return message
func (host *vmContext) GetReturnMessage() string {
	return host.returnMessage
}

// AddLogEntry will add a log entry
func (host *vmContext) AddLogEntry(entry *vmcommon.LogEntry) {
	host.logs = append(host.logs, entry)
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
	host.logs = make([]*vmcommon.LogEntry, 0)
	host.crtTransferIndex = 1
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

// UpdateCodeDeployerAddress will try to update the owner of an existing account stored in this vmContext instance.
// Errors if the account is not found because that is a programming error
func (host *vmContext) UpdateCodeDeployerAddress(scAddress string, newOwner []byte) error {
	outAcc, found := host.outputAccounts[scAddress]
	if !found {
		return vm.ErrInternalErrorWhileSettingNewOwner
	}

	outAcc.CodeDeployerAddress = newOwner

	return nil
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
				Offset:  []byte(key),
				Data:    value,
				Written: true,
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
	vmOutput.Logs = host.logs

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
func (host *vmContext) AddTxValueToSmartContract(input *vmcommon.ContractCallInput) {
	if host.isInShardSCToSCCall(input) {
		host.transferBeforeSCToSCExec(input, input.CallerAddr, "ExecuteOnDestContext")
		return
	}

	destAcc, exists := host.outputAccounts[string(input.RecipientAddr)]
	if !exists {
		destAcc = &vmcommon.OutputAccount{
			Address:      input.RecipientAddr,
			BalanceDelta: big.NewInt(0),
		}
		host.outputAccounts[string(destAcc.Address)] = destAcc
	}

	destAcc.BalanceDelta = big.NewInt(0).Add(destAcc.BalanceDelta, input.CallValue)
}

func (host *vmContext) isInShardSCToSCCall(input *vmcommon.ContractCallInput) bool {
	return host.shardCoordinator.SameShard(input.CallerAddr, input.RecipientAddr) && core.IsSmartContractAddress(input.CallerAddr)
}

// SetOwnerOperatingOnAccount will set the new owner, operating on the user account directly as the normal flow through
// SC processor is not possible
func (host *vmContext) SetOwnerOperatingOnAccount(newOwner []byte) error {
	scAccount, err := host.userAccountsDB.LoadAccount(host.scAddress)
	if err != nil {
		return err
	}

	scAccountHandler, okCast := scAccount.(state.UserAccountHandler)
	if !okCast {
		return fmt.Errorf("%w, not a user account handler", vm.ErrWrongTypeAssertion)
	}

	if len(newOwner) != len(host.scAddress) {
		return vm.ErrWrongNewOwnerAddress
	}
	scAccountHandler.SetOwnerAddress(newOwner)

	return host.userAccountsDB.SaveAccount(scAccountHandler)
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

// NextOutputTransferIndex returns next available output transfer index
func (host *vmContext) NextOutputTransferIndex() uint32 {
	index := host.crtTransferIndex
	host.crtTransferIndex++
	return index
}

// GetCrtTransferIndex returns the current output transfer index
func (host *vmContext) GetCrtTransferIndex() uint32 {
	return host.crtTransferIndex
}

// IsInterfaceNil returns if the underlying implementation is nil
func (host *vmContext) IsInterfaceNil() bool {
	return host == nil
}
