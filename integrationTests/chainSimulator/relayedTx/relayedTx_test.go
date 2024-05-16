package relayedTx

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig              = "../../../cmd/node/config/"
	minGasPrice                             = 1_000_000_000
	minGasLimit                             = 50_000
	txVersion                               = 2
	mockTxSignature                         = "sig"
	maxNumOfBlocksToGenerateWhenExecutingTx = 10
)

var oneEGLD = big.NewInt(1000000000000000000)

func TestRelayedTransactionInMultiShardEnvironmentWithChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t)
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(30000))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	receiver, err := cs.GenerateAndMintWalletAddress(1, big.NewInt(0))
	require.NoError(t, err)

	innerTx := generateTransaction(sender.Bytes, 0, receiver.Bytes, oneEGLD, "", minGasLimit)
	innerTx.RelayerAddr = relayer.Bytes

	sender2, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	receiver2, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	innerTx2 := generateTransaction(sender2.Bytes, 0, receiver2.Bytes, oneEGLD, "", minGasLimit)
	innerTx2.RelayerAddr = relayer.Bytes

	pkConv := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()

	// innerTx3Failure should fail due to less gas limit
	// deploy a wrapper contract
	owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	scCode := wasm.GetSCCode("testData/egld-esdt-swap.wasm")
	params := []string{scCode, wasm.VMTypeHex, wasm.DummyCodeMetadataHex, hex.EncodeToString([]byte("WEGLD"))}
	txDataDeploy := strings.Join(params, "@")
	deployTx := generateTransaction(owner.Bytes, 0, make([]byte, 32), big.NewInt(0), txDataDeploy, 600000000)

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	scAddress := result.Logs.Events[0].Address
	scAddressBytes, _ := pkConv.Decode(scAddress)

	// try a wrap transaction which will fail as the contract is paused
	txDataWrap := "wrapEgld"
	gasLimit := 2300000
	innerTx3Failure := generateTransaction(owner.Bytes, 1, scAddressBytes, big.NewInt(1), txDataWrap, uint64(gasLimit))
	innerTx3Failure.RelayerAddr = relayer.Bytes

	innerTx3 := generateTransaction(sender.Bytes, 1, receiver2.Bytes, oneEGLD, "", minGasLimit)
	innerTx3.RelayerAddr = relayer.Bytes

	innerTxs := []*transaction.Transaction{innerTx, innerTx2, innerTx3Failure, innerTx3}

	// relayer will consume first a move balance for each inner tx, then the specific gas for each inner tx
	relayedTxGasLimit := uint64(minGasLimit)
	for _, tx := range innerTxs {
		relayedTxGasLimit += minGasLimit + tx.GasLimit
	}
	relayedTx := generateTransaction(relayer.Bytes, 0, relayer.Bytes, big.NewInt(0), "", relayedTxGasLimit)
	relayedTx.InnerTransactions = innerTxs

	result, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	// generate few more blocks for the cross shard scrs to be done
	err = cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	relayerAccount, err := cs.GetAccount(relayer)
	require.NoError(t, err)
	economicsData := cs.GetNodeHandler(0).GetCoreComponents().EconomicsData()
	relayerMoveBalanceFee := economicsData.ComputeMoveBalanceFee(relayedTx)
	expectedRelayerFee := big.NewInt(0).Mul(relayerMoveBalanceFee, big.NewInt(int64(len(relayedTx.InnerTransactions))))
	for _, tx := range innerTxs {
		expectedRelayerFee.Add(expectedRelayerFee, economicsData.ComputeTxFee(tx))
	}
	assert.Equal(t, big.NewInt(0).Sub(initialBalance, expectedRelayerFee).String(), relayerAccount.Balance)

	senderAccount, err := cs.GetAccount(sender)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0).Sub(initialBalance, big.NewInt(0).Mul(oneEGLD, big.NewInt(2))).String(), senderAccount.Balance)

	sender2Account, err := cs.GetAccount(sender2)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0).Sub(initialBalance, oneEGLD).String(), sender2Account.Balance)

	receiverAccount, err := cs.GetAccount(receiver)
	require.NoError(t, err)
	assert.Equal(t, oneEGLD.String(), receiverAccount.Balance)

	receiver2Account, err := cs.GetAccount(receiver2)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0).Mul(oneEGLD, big.NewInt(2)).String(), receiver2Account.Balance)

	// check SCRs
	shardC := cs.GetNodeHandler(0).GetShardCoordinator()
	for _, scr := range result.SmartContractResults {
		checkSCRStatus(t, cs, pkConv, shardC, scr)
	}

	// check log events
	require.Equal(t, 3, len(result.Logs.Events))
	require.True(t, strings.Contains(string(result.Logs.Events[2].Data), "contract is paused"))
}

