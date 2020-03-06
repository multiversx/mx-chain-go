// +build cgo

package vm

import (
	"encoding/hex"
	"math"
	"math/big"
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/accounts"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &marshal.JsonMarshalizer{}
var testHasher = sha256.Sha256{}
var oneShardCoordinator = mock.NewMultiShardsCoordinatorMock(2)
var addrConv, _ = addressConverters.NewPlainAddressConverter(32, "0x")

var log = logger.GetOrCreate("integrationtests")

type accountFactory struct {
}

func (af *accountFactory) CreateAccount(address state.AddressContainer) (state.AccountHandler, error) {
	return accounts.NewUserAccount(address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *accountFactory) IsInterfaceNil() bool {
	if af == nil {
		return true
	}
	return false
}

func CreateEmptyAddress() state.AddressContainer {
	buff := make([]byte, testHasher.Size())

	return state.NewAddress(buff)
}

func CreateMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)

	unit, _ := storageUnit.NewStorageUnit(cache, memorydb.New())
	return unit
}

func CreateInMemoryShardAccountsDB() *state.AccountsDB {
	marsh := &marshal.JsonMarshalizer{}
	store := CreateMemUnit()
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, memorydb.New(), marsh)
	trieStorage, _ := trie.NewTrieStorageManager(store, config.DBConfig{}, ewl)

	tr, _ := trie.NewTrie(trieStorage, marsh, testHasher)
	adb, _ := state.NewAccountsDB(tr, testHasher, marsh, &accountFactory{})

	return adb
}

func CreateAccount(accnts state.AccountsAdapter, pubKey []byte, nonce uint64, balance *big.Int) ([]byte, error) {
	address, err := addrConv.CreateAddressFromPublicKeyBytes(pubKey)
	if err != nil {
		return nil, err
	}

	account, err := accnts.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	account.(state.UserAccountHandler).SetNonce(nonce)
	_ = account.(state.UserAccountHandler).AddToBalance(balance)

	err = accnts.SaveAccount(account)
	if err != nil {
		return nil, err
	}

	hashCreated, err := accnts.Commit()
	if err != nil {
		return nil, err
	}

	return hashCreated, nil
}

func CreateTxProcessorWithOneSCExecutorMockVM(accnts state.AccountsAdapter, opGas uint64) process.TransactionProcessor {
	args := hooks.ArgBlockChainHook{
		Accounts:         accnts,
		AddrConv:         addrConv,
		StorageService:   &mock.ChainStorerMock{},
		BlockChain:       &mock.BlockChainMock{},
		ShardCoordinator: oneShardCoordinator,
		Marshalizer:      testMarshalizer,
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)
	vm.GasForOperation = opGas

	vmContainer := &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return vm, nil
		}}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(
		addrConv,
		oneShardCoordinator,
		accnts)

	gasSchedule := make(map[string]map[string]uint64)
	FillGasMapInternal(gasSchedule, 1)

	argsParser := vmcommon.NewAtArgumentParser()
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:  vmContainer,
		ArgsParser:   argsParser,
		Hasher:       testHasher,
		Marshalizer:  testMarshalizer,
		AccountsDB:   accnts,
		TempAccounts: blockChainHook,
		AdrConv:      addrConv,
		Coordinator:  oneShardCoordinator,
		ScrForwarder: &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler: &mock.UnsignedTxHandlerMock{},
		EconomicsFee: &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.0
			},
		},
		TxTypeHandler: txTypeHandler,
		GasHandler: &mock.GasHandlerMock{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasMap: gasSchedule,
	}
	scProcessor, _ := smartContract.NewSmartContractProcessor(argsNewSCProcessor)

	txProcessor, _ := transaction.NewTxProcessor(
		accnts,
		testHasher,
		addrConv,
		testMarshalizer,
		oneShardCoordinator,
		scProcessor,
		&mock.UnsignedTxHandlerMock{},
		txTypeHandler,
		&mock.FeeHandlerStub{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	return txProcessor
}

func CreateOneSCExecutorMockVM(accnts state.AccountsAdapter) vmcommon.VMExecutionHandler {
	args := hooks.ArgBlockChainHook{
		Accounts:         accnts,
		AddrConv:         addrConv,
		StorageService:   &mock.ChainStorerMock{},
		BlockChain:       &mock.BlockChainMock{},
		ShardCoordinator: oneShardCoordinator,
		Marshalizer:      testMarshalizer,
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
	}
	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)

	return vm
}

