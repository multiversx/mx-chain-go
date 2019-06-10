package mockVM

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/smartContract"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
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

func createDummyAddress() state.AddressContainer {
	buff := make([]byte, sha256.Sha256{}.Size())
	_, _ = rand.Reader.Read(buff)

	return state.NewAddress(buff)
}

func createEmptyAddress() state.AddressContainer {
	buff := make([]byte, sha256.Sha256{}.Size())

	return state.NewAddress(buff)
}

func createMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storageUnit.NewStorageUnit(cache, persist)
	return unit
}

func createInMemoryShardAccountsDB() *state.AccountsDB {
	marsh := &marshal.JsonMarshalizer{}

	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marsh, &accountFactory{})

	return adb
}

func createAccount(accnts state.AccountsAdapter, pubKey []byte, nonce uint64, balance *big.Int) []byte {
	address, _ := addrConv.CreateAddressFromPublicKeyBytes(pubKey)
	account, _ := accnts.GetAccountWithJournal(address)
	_ = account.(*state.Account).SetNonceWithJournal(nonce)
	_ = account.(*state.Account).SetBalanceWithJournal(balance)

	hashCreated, _ := accnts.Commit()
	return hashCreated
}

func createTxProcessorWithOneSCExecutorMockVM(accnts state.AccountsAdapter, opGas uint64) process.TransactionProcessor {
	blockChainHook, _ := hooks.NewVMAccountsDB(accnts, addrConv)
	cryptoHook := &hooks.VMCryptoHook{}
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, cryptoHook)
	vm.GasForOperation = opGas
	argsParser, _ := smartContract.NewAtArgumentParser()
	scProcessor, _ := smartContract.NewSmartContractProcessor(
		vm,
		argsParser,
		testHasher,
		testMarshalizer,
		accnts,
		blockChainHook,
		addrConv,
		oneShardCoordinator,
	)
	txProcessor, _ := transaction.NewTxProcessor(accnts, testHasher, addrConv, testMarshalizer, oneShardCoordinator, scProcessor)

	return txProcessor
}

func testDeployedContractContents(
	t *testing.T,
	destinationAddressBytes []byte,
	accnts state.AccountsAdapter,
	requiredBalance *big.Int,
	scCode string,
	dataValues map[string]*big.Int,
) {

	destinationAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(destinationAddressBytes)
	destinationRecovAccount, _ := accnts.GetExistingAccount(destinationAddress)
	destinationRecovShardAccount, ok := destinationRecovAccount.(*state.Account)

	assert.True(t, ok)
	assert.NotNil(t, destinationRecovShardAccount)
	assert.Equal(t, uint64(0), destinationRecovShardAccount.GetNonce())
	assert.Equal(t, requiredBalance, destinationRecovShardAccount.Balance)
	//test codehash
	assert.Equal(t, testHasher.Compute(scCode), destinationRecovAccount.GetCodeHash())
	//test code
	assert.Equal(t, []byte(scCode), destinationRecovAccount.GetCode())
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

func computeSCDestinationAddressBytes(senderNonce uint64, senderAddressBytes []byte) []byte {
	//TODO change this when receipts are implemented, take the newly created account address (SC account)
	// from receipt, do not recompute here
	senderNonceBytes := big.NewInt(0).SetUint64(senderNonce).Bytes()
	return keccak.Keccak{}.Compute(string(append(senderAddressBytes, senderNonceBytes...)))
}
