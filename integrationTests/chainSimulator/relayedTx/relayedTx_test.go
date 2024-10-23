package relayedTx

import (
	"encoding/hex"
	"encoding/json"
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
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig              = "../../../cmd/node/config/"
	minGasPrice                             = 1_000_000_000
	minGasLimit                             = 50_000
	guardAccountCost                        = 250_000
	extraGasLimitForGuarded                 = minGasLimit
	gasPerDataByte                          = 1_500
	txVersion                               = 2
	mockTxSignature                         = "sig"
	maxNumOfBlocksToGenerateWhenExecutingTx = 10
	roundsPerEpoch                          = 30
)

var (
	oneEGLD                                  = big.NewInt(1000000000000000000)
	alterConfigsFuncRelayedV3EarlyActivation = func(cfg *config.Configs) {
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = 1
	}
)

func TestRelayedTransactionInMultiShardEnvironmentWithChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t, alterConfigsFuncRelayedV3EarlyActivation)
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(30000))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	receiver, err := cs.GenerateAndMintWalletAddress(1, big.NewInt(0))
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	innerTx := generateTransaction(sender.Bytes, 0, receiver.Bytes, oneEGLD, "", minGasLimit)
	innerTx.RelayerAddr = relayer.Bytes

	sender2, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	receiver2, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	innerTx2 := generateTransaction(sender2.Bytes, 0, receiver2.Bytes, oneEGLD, "", minGasLimit)
	innerTx2.RelayerAddr = relayer.Bytes

	pkConv := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()

	// innerTx3Failure should fail due to less gas limit
	// deploy a wrapper contract
	owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

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
	relayedTxGasLimit := uint64(0)
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

	economicsData := cs.GetNodeHandler(0).GetCoreComponents().EconomicsData()
	relayerMoveBalanceFee := economicsData.ComputeMoveBalanceFee(relayedTx)
	expectedRelayerFee := big.NewInt(0).Mul(relayerMoveBalanceFee, big.NewInt(int64(len(relayedTx.InnerTransactions))))
	for _, tx := range innerTxs {
		expectedRelayerFee.Add(expectedRelayerFee, economicsData.ComputeTxFee(tx))
	}
	checkBalance(t, cs, relayer, big.NewInt(0).Sub(initialBalance, expectedRelayerFee))

	checkBalance(t, cs, sender, big.NewInt(0).Sub(initialBalance, big.NewInt(0).Mul(oneEGLD, big.NewInt(2))))

	checkBalance(t, cs, sender2, big.NewInt(0).Sub(initialBalance, oneEGLD))

	checkBalance(t, cs, receiver, oneEGLD)

	checkBalance(t, cs, receiver2, big.NewInt(0).Mul(oneEGLD, big.NewInt(2)))

	// check SCRs
	shardC := cs.GetNodeHandler(0).GetShardCoordinator()
	for _, scr := range result.SmartContractResults {
		checkSCRSucceeded(t, cs, pkConv, shardC, scr)
	}

	// 3 log events from the failed sc call
	require.Equal(t, 3, len(result.Logs.Events))
	require.True(t, strings.Contains(string(result.Logs.Events[2].Data), "contract is paused"))
}

