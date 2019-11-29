package sctests

import (
	"encoding/hex"
	"io/ioutil"
	"math"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/stretchr/testify/assert"
)

type testContext struct {
	T *testing.T

	Round uint64

	Owner testParticipant
	Alice testParticipant
	Bob   testParticipant

	ScAddress    []byte
	Accounts     *state.AccountsDB
	TxProcessor  process.TransactionProcessor
	QueryService external.SCQueryService
}

type testParticipant struct {
	Nonce   uint64
	Address []byte
	Balance *big.Int
}

func (participant *testParticipant) AddressHex() string {
	return hex.EncodeToString(participant.Address)
}

func setupTestContext(t *testing.T) testContext {
	context := testContext{}
	context.T = t
	context.Round = 500

	context.initAccounts()

	gasSchedule, err := core.LoadGasScheduleConfig("./../gasSchedule.toml")
	assert.Nil(t, err)

	vmContainer, blockChainHook := vm.CreateVMAndBlockchainHook(context.Accounts, gasSchedule)
	context.TxProcessor = vm.CreateTxProcessorWithOneSCExecutorWithVMs(context.Accounts, vmContainer, blockChainHook)
	context.ScAddress, _ = blockChainHook.NewAddress(context.Owner.Address, context.Owner.Nonce, factory.ArwenVirtualMachine)
	context.QueryService, _ = smartContract.NewSCQueryService(vmContainer, math.MaxInt32)
	return context
}

func (context *testContext) initAccounts() {
	context.Owner = testParticipant{}
	context.Owner.Address = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o', 'o'}
	context.Owner.Nonce = uint64(1)
	context.Owner.Balance = big.NewInt(math.MaxInt64)

	context.Alice = testParticipant{}
	context.Alice.Address = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a'}
	context.Alice.Nonce = uint64(1)
	context.Alice.Balance = big.NewInt(math.MaxInt64)

	context.Bob = testParticipant{}
	context.Bob.Address = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b', 'b'}
	context.Bob.Nonce = uint64(1)
	context.Bob.Balance = big.NewInt(math.MaxInt64)

	context.Accounts = vm.CreateInMemoryShardAccountsDB()
	context.createAccount(&context.Owner)
	context.createAccount(&context.Alice)
	context.createAccount(&context.Bob)
}

func (context *testContext) createAccount(participant *testParticipant) {
	_ = vm.CreateAccount(context.Accounts, participant.Address, participant.Nonce, participant.Balance)
}

func (context *testContext) deploySC(wasmPath string, parametersString string) {
	smartContractCode := getSCCode(wasmPath)
	owner := &context.Owner

	txData := smartContractCode + "@" + hex.EncodeToString(factory.ArwenVirtualMachine)
	if parametersString != "" {
		txData = txData + "@" + parametersString
	}

	tx := &transaction.Transaction{
		Nonce:    owner.Nonce,
		Value:    big.NewInt(0),
		RcvAddr:  vm.CreateEmptyAddress().Bytes(),
		SndAddr:  owner.Address,
		GasPrice: 1,
		GasLimit: math.MaxInt32,
		Data:     txData,
	}

	err := context.TxProcessor.ProcessTransaction(tx, context.Round)
	if err != nil {
		assert.FailNow(context.T, err.Error())
	}

	owner.Nonce++
	context.Accounts.Commit()
}

func getSCCode(fileName string) string {
	code, _ := ioutil.ReadFile(fileName)
	codeEncoded := hex.EncodeToString(code)
	return codeEncoded
}

func (context *testContext) executeSC(sender *testParticipant, txData string) {
	context.executeSCWithValue(sender, txData, big.NewInt(0))
}

func (context *testContext) executeSCWithValue(sender *testParticipant, txData string, value *big.Int) {
	tx := &transaction.Transaction{
		Nonce:    sender.Nonce,
		Value:    value,
		RcvAddr:  context.ScAddress,
		SndAddr:  sender.Address,
		GasPrice: 1,
		GasLimit: math.MaxInt32,
		Data:     txData,
	}

	err := context.TxProcessor.ProcessTransaction(tx, context.Round)
	if err != nil {
		assert.FailNow(context.T, err.Error())
	}

	sender.Nonce++
	context.Accounts.Commit()
}

func (context *testContext) querySCInt(function string, args [][]byte) uint64 {
	bytes := context.querySC(function, args)
	result := big.NewInt(0).SetBytes(bytes).Uint64()
	return result
}

func (context *testContext) querySC(function string, args [][]byte) []byte {
	query := process.SCQuery{
		ScAddress: context.ScAddress,
		FuncName:  function,
		Arguments: args,
	}

	vmOutput, err := context.QueryService.ExecuteQuery(&query)
	if err != nil {
		assert.FailNow(context.T, err.Error())
	}

	firstResult := vmOutput.ReturnData[0]
	return firstResult
}

func formatHexNumber(number uint64) string {
	bytes := big.NewInt(0).SetUint64(number).Bytes()
	str := hex.EncodeToString(bytes)
	return str
}
