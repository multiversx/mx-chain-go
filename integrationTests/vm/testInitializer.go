// +build cgo

package vm

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &marshal.JsonMarshalizer{}
var testHasher = sha256.Sha256{}
var oneShardCoordinator = mock.NewMultiShardsCoordinatorMock(1)
var addrConv, _ = addressConverters.NewPlainAddressConverter(32, "0x")

type accountFactory struct {
}

func (af *accountFactory) CreateAccount(address state.AddressContainer, tracker state.AccountTracker) (state.AccountHandler, error) {
	return state.NewAccount(address, tracker)
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
	persist, _ := memorydb.New()

	unit, _ := storageUnit.NewStorageUnit(cache, persist)
	return unit
}

func CreateInMemoryShardAccountsDB() *state.AccountsDB {
	marsh := &marshal.JsonMarshalizer{}
	store := CreateMemUnit()

	tr, _ := trie.NewTrie(store, marsh, testHasher)
	adb, _ := state.NewAccountsDB(tr, testHasher, marsh, &accountFactory{})

	return adb
}

func CreateAccount(accnts state.AccountsAdapter, pubKey []byte, nonce uint64, balance *big.Int) []byte {
	address, _ := addrConv.CreateAddressFromPublicKeyBytes(pubKey)
	account, _ := accnts.GetAccountWithJournal(address)
	_ = account.(*state.Account).SetNonceWithJournal(nonce)
	_ = account.(*state.Account).SetBalanceWithJournal(balance)

	hashCreated, _ := accnts.Commit()
	return hashCreated
}

func CreateTxProcessorWithOneSCExecutorMockVM(accnts state.AccountsAdapter, opGas uint64) process.TransactionProcessor {
	blockChainHook, _ := hooks.NewVMAccountsDB(accnts, addrConv)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)
	vm.GasForOperation = opGas

	vmContainer := &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return vm, nil
		}}

	argsParser, _ := smartContract.NewAtArgumentParser()
	scProcessor, _ := smartContract.NewSmartContractProcessor(
		vmContainer,
		argsParser,
		testHasher,
		testMarshalizer,
		accnts,
		blockChainHook,
		addrConv,
		oneShardCoordinator,
		&mock.IntermediateTransactionHandlerMock{},
		&mock.UnsignedTxHandlerMock{},
	)

	txTypeHandler, _ := coordinator.NewTxTypeHandler(
		addrConv,
		oneShardCoordinator,
		accnts)

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
	)

	return txProcessor
}

func CreateOneSCExecutorMockVM(accnts state.AccountsAdapter) vmcommon.VMExecutionHandler {
	blockChainHook, _ := hooks.NewVMAccountsDB(accnts, addrConv)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)

	return vm
}

func CreateVMsContainerAndBlockchainHook(accnts state.AccountsAdapter) (process.VirtualMachinesContainer, *hooks.VMAccountsDB) {
	blockChainHook, _ := hooks.NewVMAccountsDB(accnts, addrConv)

	vmFactory, _ := shard.NewVMContainerFactory(accnts, addrConv)
	vmContainer, _ := vmFactory.Create()

	return vmContainer, blockChainHook
}

func CreateTxProcessorWithOneSCExecutorWithVMs(
	accnts state.AccountsAdapter,
) (process.TransactionProcessor, vmcommon.BlockchainHook) {

	vmFactory, _ := shard.NewVMContainerFactory(accnts, addrConv)
	vmContainer, _ := vmFactory.Create()

	argsParser, _ := smartContract.NewAtArgumentParser()
	scProcessor, _ := smartContract.NewSmartContractProcessor(
		vmContainer,
		argsParser,
		testHasher,
		testMarshalizer,
		accnts,
		vmFactory.VMAccountsDB(),
		addrConv,
		oneShardCoordinator,
		&mock.IntermediateTransactionHandlerMock{},
		&mock.UnsignedTxHandlerMock{},
	)

	txTypeHandler, _ := coordinator.NewTxTypeHandler(
		addrConv,
		oneShardCoordinator,
		accnts)

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
	)

	return txProcessor, vmFactory.VMAccountsDB()
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
	destinationRecovShardAccount, ok := destinationRecovAccount.(*state.Account)

	assert.True(t, ok)
	assert.NotNil(t, destinationRecovShardAccount)
	assert.Equal(t, uint64(0), destinationRecovShardAccount.GetNonce())
	assert.Equal(t, requiredBalance, destinationRecovShardAccount.Balance)
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
	_ = CreateAccount(accnts, senderAddressBytes, senderNonce, senderBalance)

	txProcessor, blockchainHook := CreateTxProcessorWithOneSCExecutorWithVMs(accnts)
	assert.NotNil(tb, txProcessor)

	return txProcessor, accnts, blockchainHook
}

func CreatePreparedTxProcessorAndAccountsWithMockedVM(
	t *testing.T,
	vmOpGas uint64,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
) (process.TransactionProcessor, state.AccountsAdapter) {

	accnts := CreateInMemoryShardAccountsDB()
	_ = CreateAccount(accnts, senderAddressBytes, senderNonce, senderBalance)

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
		Data:     txData,
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
	scCodeAndVMType string,
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
) {

	senderAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(senderAddressBytes)
	senderRecovAccount, _ := accnts.GetExistingAccount(senderAddress)
	senderRecovShardAccount := senderRecovAccount.(*state.Account)

	assert.Equal(t, expectedNonce, senderRecovShardAccount.GetNonce())
	assert.Equal(t, expectedBalance, senderRecovShardAccount.Balance)
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
	shardAccnt, _ := accnt.(*state.Account)

	return shardAccnt.Balance
}

func GetIntValueFromSC(accnts state.AccountsAdapter, scAddressBytes []byte, funcName string, args ...[]byte) *big.Int {
	returnedVals := GetBytesValueFromSC(accnts, scAddressBytes, funcName, args...)

	return big.NewInt(0).SetBytes(returnedVals)
}

func GetBytesValueFromSC(accnts state.AccountsAdapter, scAddressBytes []byte, funcName string, args ...[]byte) []byte {
	vmContainer, _ := CreateVMsContainerAndBlockchainHook(accnts)
	scDataGetter, _ := smartContract.NewSCDataGetter(vmContainer)

	returnedVals, _ := scDataGetter.Get(scAddressBytes, funcName, args...)

	return returnedVals
}

func CreateTopUpTx(nonce uint64, value *big.Int, scAddrress []byte, sndAddress []byte) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		Nonce:    nonce,
		Value:    value,
		RcvAddr:  scAddrress,
		SndAddr:  sndAddress,
		GasPrice: 0,
		GasLimit: 5000,
		Data:     "topUp@0",
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
		GasLimit: 5000,
		Data:     "transfer@" + hex.EncodeToString(rcvAddress) + "@" + value.String(),
	}
}
