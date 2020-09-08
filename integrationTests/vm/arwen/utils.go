package arwen

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	processTransaction "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/require"
)

// VMTypeHex -
const VMTypeHex = "0500"

// DummyCodeMetadataHex -
const DummyCodeMetadataHex = "0102"

const maxGasLimit = 100000000000

var marshalizer = &marshal.GogoProtoMarshalizer{}
var hasher = sha256.Sha256{}
var oneShardCoordinator = mock.NewMultiShardsCoordinatorMock(2)
var pkConverter, _ = pubkeyConverter.NewHexPubkeyConverter(32)

// GasSchedulePath --
var GasSchedulePath = "../gasSchedule.toml"

// DNSAddresses --
var DNSAddresses = make(map[string]struct{})

// TestContext -
type TestContext struct {
	T *testing.T

	Round uint64

	Owner        testParticipant
	Alice        testParticipant
	Bob          testParticipant
	Carol        testParticipant
	Participants []*testParticipant

	GasLimit    uint64
	GasSchedule map[string]map[string]uint64

	UnsignexTxHandler process.TransactionFeeHandler
	EconomicsFee      process.FeeHandler
	LastConsumedFee   uint64

	ScAddress        []byte
	ScCodeMetadata   vmcommon.CodeMetadata
	Accounts         *state.AccountsDB
	TxProcessor      process.TransactionProcessor
	ScProcessor      process.SmartContractProcessor
	QueryService     external.SCQueryService
	VMContainer      process.VirtualMachinesContainer
	BlockchainHook   *hooks.BlockChainHookImpl
	RewardsProcessor RewardsProcessor

	LastTxHash    []byte
	SCRForwarder  *mock.IntermediateTransactionHandlerMock
	LastSCResults []*smartContractResult.SmartContractResult
}

type testParticipant struct {
	Nonce           uint64
	Address         []byte
	BalanceSnapshot *big.Int
}

type RewardsProcessor interface {
	ProcessRewardTransaction(rTx *rewardTx.RewardTx) error
}

// AddressHex will return the participant address in hex string format
func (participant *testParticipant) AddressHex() string {
	return hex.EncodeToString(participant.Address)
}

// SetupTestContext -
func SetupTestContext(t *testing.T) *TestContext {
	var err error

	context := &TestContext{}
	context.T = t
	context.Round = 500

	context.initAccounts()

	context.GasSchedule, err = core.LoadGasScheduleConfig(GasSchedulePath)
	require.Nil(t, err)

	context.initFeeHandlers()
	context.initVMAndBlockchainHook()
	context.initTxProcessorWithOneSCExecutorWithVMs()
	context.ScAddress, _ = context.BlockchainHook.NewAddress(context.Owner.Address, context.Owner.Nonce, factory.ArwenVirtualMachine)
	context.QueryService, _ = smartContract.NewSCQueryService(context.VMContainer, context.EconomicsFee)

	context.RewardsProcessor, err = rewardTransaction.NewRewardTxProcessor(context.Accounts, pkConverter, oneShardCoordinator)
	require.Nil(t, err)

	require.NotNil(t, context.TxProcessor)
	require.NotNil(t, context.ScProcessor)
	require.NotNil(t, context.QueryService)
	require.NotNil(t, context.VMContainer)

	return context
}

func (context *TestContext) initFeeHandlers() {
	context.UnsignexTxHandler = &mock.UnsignedTxHandlerMock{
		ProcessTransactionFeeCalled: func(cost *big.Int, hash []byte) {
			context.LastConsumedFee = cost.Uint64()
		},
	}

	context.EconomicsFee = &mock.FeeHandlerStub{
		DeveloperPercentageCalled: func() float64 {
			return 0.0
		},
		MaxGasLimitPerBlockCalled: func() uint64 {
			return maxGasLimit
		},
	}
}

func (context *TestContext) initVMAndBlockchainHook() {
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:          context.GasSchedule,
		MapDNSAddresses: DNSAddresses,
		Marshalizer:     marshalizer,
	}

	builtInFuncs, err := builtInFunctions.CreateBuiltInFunctionContainer(argsBuiltIn)
	require.Nil(context.T, err)
	blockchainMock := &mock.BlockChainMock{}
	chainStorer := &mock.ChainStorerMock{}

	args := hooks.ArgBlockChainHook{
		Accounts:         context.Accounts,
		PubkeyConv:       pkConverter,
		StorageService:   chainStorer,
		BlockChain:       blockchainMock,
		ShardCoordinator: oneShardCoordinator,
		Marshalizer:      marshalizer,
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions: builtInFuncs,
	}

	vmFactoryConfig := config.VirtualMachineConfig{
		OutOfProcessEnabled: false,
		OutOfProcessConfig:  config.VirtualMachineOutOfProcessConfig{MaxLoopTime: 1000},
	}

	vmFactory, err := shard.NewVMContainerFactory(vmFactoryConfig, maxGasLimit, context.GasSchedule, args)
	require.Nil(context.T, err)

	context.VMContainer, err = vmFactory.Create()
	require.Nil(context.T, err)

	context.BlockchainHook = vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
}