func TestRelayedTransactionInMultiShardEnvironmentWithChainSimulatorAndInvalidNonces(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t, alterConfigsFuncRelayedV3EarlyActivation)
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(30000))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	receiver, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// bump sender nonce to 3
	tx0 := generateTransaction(sender.Bytes, 0, sender.Bytes, big.NewInt(0), "", minGasLimit)
	tx1 := generateTransaction(sender.Bytes, 1, sender.Bytes, big.NewInt(0), "", minGasLimit)
	tx2 := generateTransaction(sender.Bytes, 2, sender.Bytes, big.NewInt(0), "", minGasLimit)
	_, err = cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx0, tx1, tx2}, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.Nil(t, err)

	// higher nonce
	innerTx1 := generateTransaction(sender.Bytes, 10, receiver.Bytes, oneEGLD, "", minGasLimit)
	innerTx1.RelayerAddr = relayer.Bytes

	// higher nonce
	innerTx2 := generateTransaction(sender.Bytes, 9, receiver.Bytes, oneEGLD, "", minGasLimit)
	innerTx2.RelayerAddr = relayer.Bytes

	// nonce ok
	innerTx3 := generateTransaction(sender.Bytes, 3, receiver.Bytes, oneEGLD, "", minGasLimit)
	innerTx3.RelayerAddr = relayer.Bytes

	// higher nonce
	innerTx4 := generateTransaction(sender.Bytes, 8, receiver.Bytes, oneEGLD, "", minGasLimit)
	innerTx4.RelayerAddr = relayer.Bytes

	// lower nonce - initial one
	innerTx5 := generateTransaction(sender.Bytes, 3, receiver.Bytes, oneEGLD, "", minGasLimit)
	innerTx5.RelayerAddr = relayer.Bytes

	innerTxs := []*transaction.Transaction{innerTx1, innerTx2, innerTx3, innerTx4, innerTx5}

	// relayer will consume first a move balance for each inner tx, then the specific gas for each inner tx
	relayedTxGasLimit := uint64(0)
	for _, tx := range innerTxs {
		relayedTxGasLimit += minGasLimit + tx.GasLimit
	}
	relayedTx := generateTransaction(relayer.Bytes, 0, relayer.Bytes, big.NewInt(0), "", relayedTxGasLimit)
	relayedTx.InnerTransactions = innerTxs

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	// 5 scrs, 4 from the failed txs + 1 with success
	require.Equal(t, 5, len(result.SmartContractResults))
	scrsMap := make(map[string]int, len(result.SmartContractResults))
	for _, scr := range result.SmartContractResults {
		if len(scr.ReturnMessage) == 0 {
			scrsMap["success"]++
		}
		if strings.Contains(scr.ReturnMessage, process.ErrHigherNonceInTransaction.Error()) {
			scrsMap[process.ErrHigherNonceInTransaction.Error()]++
		}
		if strings.Contains(scr.ReturnMessage, process.ErrLowerNonceInTransaction.Error()) {
			scrsMap[process.ErrLowerNonceInTransaction.Error()]++
		}
	}
	require.Equal(t, 1, scrsMap["success"])
	require.Equal(t, 3, scrsMap[process.ErrHigherNonceInTransaction.Error()])
	require.Equal(t, 1, scrsMap[process.ErrLowerNonceInTransaction.Error()])

	// 4 log events from the failed txs
	require.Equal(t, 4, len(result.Logs.Events))
	for _, event := range result.Logs.Events {
		require.Equal(t, core.SignalErrorOperation, event.Identifier)
	}
}

func TestRelayedTransactionInMultiShardEnvironmentWithChainSimulatorScCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t, alterConfigsFuncRelayedV3EarlyActivation)
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	pkConv := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()
	shardC := cs.GetNodeHandler(0).GetShardCoordinator()

	// deploy adder contract
	owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	ownerNonce := uint64(0)
	scAddressBytes := deployAdder(t, cs, owner, ownerNonce)
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

	relayedTxGasLimit := uint64(0)
	for _, tx := range innerTxs {
		relayedTxGasLimit += minGasLimit + tx.GasLimit
	}
	relayedTx := generateTransaction(relayer.Bytes, 0, relayer.Bytes, big.NewInt(0), "", relayedTxGasLimit)
	relayedTx.InnerTransactions = innerTxs

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	checkSum(t, scShardNodeHandler, scAddressBytes, owner.Bytes, 4)

	// 8 scrs, 4 from the succeeded txs + 4 with refunded gas to relayer
	require.Equal(t, 8, len(result.SmartContractResults))
	for _, scr := range result.SmartContractResults {
		if strings.Contains(scr.ReturnMessage, "gas refund for relayer") {
			continue
		}

		checkSCRSucceeded(t, cs, pkConv, shardC, scr)
	}

	// 6 events, 3 with signalError + 3 with the actual errors
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

