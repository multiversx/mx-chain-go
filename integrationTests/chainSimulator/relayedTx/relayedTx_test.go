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
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig              = "../../../cmd/node/config/"
	minGasPrice                             = 1_000_000_000
	minGasLimit                             = 50_000
	gasPerDataByte                          = 1_500
	deductionFactor                         = 100
	txVersion                               = 2
	mockTxSignature                         = "ssig"
	mockRelayerTxSignature                  = "rsig"
	maxNumOfBlocksToGenerateWhenExecutingTx = 10
	roundsPerEpoch                          = 30
)

var (
	oneEGLD = big.NewInt(1000000000000000000)
)

func TestRelayedV3WithChainSimulator(t *testing.T) {
	t.Run("successful intra shard move balance with exact gas", testRelayedV3MoveBalance(0, 0, false))
	t.Run("successful intra shard move balance with extra gas", testRelayedV3MoveBalance(0, 0, true))
	t.Run("successful cross shard move balance with exact gas", testRelayedV3MoveBalance(0, 1, false))
	t.Run("successful cross shard move balance with extra gas", testRelayedV3MoveBalance(0, 1, true))
	t.Run("intra shard move balance, lower nonce", testRelayedV3MoveBalanceLowerNonce(0, 0))
	t.Run("cross shard move balance, lower nonce", testRelayedV3MoveBalanceLowerNonce(0, 1))
	t.Run("intra shard move balance, invalid gas", testRelayedV3MoveInvalidGasLimit(0, 0))
	t.Run("cross shard move balance, invalid gas", testRelayedV3MoveInvalidGasLimit(0, 1))

	t.Run("successful intra shard sc call with refunds", testRelayedV3ScCall(0, 0))
	t.Run("successful cross shard sc call with refunds", testRelayedV3ScCall(0, 1))
	t.Run("intra shard sc call, invalid gas", testRelayedV3ScCallInvalidGasLimit(0, 0))
	t.Run("cross shard sc call, invalid gas", testRelayedV3ScCallInvalidGasLimit(0, 1))
	t.Run("intra shard sc call, invalid method", testRelayedV3ScCallInvalidMethod(0, 0))
	t.Run("cross shard sc call, invalid method", testRelayedV3ScCallInvalidMethod(0, 1))
}

func testRelayedV3MoveBalance(
	relayerShard uint32,
	destinationShard uint32,
	extraGas bool,
) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip("this is not a short test")
		}

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		receiver, err := cs.GenerateAndMintWalletAddress(destinationShard, big.NewInt(0))
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		gasLimit := minGasLimit * 2
		extraGasLimit := 0
		if extraGas {
			extraGasLimit = minGasLimit
		}
		relayedTx := generateRelayedV3Transaction(sender.Bytes, 0, receiver.Bytes, relayer.Bytes, oneEGLD, "", uint64(gasLimit+extraGasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check fee fields
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, big.NewInt(0), true)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := getBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance
		senderBalanceAfter := getBalance(t, cs, sender)
		senderBalanceDiff := big.NewInt(0).Sub(initialBalance, senderBalanceAfter)
		require.Equal(t, oneEGLD.String(), senderBalanceDiff.String())

		// check receiver balance
		receiverBalanceAfter := getBalance(t, cs, receiver)
		require.Equal(t, oneEGLD.String(), receiverBalanceAfter.String())

		// check scr
		require.Equal(t, 1, len(result.SmartContractResults))
		require.Equal(t, relayer.Bech32, result.SmartContractResults[0].RelayerAddr)
		require.Equal(t, sender.Bech32, result.SmartContractResults[0].SndAddr)
		require.Equal(t, receiver.Bech32, result.SmartContractResults[0].RcvAddr)
		require.Equal(t, relayedTx.Value, result.SmartContractResults[0].Value)

		// check intra shard logs, should be none
		require.Nil(t, result.Logs)

		// check cross shard log, should be one completedTxEvent
		if relayerShard == destinationShard {
			return
		}
		scrResult, err := cs.GetNodeHandler(destinationShard).GetFacadeHandler().GetTransaction(result.SmartContractResults[0].Hash, true)
		require.NoError(t, err)
		require.NotNil(t, scrResult.Logs)
		require.Equal(t, 1, len(scrResult.Logs.Events))
		require.Contains(t, scrResult.Logs.Events[0].Identifier, core.CompletedTxEventIdentifier)
	}
}