func createAndAddIeleVM(
	vmContainer process.VirtualMachinesContainer,
	blockChainHook vmcommon.BlockchainHook,
) {
	cryptoHook := hooks.NewVMCryptoHook()
	ieleVM := endpoint.NewElrondIeleVM(factory.IELEVirtualMachine, endpoint.ElrondTestnet, blockChainHook, cryptoHook)
	_ = vmContainer.Add(factory.IELEVirtualMachine, ieleVM)
}

func CreateVMAndBlockchainHook(
	accnts state.AccountsAdapter,
	gasSchedule map[string]map[string]uint64,
) (process.VirtualMachinesContainer, *hooks.BlockChainHookImpl) {
	args := hooks.ArgBlockChainHook{
		Accounts:         accnts,
		AddrConv:         addrConv,
		StorageService:   &mock.ChainStorerMock{},
		BlockChain:       &mock.BlockChainMock{},
		ShardCoordinator: oneShardCoordinator,
		Marshalizer:      testMarshalizer,
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
	}

	//Uncomment this to enable trace printing of the vm
	//vm.SetTracePretty()

	maxGasLimitPerBlock := uint64(0xFFFFFFFFFFFFFFFF)

	actualGasSchedule := gasSchedule
	if gasSchedule == nil {
		actualGasSchedule = arwenConfig.MakeGasMap(1)
		FillGasMapInternal(actualGasSchedule, 1)
	}

	vmFactory, err := shard.NewVMContainerFactory(maxGasLimitPerBlock, actualGasSchedule, args)
	if err != nil {
		log.LogIfError(err)
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		log.LogIfError(err)
	}

	blockChainHook, _ := vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
	createAndAddIeleVM(vmContainer, blockChainHook)

	return vmContainer, blockChainHook
}

func CreateTxProcessorWithOneSCExecutorWithVMs(
	accnts state.AccountsAdapter,
	vmContainer process.VirtualMachinesContainer,
	blockChainHook *hooks.BlockChainHookImpl,
) process.TransactionProcessor {
	argsParser := vmcommon.NewAtArgumentParser()
	txTypeHandler, _ := coordinator.NewTxTypeHandler(
		addrConv,
		oneShardCoordinator,
		accnts)

	gasSchedule := make(map[string]map[string]uint64)
	FillGasMapInternal(gasSchedule, 1)
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:  vmContainer,
		ArgsParser:   argsParser,
		Hasher:       testHasher,
		Marshalizer:  testMarshalizer,
		AccountsDB:   accnts,
		TempAccounts: blockChainHook,
		AdrConv:      addrConv,
		Coordinator:  oneShardCoordinator,
		ScrForwarder: &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler: &mock.UnsignedTxHandlerMock{},
		EconomicsFee: &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.0
			},
		},
		TxTypeHandler: txTypeHandler,
		GasHandler: &mock.GasHandlerMock{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasMap: gasSchedule,
	}
	scProcessor, _ := smartContract.NewSmartContractProcessor(argsNewSCProcessor)

	txProcessor, _ := transaction.NewTxProcessor(
		accnts,
		testHasher,
		addrConv,
		testMarshalizer,
		oneShardCoordinator,
		scProcessor,
		&mock.UnsignedTxHandlerMock{},
		txTypeHandler,
		&mock.FeeHandlerStub{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	return txProcessor
}

func TestDeployedContractContents(
	t *testing.T,
	destinationAddressBytes []byte,
	accnts state.AccountsAdapter,
	requiredBalance *big.Int,
	scCode string,
	dataValues map[string]*big.Int,
) {

	scCodeBytes, _ := hex.DecodeString(scCode)
	destinationAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(destinationAddressBytes)
	destinationRecovAccount, _ := accnts.GetExistingAccount(destinationAddress)
	destinationRecovShardAccount, ok := destinationRecovAccount.(state.UserAccountHandler)

	assert.True(t, ok)
	assert.NotNil(t, destinationRecovShardAccount)
	assert.Equal(t, uint64(0), destinationRecovShardAccount.GetNonce())
	assert.Equal(t, requiredBalance, destinationRecovShardAccount.GetBalance())
	//test codehash
	assert.Equal(t, testHasher.Compute(string(scCodeBytes)), destinationRecovAccount.GetCodeHash())
	//test code
	assert.Equal(t, scCodeBytes, destinationRecovAccount.GetCode())
	//in this test we know we have a as a variable inside the contract, we can ask directly its value
	// using trackableDataTrie functionality
	assert.NotNil(t, destinationRecovShardAccount.GetRootHash())

	for variable, requiredVal := range dataValues {
		contractVariableData, err := destinationRecovShardAccount.DataTrieTracker().RetrieveValue([]byte(variable))
		assert.Nil(t, err)
		assert.NotNil(t, contractVariableData)

		contractVariableValue := big.NewInt(0).SetBytes(contractVariableData)
		assert.Equal(t, requiredVal, contractVariableValue)
	}
}

func AccountExists(accnts state.AccountsAdapter, addressBytes []byte) bool {
	address, _ := addrConv.CreateAddressFromPublicKeyBytes(addressBytes)
	accnt, _ := accnts.GetExistingAccount(address)

	return accnt != nil
}

func CreatePreparedTxProcessorAndAccountsWithVMs(
	tb testing.TB,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
) (process.TransactionProcessor, state.AccountsAdapter, vmcommon.BlockchainHook) {

	accnts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accnts, senderAddressBytes, senderNonce, senderBalance)

	vmContainer, blockChainHook := CreateVMAndBlockchainHook(accnts, nil)

	txProcessor := CreateTxProcessorWithOneSCExecutorWithVMs(accnts, vmContainer, blockChainHook)
	assert.NotNil(tb, txProcessor)

	return txProcessor, accnts, blockChainHook
}

