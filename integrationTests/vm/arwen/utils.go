package arwen

import (
	"encoding/hex"
	"io/ioutil"
	"math"
	"math/big"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

// VMTypeHex -
const VMTypeHex = "0500"

// DummyCodeMetadataHex -
const DummyCodeMetadataHex = "0102"

// TestContext -
type TestContext struct {
	T *testing.T

	Round uint64

	Owner testParticipant
	Alice testParticipant
	Bob   testParticipant
	Carol testParticipant

	GasLimit uint64

	ScAddress      []byte
	ScCodeMetadata vmcommon.CodeMetadata
	Accounts       *state.AccountsDB
	TxProcessor    process.TransactionProcessor
	ScProcessor    process.SmartContractProcessor
	QueryService   external.SCQueryService
	VMContainer    process.VirtualMachinesContainer
}

type testParticipant struct {
	Nonce   uint64
	Address []byte
	Balance *big.Int
}

// AddressHex will return the participant address in hex string format
func (participant *testParticipant) AddressHex() string {
	return hex.EncodeToString(participant.Address)
}

// SetupTestContext -
func SetupTestContext(t *testing.T) TestContext {
	context := TestContext{}
	context.T = t
	context.Round = 500

	context.initAccounts()

	gasSchedule, err := core.LoadGasScheduleConfig("../gasSchedule.toml")
	require.Nil(t, err)

	vmContainer, blockChainHook := vm.CreateVMAndBlockchainHook(context.Accounts, gasSchedule, false)
	context.TxProcessor, context.ScProcessor = vm.CreateTxProcessorWithOneSCExecutorWithVMs(context.Accounts, vmContainer, blockChainHook)
	context.ScAddress, _ = blockChainHook.NewAddress(context.Owner.Address, context.Owner.Nonce, factory.ArwenVirtualMachine)
	context.QueryService, _ = smartContract.NewSCQueryService(vmContainer, &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	})
	context.VMContainer = vmContainer

	require.NotNil(t, context.TxProcessor)
	require.NotNil(t, context.ScProcessor)
	require.NotNil(t, context.QueryService)
	require.NotNil(t, context.VMContainer)

	return context
}

// Close closes the test context
func (context *TestContext) Close() {
	_ = context.VMContainer.Close()
}

func (context *TestContext) initAccounts() {
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

func (context *TestContext) createAccount(participant *testParticipant) {
	_, err := vm.CreateAccount(context.Accounts, participant.Address, participant.Nonce, participant.Balance)
	require.Nil(context.T, err)
}

// DeploySC -
func (context *TestContext) DeploySC(wasmPath string, parametersString string) error {
	scCode := GetSCCode(wasmPath)
	owner := &context.Owner

	codeMetadataHex := hex.EncodeToString(context.ScCodeMetadata.ToBytes())
	txData := strings.Join([]string{scCode, VMTypeHex, codeMetadataHex}, "@")
	if parametersString != "" {
		txData = txData + "@" + parametersString
	}

	tx := &transaction.Transaction{
		Nonce:    owner.Nonce,
		Value:    big.NewInt(0),
		RcvAddr:  vm.CreateEmptyAddress(),
		SndAddr:  owner.Address,
		GasPrice: 1,
		GasLimit: context.GasLimit,
		Data:     []byte(txData),
	}

	// Add default gas limit for tests
	if tx.GasLimit == 0 {
		tx.GasLimit = math.MaxUint32
	}

	_, err := context.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return err
	}

	owner.Nonce++
	_, err = context.Accounts.Commit()
	if err != nil {
		return err
	}

	err = context.GetLatestError()
	if err != nil {
		return err
	}

	return nil
}

// UpgradeSC -
func (context *TestContext) UpgradeSC(wasmPath string, parametersString string) error {
	scCode := GetSCCode(wasmPath)
	owner := &context.Owner

	codeMetadataHex := hex.EncodeToString(context.ScCodeMetadata.ToBytes())
	txData := strings.Join([]string{"upgradeContract", scCode, codeMetadataHex}, "@")
	if parametersString != "" {
		txData = txData + "@" + parametersString
	}

	tx := &transaction.Transaction{
		Nonce:    owner.Nonce,
		Value:    big.NewInt(0),
		RcvAddr:  context.ScAddress,
		SndAddr:  owner.Address,
		GasPrice: 1,
		GasLimit: math.MaxInt32,
		Data:     []byte(txData),
	}

	_, err := context.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return err
	}

	owner.Nonce++
	_, err = context.Accounts.Commit()
	if err != nil {
		return err
	}

	err = context.GetLatestError()
	if err != nil {
		return err
	}

	return nil
}

// GetSCCode -
func GetSCCode(fileName string) string {
	code, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		panic("Could not get SC code.")
	}

	codeEncoded := hex.EncodeToString(code)
	return codeEncoded
}

// CreateDeployTxData -
func CreateDeployTxData(scCode string) string {
	return strings.Join([]string{scCode, VMTypeHex, DummyCodeMetadataHex}, "@")
}

// ExecuteSC -
func (context *TestContext) ExecuteSC(sender *testParticipant, txData string) error {
	return context.executeSCWithValue(sender, txData, big.NewInt(0))
}

func (context *TestContext) executeSCWithValue(sender *testParticipant, txData string, value *big.Int) error {
	tx := &transaction.Transaction{
		Nonce:    sender.Nonce,
		Value:    new(big.Int).Set(value),
		RcvAddr:  context.ScAddress,
		SndAddr:  sender.Address,
		GasPrice: 1,
		GasLimit: math.MaxInt32,
		Data:     []byte(txData),
	}

	_, err := context.TxProcessor.ProcessTransaction(tx)
	if err != nil {
		return err
	}

	sender.Nonce++
	_, err = context.Accounts.Commit()
	if err != nil {
		return err
	}

	err = context.GetLatestError()
	if err != nil {
		return err
	}

	return nil
}

// QuerySCInt -
func (context *TestContext) QuerySCInt(function string, args [][]byte) uint64 {
	bytes := context.querySC(function, args)
	result := big.NewInt(0).SetBytes(bytes).Uint64()

	return result
}

// QuerySCString -
func (context *TestContext) QuerySCString(function string, args [][]byte) string {
	bytes := context.querySC(function, args)
	return string(bytes)
}

// QuerySCBytes -
func (context *TestContext) QuerySCBytes(function string, args [][]byte) []byte {
	bytes := context.querySC(function, args)
	return bytes
}

func (context *TestContext) querySC(function string, args [][]byte) []byte {
	query := process.SCQuery{
		ScAddress: context.ScAddress,
		FuncName:  function,
		Arguments: args,
	}

	vmOutput, err := context.QueryService.ExecuteQuery(&query)
	require.Nil(context.T, err)

	firstResult := vmOutput.ReturnData[0]
	return firstResult
}

// GetLatestError -
func (context *TestContext) GetLatestError() error {
	return smartContract.GetLatestTestError(context.ScProcessor)
}

// FormatHexNumber -
func FormatHexNumber(number uint64) string {
	bytes := big.NewInt(0).SetUint64(number).Bytes()
	str := hex.EncodeToString(bytes)

	return str
}