func TestRelayedTransactionInMultiShardEnvironmentWithChainSimulatorInnerMoveBalanceToNonPayableSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t, func(cfg *config.Configs) {
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.FixRelayedMoveBalanceToNonPayableSCEnableEpoch = 1
	})
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	pkConv := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()

	// deploy adder contract
	owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	ownerNonce := uint64(0)
	scCode := wasm.GetSCCode("testData/adder.wasm")
	params := []string{scCode, wasm.VMTypeHex, "0000", "00"}
	txDataDeploy := strings.Join(params, "@")
	deployTx := generateTransaction(owner.Bytes, ownerNonce, make([]byte, 32), big.NewInt(0), txDataDeploy, 100000000)

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	scAddress := result.Logs.Events[0].Address
	scAddressBytes, _ := pkConv.Decode(scAddress)

	balanceRelayerBefore := getBalance(t, cs, relayer)
	balanceOwnerBefore := getBalance(t, cs, owner)

	// move balance to non-payable contract should only consume fees and sender's nonce
	ownerNonce++
	innerTx1 := generateTransaction(owner.Bytes, ownerNonce, scAddressBytes, oneEGLD, "", 50000)
	innerTx1.RelayerAddr = relayer.Bytes

	// move balance to meta contract should only consume fees and sender's nonce
	ownerNonce++
	innerTx2 := generateTransaction(owner.Bytes, ownerNonce, core.ESDTSCAddress, oneEGLD, "", 50000)
	innerTx2.RelayerAddr = relayer.Bytes

	innerTxs := []*transaction.Transaction{innerTx1, innerTx2}

	relayedTxGasLimit := uint64(0)
	for _, tx := range innerTxs {
		relayedTxGasLimit += minGasLimit + tx.GasLimit
	}
	relayedTx := generateTransaction(relayer.Bytes, 0, relayer.Bytes, big.NewInt(0), "", relayedTxGasLimit)
	relayedTx.InnerTransactions = innerTxs

	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	balanceRelayerAfter := getBalance(t, cs, relayer)
	balanceOwnerAfter := getBalance(t, cs, owner)
	consumedRelayedFee := core.SafeMul(relayedTxGasLimit, minGasPrice)
	expectedBalanceRelayerAfter := big.NewInt(0).Sub(balanceRelayerBefore, consumedRelayedFee)
	require.Equal(t, balanceOwnerBefore.String(), balanceOwnerAfter.String())
	require.Equal(t, expectedBalanceRelayerAfter.String(), balanceRelayerAfter.String())
}

func TestFixRelayedMoveBalanceWithChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	expectedFeeScCallBefore := "815294920000000"
	expectedFeeScCallAfter := "873704920000000"
	t.Run("sc call", testFixRelayedMoveBalanceWithChainSimulatorScCall(expectedFeeScCallBefore, expectedFeeScCallAfter))

	expectedFeeMoveBalanceBefore := "797500000000000" // 498 * 1500 + 50000 + 5000
	expectedFeeMoveBalanceAfter := "847000000000000"  // 498 * 1500 + 50000 + 50000
	t.Run("move balance", testFixRelayedMoveBalanceWithChainSimulatorMoveBalance(expectedFeeMoveBalanceBefore, expectedFeeMoveBalanceAfter))
}