func CreateTxProcessorArwenVMWithGasSchedule(
	tb testing.TB,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	gasSchedule map[string]map[string]uint64,
) (process.TransactionProcessor, state.AccountsAdapter, vmcommon.BlockchainHook) {

	accnts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accnts, senderAddressBytes, senderNonce, senderBalance)

	vmContainer, blockChainHook := CreateVMAndBlockchainHook(accnts, gasSchedule)
	txProcessor := CreateTxProcessorWithOneSCExecutorWithVMs(accnts, vmContainer, blockChainHook)
	assert.NotNil(tb, txProcessor)

	return txProcessor, accnts, blockChainHook
}

func CreatePreparedTxProcessorAndAccountsWithMockedVM(
	t *testing.T,
	vmOpGas uint64,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
) (process.TransactionProcessor, state.AccountsAdapter) {

	accnts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accnts, senderAddressBytes, senderNonce, senderBalance)

	txProcessor := CreateTxProcessorWithOneSCExecutorMockVM(accnts, vmOpGas)
	assert.NotNil(t, txProcessor)

	return txProcessor, accnts
}

func CreateTx(
	tb testing.TB,
	senderAddressBytes []byte,
	receiverAddressBytes []byte,
	senderNonce uint64,
	value *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	scCodeOrFunc string,
) *dataTransaction.Transaction {

	txData := scCodeOrFunc
	tx := &dataTransaction.Transaction{
		Nonce:    senderNonce,
		Value:    value,
		SndAddr:  senderAddressBytes,
		RcvAddr:  receiverAddressBytes,
		Data:     []byte(txData),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}
	assert.NotNil(tb, tx)

	return tx
}

func CreateDeployTx(
	senderAddressBytes []byte,
	senderNonce uint64,
	value *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	scCodeAndVMType []byte,
) *dataTransaction.Transaction {

	return &dataTransaction.Transaction{
		Nonce:    senderNonce,
		Value:    value,
		SndAddr:  senderAddressBytes,
		RcvAddr:  CreateEmptyAddress().Bytes(),
		Data:     scCodeAndVMType,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}
}

func TestAccount(
	t *testing.T,
	accnts state.AccountsAdapter,
	senderAddressBytes []byte,
	expectedNonce uint64,
	expectedBalance *big.Int,
) *big.Int {

	senderAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(senderAddressBytes)
	senderRecovAccount, _ := accnts.GetExistingAccount(senderAddress)
	senderRecovShardAccount := senderRecovAccount.(state.UserAccountHandler)

	assert.Equal(t, expectedNonce, senderRecovShardAccount.GetNonce())
	assert.Equal(t, expectedBalance, senderRecovShardAccount.GetBalance())
	return senderRecovShardAccount.GetBalance()
}

