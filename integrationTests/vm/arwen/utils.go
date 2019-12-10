package arwen

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
	Carol testParticipant

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

	gasSchedule, err := core.LoadGasScheduleConfig("./gasSchedule.toml")
	assert.Nil(t, err)

	vmContainer, blockChainHook := vm.CreateVMAndBlockchainHook(context.Accounts, gasSchedule)
	context.TxProcessor = vm.CreateTxProcessorWithOneSCExecutorWithVMs(context.Accounts, vmContainer, blockChainHook)
	context.ScAddress, _ = blockChainHook.NewAddress(context.Owner.Address, context.Owner.Nonce, factory.ArwenVirtualMachine)
	context.QueryService, _ = smartContract.NewSCQueryService(vmContainer, math.MaxInt32)

	return context
}

func (context *testContext) initAccounts() {
	initialNonce := uint64(1)

	context.Owner = testParticipant{}
	context.Owner.Address, _ = hex.DecodeString("d4105de8e44aee9d4be670401cec546e5df381028e805012386a05acf76518d9")
	context.Owner.Nonce = initialNonce
	context.Owner.Balance = big.NewInt(math.MaxInt64)

	context.Alice = testParticipant{}
	context.Alice.Address, _ = hex.DecodeString("0f36a982b79d3c1fda9b82a646a2b423cb3e7223cffbae73a4e3d2c1ea62ee5e")
	context.Alice.Nonce = initialNonce
	context.Alice.Balance = big.NewInt(math.MaxInt64)

	context.Bob = testParticipant{}
	context.Bob.Address, _ = hex.DecodeString("afb051dc3a1dfb029866730243c2cbc51d8b8ef15951e4da3929f9c8391f307a")
	context.Bob.Nonce = initialNonce
	context.Bob.Balance = big.NewInt(math.MaxInt64)

	context.Carol = testParticipant{}
	context.Carol.Address, _ = hex.DecodeString("5bdf4c81489bea69ba29cd3eea2670c1bb6cb5d922fa8cb6e17bca71dfdd49f0")
	context.Carol.Nonce = initialNonce
	context.Carol.Balance = big.NewInt(math.MaxInt64)

	context.Accounts = vm.CreateInMemoryShardAccountsDB()
	context.createAccount(&context.Owner)
	context.createAccount(&context.Alice)
	context.createAccount(&context.Bob)
	context.createAccount(&context.Carol)
}

func (context *testContext) createAccount(participant *testParticipant) {
	_, err := vm.CreateAccount(context.Accounts, participant.Address, participant.Nonce, participant.Balance)
	if err != nil {
		assert.FailNow(context.T, err.Error())
	}
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

	_, err = context.Accounts.Commit()
	if err != nil {
		assert.FailNow(context.T, err.Error())
	}
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

	_, err = context.Accounts.Commit()
	if err != nil {
		assert.FailNow(context.T, err.Error())
	}
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