func testRelayedV3MoveBalanceLowerNonce(
	relayerShard uint32,
	receiverShard uint32,
) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip("this is not a short test")
		}

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		receiver, err := cs.GenerateAndMintWalletAddress(receiverShard, big.NewInt(0))
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		gasLimit := minGasLimit * 2
		relayedTx := generateRelayedV3Transaction(sender.Bytes, 0, receiver.Bytes, relayer.Bytes, oneEGLD, "", uint64(gasLimit))

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// send same tx again, lower nonce
		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.Contains(t, err.Error(), process.ErrWrongTransaction.Error())
		require.Nil(t, result)
	}
}

func testRelayedV3MoveInvalidGasLimit(
	relayerShard uint32,
	receiverShard uint32,
) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip("this is not a short test")
		}

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		receiver, err := cs.GenerateAndMintWalletAddress(receiverShard, big.NewInt(0))
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		gasLimit := minGasLimit
		relayedTx := generateRelayedV3Transaction(sender.Bytes, 0, receiver.Bytes, relayer.Bytes, oneEGLD, "", uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.Contains(t, err.Error(), process.ErrInsufficientGasLimitInTx.Error())
		require.Nil(t, result)
	}
}

func testRelayedV3ScCall(
	relayerShard uint32,
	ownerShard uint32,
) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip("this is not a short test")
		}

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		owner, err := cs.GenerateAndMintWalletAddress(ownerShard, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		resultDeploy, scAddressBytes := deployAdder(t, cs, owner, 0)
		refundDeploy := getRefundValue(resultDeploy.SmartContractResults)

		// send relayed tx
		txDataAdd := "add@" + hex.EncodeToString(big.NewInt(1).Bytes())
		gasLimit := uint64(3000000)
		relayedTx := generateRelayedV3Transaction(sender.Bytes, 0, scAddressBytes, relayer.Bytes, big.NewInt(0), txDataAdd, gasLimit)

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// if cross shard, generate few more blocks for eventual refunds to be executed
		if relayerShard != ownerShard {
			require.NoError(t, cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx))
		}

		checkSum(t, cs.GetNodeHandler(ownerShard), scAddressBytes, owner.Bytes, 1)

		refundValue := getRefundValue(result.SmartContractResults)
		require.NotZero(t, refundValue.Uint64())

		// check fee fields
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, refundValue, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := getBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance
		senderBalanceAfter := getBalance(t, cs, sender)
		require.Equal(t, initialBalance.String(), senderBalanceAfter.String())

		// check owner balance
		_, feeDeploy, _ := computeTxGasAndFeeBasedOnRefund(resultDeploy, refundDeploy, false)
		ownerBalanceAfter := getBalance(t, cs, owner)
		ownerFee := big.NewInt(0).Sub(initialBalance, ownerBalanceAfter)
		require.Equal(t, feeDeploy.String(), ownerFee.String())

		// check scrs
		require.Equal(t, 2, len(result.SmartContractResults))
		for _, scr := range result.SmartContractResults {
			checkSCRSucceeded(t, cs, scr)
		}
	}
}

func testRelayedV3ScCallInvalidGasLimit(
	relayerShard uint32,
	ownerShard uint32,
) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip("this is not a short test")
		}

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		owner, err := cs.GenerateAndMintWalletAddress(ownerShard, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		_, scAddressBytes := deployAdder(t, cs, owner, 0)

		// send relayed tx with less gas limit
		txDataAdd := "add@" + hex.EncodeToString(big.NewInt(1).Bytes())
		gasLimit := gasPerDataByte*len(txDataAdd) + minGasLimit + minGasLimit
		relayedTx := generateRelayedV3Transaction(sender.Bytes, 0, scAddressBytes, relayer.Bytes, big.NewInt(0), txDataAdd, uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		logs := result.Logs
		// if cross shard, generate few more blocks for cross shard scrs
		if relayerShard != ownerShard {
			require.NoError(t, cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx))
			logs = result.SmartContractResults[0].Logs
		}

		require.NotNil(t, logs)
		require.Equal(t, 2, len(logs.Events))
		for _, event := range logs.Events {
			if event.Identifier == core.SignalErrorOperation {
				continue
			}

			require.Equal(t, 1, len(event.AdditionalData))
			require.Contains(t, string(event.AdditionalData[0]), "[not enough gas]")
		}

		refundValue := getRefundValue(result.SmartContractResults)
		require.Zero(t, refundValue.Uint64())

		// check fee fields, should consume full gas
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, refundValue, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, result.InitiallyPaidFee, result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)
		require.Equal(t, relayedTx.GasLimit, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := getBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance
		senderBalanceAfter := getBalance(t, cs, sender)
		require.Equal(t, initialBalance.String(), senderBalanceAfter.String())
	}
}