func (context *TestContext) initTxProcessorWithOneSCExecutorWithVMs() {
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  pkConverter,
		ShardCoordinator: oneShardCoordinator,
		BuiltInFuncNames: context.BlockchainHook.GetBuiltInFunctions().Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}

	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	require.Nil(context.T, err)

	gasSchedule := make(map[string]map[string]uint64)
	defaults.FillGasMapInternal(gasSchedule, 1)

	context.SCRForwarder = &mock.IntermediateTransactionHandlerMock{}
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:    context.VMContainer,
		ArgsParser:     smartContract.NewArgumentParser(),
		Hasher:         hasher,
		Marshalizer:    marshalizer,
		AccountsDB:     context.Accounts,
		BlockChainHook: context.BlockchainHook,
		PubkeyConv:     pkConverter,
		Coordinator:    oneShardCoordinator,
		ScrForwarder:   context.SCRForwarder,
		BadTxForwarder: &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:   context.UnsignexTxHandler,
		EconomicsFee:   context.EconomicsFee,
		TxTypeHandler:  txTypeHandler,
		GasHandler: &mock.GasHandlerMock{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasSchedule:      gasSchedule,
		BuiltInFunctions: context.BlockchainHook.GetBuiltInFunctions(),
		TxLogsProcessor:  &mock.TxLogsProcessorStub{},
		EpochNotifier:    forking.NewGenericEpochNotifier(),
	}

	context.ScProcessor, err = smartContract.NewSmartContractProcessor(argsNewSCProcessor)
	require.Nil(context.T, err)

	argsNewTxProcessor := processTransaction.ArgsNewTxProcessor{
		Accounts:                       context.Accounts,
		Hasher:                         hasher,
		PubkeyConv:                     pkConverter,
		Marshalizer:                    marshalizer,
		SignMarshalizer:                marshalizer,
		ShardCoordinator:               oneShardCoordinator,
		ScProcessor:                    context.ScProcessor,
		TxFeeHandler:                   context.UnsignexTxHandler,
		TxTypeHandler:                  txTypeHandler,
		EconomicsFee:                   context.EconomicsFee,
		ReceiptForwarder:               &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:                 &mock.IntermediateTransactionHandlerMock{},
		ArgsParser:                     smartContract.NewArgumentParser(),
		ScrForwarder:                   &mock.IntermediateTransactionHandlerMock{},
		RelayedTxEnableEpoch:           0,
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  forking.NewGenericEpochNotifier(),
	}

	context.TxProcessor, err = processTransaction.NewTxProcessor(argsNewTxProcessor)
	require.Nil(context.T, err)
}

// Close closes the test context
func (context *TestContext) Close() {
	_ = context.VMContainer.Close()
}

func (context *TestContext) initAccounts() {
	context.Accounts = vm.CreateInMemoryShardAccountsDB()

	context.Owner = testParticipant{}
	context.Owner.Address, _ = hex.DecodeString("d4105de8e44aee9d4be670401cec546e5df381028e805012386a05acf76518d9")
	context.Owner.Nonce = uint64(1)
	context.Owner.BalanceSnapshot = big.NewInt(math.MaxInt64)

	context.Alice = testParticipant{}
	context.Alice.Address, _ = hex.DecodeString("0f36a982b79d3c1fda9b82a646a2b423cb3e7223cffbae73a4e3d2c1ea62ee5e")
	context.Alice.Nonce = uint64(1)
	context.Alice.BalanceSnapshot = big.NewInt(math.MaxInt64)

	context.Bob = testParticipant{}
	context.Bob.Address, _ = hex.DecodeString("afb051dc3a1dfb029866730243c2cbc51d8b8ef15951e4da3929f9c8391f307a")
	context.Bob.Nonce = uint64(1)
	context.Bob.BalanceSnapshot = big.NewInt(math.MaxInt64)

	context.Carol = testParticipant{}
	context.Carol.Address, _ = hex.DecodeString("5bdf4c81489bea69ba29cd3eea2670c1bb6cb5d922fa8cb6e17bca71dfdd49f0")
	context.Carol.Nonce = uint64(1)
	context.Carol.BalanceSnapshot = big.NewInt(math.MaxInt64)

	context.createAccount(&context.Owner)
	context.createAccount(&context.Alice)
	context.createAccount(&context.Bob)
	context.createAccount(&context.Carol)
}

func (context *TestContext) createAccount(participant *testParticipant) {
	_, err := vm.CreateAccount(context.Accounts, participant.Address, participant.Nonce, participant.BalanceSnapshot)
	require.Nil(context.T, err)
}

func (context *TestContext) InitAdditionalParticipants(num int) {
	context.Participants = make([]*testParticipant, 0, num)

	for i := 0; i < num; i++ {
		participant := &testParticipant{
			Nonce:           1,
			BalanceSnapshot: NewBalance(10).Value,
			Address:         createDummyAddress(i + 42),
		}

		context.Participants = append(context.Participants, participant)
		context.createAccount(participant)
	}
}