func testFixRelayedMoveBalanceWithChainSimulatorScCall(
	expectedFeeBeforeFix string,
	expectedFeeAfterFix string,
) func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(7)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// deploy adder contract
		owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		ownerNonce := uint64(0)
		scAddressBytes := deployAdder(t, cs, owner, ownerNonce)

		// fast-forward until epoch 4
		err = cs.GenerateBlocksUntilEpochIsReached(int32(4))
		require.NoError(t, err)

		// send relayed tx
		txDataAdd := "add@" + hex.EncodeToString(big.NewInt(1).Bytes())
		innerTx := generateTransaction(owner.Bytes, 1, scAddressBytes, big.NewInt(0), txDataAdd, 3000000)
		marshalledTx, err := json.Marshal(innerTx)
		require.NoError(t, err)
		txData := []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit := 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx := generateTransaction(relayer.Bytes, 0, owner.Bytes, big.NewInt(0), string(txData), gasLimit)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// send relayed tx, fix still not active
		innerTx = generateTransaction(owner.Bytes, 2, scAddressBytes, big.NewInt(0), txDataAdd, 3000000)
		marshalledTx, err = json.Marshal(innerTx)
		require.NoError(t, err)
		txData = []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit = 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx = generateTransaction(relayer.Bytes, 1, owner.Bytes, big.NewInt(0), string(txData), gasLimit)

		relayerBalanceBefore := getBalance(t, cs, relayer)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)
		relayerBalanceAfter := getBalance(t, cs, relayer)

		feeConsumed := big.NewInt(0).Sub(relayerBalanceBefore, relayerBalanceAfter)

		require.Equal(t, expectedFeeBeforeFix, feeConsumed.String())

		// fast-forward until the fix is active
		err = cs.GenerateBlocksUntilEpochIsReached(int32(providedActivationEpoch))
		require.NoError(t, err)

		// send relayed tx after fix
		innerTx = generateTransaction(owner.Bytes, 3, scAddressBytes, big.NewInt(0), txDataAdd, 3000000)
		marshalledTx, err = json.Marshal(innerTx)
		require.NoError(t, err)
		txData = []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit = 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx = generateTransaction(relayer.Bytes, 2, owner.Bytes, big.NewInt(0), string(txData), gasLimit)

		relayerBalanceBefore = getBalance(t, cs, relayer)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		relayerBalanceAfter = getBalance(t, cs, relayer)

		feeConsumed = big.NewInt(0).Sub(relayerBalanceBefore, relayerBalanceAfter)

		require.Equal(t, expectedFeeAfterFix, feeConsumed.String())
	}
}

func testFixRelayedMoveBalanceWithChainSimulatorMoveBalance(
	expectedFeeBeforeFix string,
	expectedFeeAfterFix string,
) func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(5)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		receiver, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		// send relayed tx
		innerTx := generateTransaction(sender.Bytes, 0, receiver.Bytes, oneEGLD, "", 50000)
		marshalledTx, err := json.Marshal(innerTx)
		require.NoError(t, err)
		txData := []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit := 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx := generateTransaction(relayer.Bytes, 0, sender.Bytes, big.NewInt(0), string(txData), gasLimit)

		relayerBalanceBefore := getBalance(t, cs, relayer)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		relayerBalanceAfter := getBalance(t, cs, relayer)

		feeConsumed := big.NewInt(0).Sub(relayerBalanceBefore, relayerBalanceAfter)

		require.Equal(t, expectedFeeBeforeFix, feeConsumed.String())

		// fast-forward until the fix is active
		err = cs.GenerateBlocksUntilEpochIsReached(int32(providedActivationEpoch))
		require.NoError(t, err)

		// send relayed tx
		innerTx = generateTransaction(sender.Bytes, 1, receiver.Bytes, oneEGLD, "", 50000)
		marshalledTx, err = json.Marshal(innerTx)
		require.NoError(t, err)
		txData = []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit = 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx = generateTransaction(relayer.Bytes, 1, sender.Bytes, big.NewInt(0), string(txData), gasLimit)

		relayerBalanceBefore = getBalance(t, cs, relayer)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		relayerBalanceAfter = getBalance(t, cs, relayer)

		feeConsumed = big.NewInt(0).Sub(relayerBalanceBefore, relayerBalanceAfter)

		require.Equal(t, expectedFeeAfterFix, feeConsumed.String())
	}
}