func ComputeExpectedBalance(
	existing *big.Int,
	transferred *big.Int,
	gasLimit uint64,
	gasPrice uint64,
) *big.Int {

	expectedSenderBalance := big.NewInt(0).Sub(existing, transferred)
	gasFunds := big.NewInt(0).Mul(big.NewInt(0).SetUint64(gasLimit), big.NewInt(0).SetUint64(gasPrice))
	expectedSenderBalance.Sub(expectedSenderBalance, gasFunds)

	return expectedSenderBalance
}

func GetAccountsBalance(addrBytes []byte, accnts state.AccountsAdapter) *big.Int {
	address, _ := addrConv.CreateAddressFromPublicKeyBytes(addrBytes)
	accnt, _ := accnts.GetExistingAccount(address)
	shardAccnt, _ := accnt.(state.UserAccountHandler)

	return shardAccnt.GetBalance()
}

func GetIntValueFromSC(gasSchedule map[string]map[string]uint64, accnts state.AccountsAdapter, scAddressBytes []byte, funcName string, args ...[]byte) *big.Int {
	vmContainer, _ := CreateVMAndBlockchainHook(accnts, gasSchedule)
	scQueryService, _ := smartContract.NewSCQueryService(vmContainer, uint64(math.MaxUint64))

	vmOutput, _ := scQueryService.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddressBytes,
		FuncName:  funcName,
		Arguments: args,
	})

	return big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
}

func CreateTopUpTx(nonce uint64, value *big.Int, scAddrress []byte, sndAddress []byte) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		Nonce:    nonce,
		Value:    value,
		RcvAddr:  scAddrress,
		SndAddr:  sndAddress,
		GasPrice: 0,
		GasLimit: 5000000,
		Data:     []byte("topUp@00"),
	}
}

func CreateTransferTx(
	nonce uint64,
	value *big.Int,
	scAddrress []byte,
	sndAddress []byte,
	rcvAddress []byte,
) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		Nonce:    nonce,
		Value:    big.NewInt(0),
		RcvAddr:  scAddrress,
		SndAddr:  sndAddress,
		GasPrice: 0,
		GasLimit: 5000000,
		Data:     []byte("transferToken@" + hex.EncodeToString(rcvAddress) + "@" + hex.EncodeToString(value.Bytes())),
	}
}

func CreateTransferTokenTx(
	nonce uint64,
	value *big.Int,
	scAddrress []byte,
	sndAddress []byte,
	rcvAddress []byte,
) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		Nonce:    nonce,
		Value:    big.NewInt(0),
		RcvAddr:  scAddrress,
		SndAddr:  sndAddress,
		GasPrice: 0,
		GasLimit: 5000000,
		Data:     []byte("transferToken@" + hex.EncodeToString(rcvAddress) + "@" + hex.EncodeToString(value.Bytes())),
	}
}

func CreateMoveBalanceTx(
	nonce uint64,
	value *big.Int,
	sndAddress []byte,
	rcvAddress []byte,
	gasLimit uint64,
) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		Nonce:    nonce,
		Value:    big.NewInt(0).Set(value),
		RcvAddr:  rcvAddress,
		SndAddr:  sndAddress,
		GasPrice: 1,
		GasLimit: gasLimit,
	}
}

func FillGasMapInternal(gasMap map[string]map[string]uint64, value uint64) map[string]map[string]uint64 {
	gasMap[core.BaseOperationCost] = FillGasMapBaseOperationCosts(value)
	gasMap[core.BuiltInCost] = FillGasMapBuiltInCosts(value)
	gasMap[core.MetaChainSystemSCsCost] = FillGasMapMetaChainSystemSCsCosts(value)

	return gasMap
}

func FillGasMapBaseOperationCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["StorePerByte"] = value
	gasMap["DataCopyPerByte"] = value
	gasMap["ReleasePerByte"] = value
	gasMap["PersistPerByte"] = value
	gasMap["CompilePerByte"] = value

	return gasMap
}

func FillGasMapBuiltInCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["ChangeOwnerAddress"] = value

	return gasMap
}

func FillGasMapMetaChainSystemSCsCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["Stake"] = value
	gasMap["UnStake"] = value
	gasMap["UnBond"] = value
	gasMap["Claim"] = value
	gasMap["Get"] = value
	gasMap["ChangeRewardAddress"] = value
	gasMap["ChangeValidatorKeys"] = value
	gasMap["UnJail"] = value

	return gasMap
}