func TestRelayedTransactionInMultiShardEnvironmentWithChainSimulatorScCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t)
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	pkConv := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()
	shardC := cs.GetNodeHandler(0).GetShardCoordinator()

	// deploy adder contract
	owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	ownerNonce := uint64(0)
	scCode := wasm.GetSCCode("testData/adder.wasm")
	params := []string{scCode, wasm.VMTypeHex, wasm.DummyCodeMetadataHex, "00"}
	txDataDeploy := strings.Join(params, "@")
	deployTx := generateTransaction(owner.Bytes, ownerNonce, make([]byte, 32), big.NewInt(0), txDataDeploy, 100000000)

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	scAddress := result.Logs.Events[0].Address
	scAddressBytes, _ := pkConv.Decode(scAddress)
	scShard := shardC.ComputeId(scAddressBytes)
	scShardNodeHandler := cs.GetNodeHandler(scShard)

	// 1st inner tx, successful add 1
	ownerNonce++
	txDataAdd := "add@" + hex.EncodeToString(big.NewInt(1).Bytes())
	innerTx1 := generateTransaction(owner.Bytes, ownerNonce, scAddressBytes, big.NewInt(0), txDataAdd, 5000000)
	innerTx1.RelayerAddr = relayer.Bytes

	// 2nd inner tx, successful add 1
	ownerNonce++
	innerTx2 := generateTransaction(owner.Bytes, ownerNonce, scAddressBytes, big.NewInt(0), txDataAdd, 5000000)
	innerTx2.RelayerAddr = relayer.Bytes

	// 3rd inner tx, wrong number of parameters
	ownerNonce++
	innerTx3 := generateTransaction(owner.Bytes, ownerNonce, scAddressBytes, big.NewInt(0), "add", 5000000)
	innerTx3.RelayerAddr = relayer.Bytes

	// 4th inner tx, successful add 1
	ownerNonce++
	innerTx4 := generateTransaction(owner.Bytes, ownerNonce, scAddressBytes, big.NewInt(0), txDataAdd, 5000000)
	innerTx4.RelayerAddr = relayer.Bytes

	// 5th inner tx, invalid function
	ownerNonce++
	innerTx5 := generateTransaction(owner.Bytes, ownerNonce, scAddressBytes, big.NewInt(0), "substract", 5000000)
	innerTx5.RelayerAddr = relayer.Bytes

	// 6th inner tx, successful add 1
	ownerNonce++
	innerTx6 := generateTransaction(owner.Bytes, ownerNonce, scAddressBytes, big.NewInt(0), txDataAdd, 5000000)
	innerTx6.RelayerAddr = relayer.Bytes

	// 7th inner tx, not enough gas
	ownerNonce++
	innerTx7 := generateTransaction(owner.Bytes, ownerNonce, scAddressBytes, big.NewInt(0), txDataAdd, 100000)
	innerTx7.RelayerAddr = relayer.Bytes

	innerTxs := []*transaction.Transaction{innerTx1, innerTx2, innerTx3, innerTx4, innerTx5, innerTx6, innerTx7}

	relayedTxGasLimit := uint64(minGasLimit)
	for _, tx := range innerTxs {
		relayedTxGasLimit += minGasLimit + tx.GasLimit
	}
	relayedTx := generateTransaction(relayer.Bytes, 0, relayer.Bytes, big.NewInt(0), "", relayedTxGasLimit)
	relayedTx.InnerTransactions = innerTxs

	result, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	checkSum(t, scShardNodeHandler, scAddressBytes, owner.Bytes, 4)

	// 8 scrs, 4 from the succeeded txs + 4 with refunded gas to relayer
	require.Equal(t, 8, len(result.SmartContractResults))
	for _, scr := range result.SmartContractResults {
		if strings.Contains(scr.ReturnMessage, "gas refund for relayer") {
			continue
		}

		checkSCRStatus(t, cs, pkConv, shardC, scr)
	}

	// 6 scrs, 3 with signalError + 3 with the actual errors
	require.Equal(t, 6, len(result.Logs.Events))
	expectedLogEvents := map[int]string{
		1: "[wrong number of arguments]",
		3: "[invalid function (not found)] [substract]",
		5: "[not enough gas] [add]",
	}
	for idx, logEvent := range result.Logs.Events {
		if logEvent.Identifier == "signalError" {
			continue
		}

		expectedLogEvent := expectedLogEvents[idx]
		require.True(t, strings.Contains(string(logEvent.Data), expectedLogEvent))
	}
}

func startChainSimulator(t *testing.T) testsChainSimulator.ChainSimulator {
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   false,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              3,
		GenesisTimestamp:         time.Now().Unix(),
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  3,
		NumNodesWaitingListShard: 3,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = 1
			cfg.EpochConfig.EnableEpochs.FixRelayedMoveBalanceEnableEpoch = 1
		},
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	return cs
}

func generateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(data),
		GasLimit:  gasLimit,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   txVersion,
		Signature: []byte(mockTxSignature),
	}
}

func checkSum(
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

func checkSCRStatus(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	pkConv core.PubkeyConverter,
	shardC sharding.Coordinator,
	scr *transaction.ApiSmartContractResult,
) {
	addr, err := pkConv.Decode(scr.RcvAddr)
	require.NoError(t, err)

	senderShard := shardC.ComputeId(addr)
	tx, err := cs.GetNodeHandler(senderShard).GetFacadeHandler().GetTransaction(scr.Hash, true)
	require.NoError(t, err)
	assert.Equal(t, transaction.TxStatusSuccess, tx.Status)
}
