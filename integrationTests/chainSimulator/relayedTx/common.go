package relayedTx

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/require"
)

const (
	DefaultPathToInitialConfig              = "../../../../cmd/node/config/"
	MinGasPrice                             = 1_000_000_000
	MinGasLimit                             = 50_000
	GasPerDataByte                          = 1_500
	DeductionFactor                         = 100
	TxVersion                               = 2
	MockTxSignature                         = "ssig"
	MockRelayerTxSignature                  = "rsig"
	MaxNumOfBlocksToGenerateWhenExecutingTx = 10
	RoundsPerEpoch                          = 40
	GuardAccountCost                        = 250_000
	ExtraGasLimitForGuarded                 = MinGasLimit
	ExtraGasESDTTransfer                    = 250000
	ExtraGasMultiESDTTransfer               = 500000
	EgldTicker                              = "EGLD-000000"
)

var (
	OneEGLD = big.NewInt(1000000000000000000)
)

func StartChainSimulator(
	t *testing.T,
	alterConfigsFunction func(cfg *config.Configs),
) testsChainSimulator.ChainSimulator {
	roundDurationInMillis := uint64(6000)
	roundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    RoundsPerEpoch,
	}
	supernovaRoundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    RoundsPerEpoch * 10,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            DefaultPathToInitialConfig,
		NumOfShards:                    3,
		RoundDurationInMillis:          roundDurationInMillis,
		SupernovaRoundDurationInMillis: roundDurationInMillis / 10,
		RoundsPerEpoch:                 roundsPerEpochOpt,
		SupernovaRoundsPerEpoch:        supernovaRoundsPerEpochOpt,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListMeta:        3,
		NumNodesWaitingListShard:       3,
		AlterConfigsFunction:           alterConfigsFunction,
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	return cs
}

func GenerateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(data),
		GasLimit:  gasLimit,
		GasPrice:  MinGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   TxVersion,
		Signature: []byte(MockTxSignature),
	}
}

func GetBalance(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	address dtos.WalletAddress,
) *big.Int {
	account, err := cs.GetAccount(address)
	require.NoError(t, err)

	balance, ok := big.NewInt(0).SetString(account.Balance, 10)
	require.True(t, ok)

	return balance
}

func GetESDTBalance(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	address dtos.WalletAddress,
	addressShard uint32,
	ticker string,
) *big.Int {
	tokenInfo, _, err := cs.GetNodeHandler(addressShard).GetFacadeHandler().GetESDTData(address.Bech32, ticker, 0, apiData.AccountQueryOptions{})
	require.NoError(t, err)

	return tokenInfo.Value
}

func GetNonce(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	address dtos.WalletAddress,
) uint64 {
	account, err := cs.GetAccount(address)
	require.NoError(t, err)

	return account.Nonce
}

func GenerateRelayedV3Transaction(sender []byte, nonce uint64, receiver []byte, relayer []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	tx := GenerateTransaction(sender, nonce, receiver, value, data, gasLimit)
	tx.RelayerSignature = []byte(MockRelayerTxSignature)
	tx.RelayerAddr = relayer
	return tx
}

func DeployAdder(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	owner dtos.WalletAddress,
	ownerNonce uint64,
) (*transaction.ApiTransactionResult, []byte) {
	pkConv := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()

	err := cs.GenerateBlocks(1)
	require.Nil(t, err)

	scCode := wasm.GetSCCode("../testData/adder.wasm")
	params := []string{scCode, wasm.VMTypeHex, wasm.DummyCodeMetadataHex, "00"}
	txDataDeploy := strings.Join(params, "@")
	deployTx := GenerateTransaction(owner.Bytes, ownerNonce, make([]byte, 32), big.NewInt(0), txDataDeploy, 100000000)

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, MaxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	scAddress := result.Logs.Events[0].Address
	scAddressBytes, _ := pkConv.Decode(scAddress)

	return result, scAddressBytes
}

func CheckSum(
	t *testing.T,
	nodeHandler chainSimulatorProcess.NodeHandler,
	scAddress []byte,
	callerAddress []byte,
	expectedSum int,
) {
	scQuery := &process.SCQuery{
		ScAddress:  scAddress,
		FuncName:   "getSum",
		CallerAddr: callerAddress,
		CallValue:  big.NewInt(0),
	}
	result, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "ok", result.ReturnCode)

	sum, err := strconv.Atoi(hex.EncodeToString(result.ReturnData[0]))
	require.NoError(t, err)

	require.Equal(t, expectedSum, sum)
}

func GetRefundValue(scrs []*transaction.ApiSmartContractResult) *big.Int {
	for _, scr := range scrs {
		if scr.IsRefund {
			return scr.Value
		}
	}

	return big.NewInt(0)
}

func ComputeTxGasAndFeeBasedOnRefund(
	result *transaction.ApiTransactionResult,
	refund *big.Int,
	isMoveBalance bool,
	guardedTx bool,
) (*big.Int, *big.Int, uint64) {
	deductedGasPrice := uint64(MinGasPrice / DeductionFactor)

	initialTx := result.Tx
	gasForFullPrice := uint64(MinGasLimit + GasPerDataByte*len(initialTx.GetData()))
	if guardedTx {
		gasForFullPrice += ExtraGasLimitForGuarded
	}

	if common.IsRelayedTxV3(initialTx) {
		gasForFullPrice += uint64(MinGasLimit) // relayer fee
	}
	gasForDeductedPrice := initialTx.GetGasLimit() - gasForFullPrice

	initialFee := gasForFullPrice*MinGasPrice + gasForDeductedPrice*deductedGasPrice
	finalFee := initialFee - refund.Uint64()

	gasRefunded := refund.Uint64() / deductedGasPrice
	gasConsumed := gasForFullPrice + gasForDeductedPrice - gasRefunded

	if isMoveBalance {
		return big.NewInt(0).SetUint64(initialFee), big.NewInt(0).SetUint64(gasForFullPrice * MinGasPrice), gasForFullPrice
	}

	return big.NewInt(0).SetUint64(initialFee), big.NewInt(0).SetUint64(finalFee), gasConsumed
}

func CheckSCRSucceeded(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	scr *transaction.ApiSmartContractResult,
) {
	pkConv := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()
	shardC := cs.GetNodeHandler(0).GetShardCoordinator()
	addr, err := pkConv.Decode(scr.RcvAddr)
	require.NoError(t, err)

	senderShard := shardC.ComputeId(addr)
	tx, err := cs.GetNodeHandler(senderShard).GetFacadeHandler().GetTransaction(scr.Hash, true)
	require.NoError(t, err)
	require.Equal(t, transaction.TxStatusSuccess, tx.Status)

	if tx.ReturnMessage == core.GasRefundForRelayerMessage {
		return
	}

	require.GreaterOrEqual(t, len(tx.Logs.Events), 1)
	for _, event := range tx.Logs.Events {
		if event.Identifier == core.WriteLogIdentifier {
			continue
		}

		require.Equal(t, core.CompletedTxEventIdentifier, event.Identifier)
	}
}