func TestRelayedTransactionInMultiShardEnvironmentWithChainSimulatorInnerNotExecutable(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t, alterConfigsFuncRelayedV3EarlyActivation)
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	sender2, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	guardian, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	// Set guardian for sender
	senderNonce := uint64(0)
	setGuardianTxData := "SetGuardian@" + hex.EncodeToString(guardian.Bytes) + "@" + hex.EncodeToString([]byte("uuid"))
	setGuardianGasLimit := minGasLimit + 1500*len(setGuardianTxData) + guardAccountCost
	setGuardianTx := generateTransaction(sender.Bytes, senderNonce, sender.Bytes, big.NewInt(0), setGuardianTxData, uint64(setGuardianGasLimit))
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(setGuardianTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	// fast-forward until the guardian becomes active
	err = cs.GenerateBlocks(roundsPerEpoch * 20)
	require.NoError(t, err)

	// guard account
	senderNonce++
	guardAccountTxData := "GuardAccount"
	guardAccountGasLimit := minGasLimit + 1500*len(guardAccountTxData) + guardAccountCost
	guardAccountTx := generateTransaction(sender.Bytes, senderNonce, sender.Bytes, big.NewInt(0), guardAccountTxData, uint64(guardAccountGasLimit))
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(guardAccountTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	receiver, err := cs.GenerateAndMintWalletAddress(1, big.NewInt(0))
	require.NoError(t, err)

	// move balance inner transaction non-executable due to guardian mismatch
	senderNonce++
	innerTx := generateTransaction(sender.Bytes, senderNonce, receiver.Bytes, oneEGLD, "", minGasLimit+extraGasLimitForGuarded)
	innerTx.RelayerAddr = relayer.Bytes
	innerTx.GuardianAddr = sender.Bytes // this is not the real guardian
	innerTx.GuardianSignature = []byte(mockTxSignature)
	innerTx.Options = 2

	// move balance inner transaction non-executable due to higher nonce
	nonceTooHigh := uint64(100)
	innerTx2 := generateTransaction(sender2.Bytes, nonceTooHigh, receiver.Bytes, oneEGLD, "", minGasLimit)
	innerTx2.RelayerAddr = relayer.Bytes

	innerTxs := []*transaction.Transaction{innerTx, innerTx2}

	// relayer will consume first a move balance for each inner tx, then the specific gas for each inner tx
	relayedTxGasLimit := uint64(0)
	for _, tx := range innerTxs {
		relayedTxGasLimit += minGasLimit + tx.GasLimit
	}
	relayedTx := generateTransaction(relayer.Bytes, 0, relayer.Bytes, big.NewInt(0), "", relayedTxGasLimit)
	relayedTx.InnerTransactions = innerTxs

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	// generate few more blocks for the cross shard scrs to be done
	err = cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	// check the inner tx failed with the desired error
	require.Equal(t, 2, len(result.SmartContractResults))
	require.True(t, strings.Contains(result.SmartContractResults[0].ReturnMessage, process.ErrTransactionNotExecutable.Error()))
	require.True(t, strings.Contains(result.SmartContractResults[0].ReturnMessage, process.ErrTransactionAndAccountGuardianMismatch.Error()))
	require.True(t, strings.Contains(result.SmartContractResults[1].ReturnMessage, process.ErrHigherNonceInTransaction.Error()))

	// check events
	require.Equal(t, 2, len(result.Logs.Events))
	for _, event := range result.Logs.Events {
		require.Equal(t, core.SignalErrorOperation, event.Identifier)
	}

	// compute expected consumed fee for relayer
	expectedConsumedGasForGuardedInnerTx := minGasLimit + minGasLimit + extraGasLimitForGuarded // invalid guardian
	expectedConsumedGasForHigherNonceInnerTx := minGasLimit + minGasLimit                       // higher nonce
	expectedConsumeGas := expectedConsumedGasForGuardedInnerTx + expectedConsumedGasForHigherNonceInnerTx
	expectedRelayerFee := core.SafeMul(uint64(expectedConsumeGas), minGasPrice)
	checkBalance(t, cs, relayer, big.NewInt(0).Sub(initialBalance, expectedRelayerFee))

	checkBalance(t, cs, receiver, big.NewInt(0))

	relayerBalanceBeforeSuccessfullAttempt := getBalance(t, cs, relayer)

	// generate a valid guarded move balance inner tx
	// senderNonce would be the same, as previous failed tx didn't increase it(expected)
	innerTx = generateTransaction(sender.Bytes, senderNonce, receiver.Bytes, oneEGLD, "", minGasLimit+extraGasLimitForGuarded)
	innerTx.RelayerAddr = relayer.Bytes
	innerTx.GuardianAddr = guardian.Bytes
	innerTx.GuardianSignature = []byte(mockTxSignature)
	innerTx.Options = 2

	innerTxs = []*transaction.Transaction{innerTx}
	relayedTx = generateTransaction(relayer.Bytes, 1, relayer.Bytes, big.NewInt(0), "", relayedTxGasLimit)
	relayedTx.InnerTransactions = innerTxs

	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	// generate few more blocks for the cross shard scrs to be done
	err = cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	expectedRelayerFee = core.SafeMul(uint64(expectedConsumedGasForGuardedInnerTx), minGasPrice)
	checkBalance(t, cs, relayer, big.NewInt(0).Sub(relayerBalanceBeforeSuccessfullAttempt, expectedRelayerFee))

	checkBalance(t, cs, receiver, oneEGLD)
}

func TestRelayedTransactionFeeField(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t, func(cfg *config.Configs) {
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsEnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsV2EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = 1
	})
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	receiver, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	t.Run("relayed v1", func(t *testing.T) {
		innerTx := generateTransaction(sender.Bytes, 0, receiver.Bytes, oneEGLD, "", minGasLimit)
		buff, err := json.Marshal(innerTx)
		require.NoError(t, err)

		txData := []byte("relayedTx@" + hex.EncodeToString(buff))
		gasLimit := minGasLimit + len(txData)*gasPerDataByte + int(innerTx.GasLimit)
		relayedTx := generateTransaction(relayer.Bytes, 0, sender.Bytes, big.NewInt(0), string(txData), uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		expectedFee := core.SafeMul(uint64(gasLimit), minGasPrice)
		require.Equal(t, expectedFee.String(), result.Fee)
		require.Equal(t, expectedFee.String(), result.InitiallyPaidFee)
		require.Equal(t, uint64(gasLimit), result.GasUsed)
	})
	t.Run("relayed v3", func(t *testing.T) {
		innerTx := generateTransaction(sender.Bytes, 1, receiver.Bytes, oneEGLD, "", minGasLimit)
		innerTx.RelayerAddr = relayer.Bytes

		gasLimit := minGasLimit + int(innerTx.GasLimit)
		relayedTx := generateTransaction(relayer.Bytes, 1, relayer.Bytes, big.NewInt(0), "", uint64(gasLimit))
		relayedTx.InnerTransactions = []*transaction.Transaction{innerTx}

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		expectedFee := core.SafeMul(uint64(gasLimit), minGasPrice)
		require.Equal(t, expectedFee.String(), result.Fee)
		require.Equal(t, expectedFee.String(), result.InitiallyPaidFee)
		require.Equal(t, uint64(gasLimit), result.GasUsed)
	})
}

func startChainSimulator(
	t *testing.T,
	alterConfigsFunction func(cfg *config.Configs),
) testsChainSimulator.ChainSimulator {
	roundDurationInMillis := uint64(6000)
	roundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    roundsPerEpoch,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   true,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              3,
		GenesisTimestamp:         time.Now().Unix(),
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpochOpt,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  3,
		NumNodesWaitingListShard: 3,
		AlterConfigsFunction:     alterConfigsFunction,
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

func checkSCRSucceeded(
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

func getBalance(
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

func checkBalance(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	address dtos.WalletAddress,
	expectedBalance *big.Int,
) {
	balance := getBalance(t, cs, address)
	require.Equal(t, expectedBalance.String(), balance.String())
}

func deployAdder(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	owner dtos.WalletAddress,
	ownerNonce uint64,
) []byte {
	pkConv := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()

	err := cs.GenerateBlocks(1)
	require.Nil(t, err)

	scCode := wasm.GetSCCode("testData/adder.wasm")
	params := []string{scCode, wasm.VMTypeHex, wasm.DummyCodeMetadataHex, "00"}
	txDataDeploy := strings.Join(params, "@")
	deployTx := generateTransaction(owner.Bytes, ownerNonce, make([]byte, 32), big.NewInt(0), txDataDeploy, 100000000)

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(deployTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	scAddress := result.Logs.Events[0].Address
	scAddressBytes, _ := pkConv.Decode(scAddress)

	return scAddressBytes
}