func createDummyAddress(addressTag int) []byte {
	address := make([]byte, 32)
	binary.LittleEndian.PutUint64(address, uint64(addressTag))
	binary.LittleEndian.PutUint64(address[24:], uint64(addressTag))
	return address
}

func (context *TestContext) TakeAccountBalanceSnapshot(participant *testParticipant) {
	participant.BalanceSnapshot = context.GetAccountBalance(participant)
}

func (context *TestContext) GetAccountBalance(participant *testParticipant) *big.Int {
	account, err := context.Accounts.GetExistingAccount(participant.Address)
	require.Nil(context.T, err)
	accountAsUser := account.(state.UserAccountHandler)
	return accountAsUser.GetBalance()
}

func (context *TestContext) GetAccountBalanceDelta(participant *testParticipant) *big.Int {
	account, err := context.Accounts.GetExistingAccount(participant.Address)
	require.Nil(context.T, err)
	accountAsUser := account.(state.UserAccountHandler)
	currentBalance := accountAsUser.GetBalance()
	delta := currentBalance.Sub(currentBalance, participant.BalanceSnapshot)
	return delta
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
		tx.GasLimit = maxGasLimit
	}

	txHash, err := core.CalculateHash(marshalizer, hasher, tx)
	if err != nil {
		return err
	}

	context.LastTxHash = txHash

	_, err = context.TxProcessor.ProcessTransaction(tx)
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

	_ = context.UpdateLastSCResults()

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
		GasLimit: context.GasLimit,
		Data:     []byte(txData),
	}

	// Add default gas limit for tests
	if tx.GasLimit == 0 {
		tx.GasLimit = maxGasLimit
	}

	txHash, err := core.CalculateHash(marshalizer, hasher, tx)
	if err != nil {
		return err
	}

	context.LastTxHash = txHash

	_, err = context.TxProcessor.ProcessTransaction(tx)
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

	_ = context.UpdateLastSCResults()

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
	return context.ExecuteSCWithValue(sender, txData, big.NewInt(0))
}

// ExecuteSCWithValue -
func (context *TestContext) ExecuteSCWithValue(sender *testParticipant, txData string, value *big.Int) error {
	tx := &transaction.Transaction{
		Nonce:    sender.Nonce,
		Value:    new(big.Int).Set(value),
		RcvAddr:  context.ScAddress,
		SndAddr:  sender.Address,
		GasPrice: 1,
		GasLimit: context.GasLimit,
		Data:     []byte(txData),
	}

	// Add default gas limit for tests
	if tx.GasLimit == 0 {
		tx.GasLimit = maxGasLimit
	}

	txHash, err := core.CalculateHash(marshalizer, hasher, tx)
	if err != nil {
		return err
	}

	context.LastTxHash = txHash

	_, err = context.TxProcessor.ProcessTransaction(tx)
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

	_ = context.UpdateLastSCResults()

	return nil
}

// UpdateLastSCResults --
func (context *TestContext) UpdateLastSCResults() error {
	transactions := context.SCRForwarder.GetIntermediateTransactions()
	context.LastSCResults = make([]*smartContractResult.SmartContractResult, len(transactions))
	for i, tx := range transactions {
		scrTx, ok := tx.(*smartContractResult.SmartContractResult)
		if ok {
			context.LastSCResults[i] = scrTx
		} else {
			return errors.New("could not convert tx to scr")
		}
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

// QuerySCBigInt -
func (context *TestContext) QuerySCBigInt(function string, args [][]byte) *big.Int {
	bytes := context.querySC(function, args)
	return big.NewInt(0).SetBytes(bytes)
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

// GoToEpoch -
func (context *TestContext) GoToEpoch(epoch int) {
	header := &block.Header{Nonce: uint64(epoch) * 100, Round: uint64(epoch) * 100, Epoch: uint32(epoch)}
	context.BlockchainHook.SetCurrentHeader(header)
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

// Balance -
type Balance struct {
	Value *big.Int
}

// NewBalance
func NewBalance(n int) Balance {
	result := big.NewInt(0)
	_, _ = result.SetString("1000000000000000000", 10)
	result.Mul(result, big.NewInt(int64(n)))
	return Balance{Value: result}
}

// NewBalanceBig
func NewBalanceBig(bi *big.Int) Balance {
	return Balance{Value: bi}
}

// Times -
func (b Balance) Times(n int) Balance {
	result := b.Value.Mul(b.Value, big.NewInt(int64(n)))
	return Balance{Value: result}
}

// ToHex -
func (b Balance) ToHex() string {
	return "00" + hex.EncodeToString(b.Value.Bytes())
}

// AlmostEquals -
func RequireAlmostEquals(t *testing.T, expected Balance, actual Balance) {
	precision := big.NewInt(0)
	_, _ = precision.SetString("100000000000", 10)
	delta := big.NewInt(0)
	delta = delta.Sub(expected.Value, actual.Value)
	delta = delta.Abs(delta)
	require.True(t, delta.Cmp(precision) < 0, fmt.Sprintf("%s != %s", expected, actual))
}