func testRelayedV3ScCallInvalidMethod(
	relayerShard uint32,
	ownerShard uint32,
) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip("this is not a short test")
		}

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		owner, err := cs.GenerateAndMintWalletAddress(ownerShard, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		_, scAddressBytes := deployAdder(t, cs, owner, 0)

		// send relayed tx with invalid value
		txDataAdd := "invalid"
		gasLimit := uint64(3000000)
		relayedTx := generateRelayedV3Transaction(sender.Bytes, 0, scAddressBytes, relayer.Bytes, big.NewInt(0), txDataAdd, uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		logs := result.Logs
		// if cross shard, generate few more blocks for cross shard scrs
		if relayerShard != ownerShard {
			require.NoError(t, cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx))
			logs = result.SmartContractResults[0].Logs
		}

		require.NotNil(t, logs)
		require.Equal(t, 2, len(logs.Events))
		for _, event := range logs.Events {
			if event.Identifier == core.SignalErrorOperation {
				continue
			}

			require.Equal(t, 1, len(event.AdditionalData))
			require.Contains(t, string(event.AdditionalData[0]), "[invalid function (not found)]")
		}

		refundValue := getRefundValue(result.SmartContractResults)
		require.Zero(t, refundValue.Uint64()) // no refund, tx failed

		// check fee fields, should consume full gas
		initiallyPaidFee, fee, _ := computeTxGasAndFeeBasedOnRefund(result, refundValue, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)

		// check relayer balance
		relayerBalanceAfter := getBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance
		senderBalanceAfter := getBalance(t, cs, sender)
		require.Equal(t, initialBalance.String(), senderBalanceAfter.String())
	}
}

func TestFixRelayedMoveBalanceWithChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	expectedFeeScCallBefore := "827294920000000"
	expectedFeeScCallAfter := "885704920000000"
	t.Run("sc call", testFixRelayedMoveBalanceWithChainSimulatorScCall(expectedFeeScCallBefore, expectedFeeScCallAfter))

	expectedFeeMoveBalanceBefore := "809500000000000" // 506 * 1500 + 50000 + 500
	expectedFeeMoveBalanceAfter := "859000000000000"  // 506 * 1500 + 50000 + 50000
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
		_, scAddressBytes := deployAdder(t, cs, owner, ownerNonce)

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

func TestRelayedTransactionFeeField(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t, func(cfg *config.Configs) {
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsEnableEpoch = 1
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

func generateRelayedV3Transaction(sender []byte, nonce uint64, receiver []byte, relayer []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	tx := generateTransaction(sender, nonce, receiver, value, data, gasLimit)
	tx.RelayerSignature = []byte(mockRelayerTxSignature)
	tx.RelayerAddr = relayer
	return tx
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

func deployAdder(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	owner dtos.WalletAddress,
	ownerNonce uint64,
) (*transaction.ApiTransactionResult, []byte) {
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

	return result, scAddressBytes
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

func getRefundValue(scrs []*transaction.ApiSmartContractResult) *big.Int {
	for _, scr := range scrs {
		if scr.IsRefund {
			return scr.Value
		}
	}

	return big.NewInt(0)
}

func computeTxGasAndFeeBasedOnRefund(
	result *transaction.ApiTransactionResult,
	refund *big.Int,
	isMoveBalance bool,
) (*big.Int, *big.Int, uint64) {
	deductedGasPrice := uint64(minGasPrice / deductionFactor)

	initialTx := result.Tx
	gasForFullPrice := uint64(minGasLimit + gasPerDataByte*len(initialTx.GetData()))
	if result.ProcessingTypeOnSource == process.RelayedTxV3.String() {
		gasForFullPrice += uint64(minGasLimit) // relayer fee
	}
	gasForDeductedPrice := initialTx.GetGasLimit() - gasForFullPrice

	initialFee := gasForFullPrice*minGasPrice + gasForDeductedPrice*deductedGasPrice
	finalFee := initialFee - refund.Uint64()

	gasRefunded := refund.Uint64() / deductedGasPrice
	gasConsumed := gasForFullPrice + gasForDeductedPrice - gasRefunded

	if isMoveBalance {
		return big.NewInt(0).SetUint64(initialFee), big.NewInt(0).SetUint64(gasForFullPrice * minGasPrice), gasForFullPrice
	}

	return big.NewInt(0).SetUint64(initialFee), big.NewInt(0).SetUint64(finalFee), gasConsumed
}

func checkSCRSucceeded(
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
