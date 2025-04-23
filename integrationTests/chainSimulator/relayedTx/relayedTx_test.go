package relayedTx

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	apiData "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"

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
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/vm"
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
	guardAccountCost                        = 250_000
	extraGasLimitForGuarded                 = minGasLimit
	extraGasESDTTransfer                    = 250000
	extraGasMultiESDTTransferPerToken       = 1100000
	egldTicker                              = "EGLD-000000"
)

var (
	oneEGLD = big.NewInt(1000000000000000000)
)

func TestRelayedV3WithChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("successful intra shard guarded move balance", testRelayedV3MoveBalance(0, 0, false, true))
	t.Run("sender == relayer move balance should consume fee", testRelayedV3RelayedBySenderMoveBalance())
	t.Run("receiver == relayer move balance should consume fee", testRelayedV3RelayedByReceiverMoveBalance())
	t.Run("successful intra shard move balance", testRelayedV3MoveBalance(0, 0, false, false))
	t.Run("successful intra shard move balance with extra gas", testRelayedV3MoveBalance(0, 0, true, false))
	t.Run("successful cross shard move balance", testRelayedV3MoveBalance(0, 1, false, false))
	t.Run("successful cross shard guarded move balance", testRelayedV3MoveBalance(0, 1, false, true))
	t.Run("successful cross shard move balance with extra gas", testRelayedV3MoveBalance(0, 1, true, false))
	t.Run("intra shard move balance, lower nonce", testRelayedV3MoveBalanceLowerNonce(0, 0))
	t.Run("cross shard move balance, lower nonce", testRelayedV3MoveBalanceLowerNonce(0, 1))
	t.Run("intra shard move balance, invalid gas", testRelayedV3MoveInvalidGasLimit(0, 0))
	t.Run("cross shard move balance, invalid gas", testRelayedV3MoveInvalidGasLimit(0, 1))

	t.Run("successful intra shard sc call with refunds, existing sender", testRelayedV3ScCall(0, 0, true, false))
	t.Run("successful intra shard sc call with refunds, existing sender, relayed by sender", testRelayedV3ScCall(0, 0, true, true))
	t.Run("successful intra shard sc call with refunds, new sender", testRelayedV3ScCall(0, 0, false, false))
	t.Run("successful cross shard sc call with refunds, existing sender", testRelayedV3ScCall(0, 1, true, false))
	t.Run("successful cross shard sc call with refunds, existing sender, relayed by sender", testRelayedV3ScCall(0, 1, true, true))
	t.Run("successful cross shard sc call with refunds, new sender", testRelayedV3ScCall(0, 1, false, false))
	t.Run("intra shard sc call, invalid gas", testRelayedV3ScCallInvalidGasLimit(0, 0))
	t.Run("cross shard sc call, invalid gas", testRelayedV3ScCallInvalidGasLimit(0, 1))
	t.Run("intra shard sc call, invalid method", testRelayedV3ScCallInvalidMethod(0, 0))
	t.Run("cross shard sc call, invalid method", testRelayedV3ScCallInvalidMethod(0, 1))

	t.Run("create new delegation contract", testRelayedV3MetaInteraction())

	t.Run("esdt transfer", func(t *testing.T) {
		t.Run("receiver == relayer, sender is a new account", testRelayedV3ESDTTransfer(true, false))
		t.Run("receiver == relayer, sender is token issuer", testRelayedV3ESDTTransfer(true, true))
		t.Run("receiver != relayer, sender is a new account", testRelayedV3ESDTTransfer(false, false))
		t.Run("receiver != relayer, sender is token issuer", testRelayedV3ESDTTransfer(false, true))
	})

	t.Run("multi esdt transfer", func(t *testing.T) {
		t.Run("receiver == relayer, sender is a new account", testRelayedV3MultiESDTTransferWithEGLD(true, false))
		t.Run("receiver == relayer, sender is token issuer", testRelayedV3MultiESDTTransferWithEGLD(true, true))
		t.Run("receiver != relayer, sender is a new account", testRelayedV3MultiESDTTransferWithEGLD(false, false))
		t.Run("receiver != relayer, sender is token issuer", testRelayedV3MultiESDTTransferWithEGLD(false, true))
	})
}

func testRelayedV3MoveBalance(
	relayerShard uint32,
	destinationShard uint32,
	extraGas bool,
	guardedTx bool,
) func(t *testing.T) {
	return func(t *testing.T) {
		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
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

		guardian, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		senderNonce := uint64(0)
		if guardedTx {
			// Set guardian for sender
			setGuardianTxData := "SetGuardian@" + hex.EncodeToString(guardian.Bytes) + "@" + hex.EncodeToString([]byte("uuid"))
			setGuardianGasLimit := minGasLimit + 1500*len(setGuardianTxData) + guardAccountCost
			setGuardianTx := generateTransaction(sender.Bytes, senderNonce, sender.Bytes, big.NewInt(0), setGuardianTxData, uint64(setGuardianGasLimit))
			_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(setGuardianTx, maxNumOfBlocksToGenerateWhenExecutingTx)
			require.NoError(t, err)
			senderNonce++

			// fast-forward until the guardian becomes active
			err = cs.GenerateBlocks(roundsPerEpoch * 20)
			require.NoError(t, err)

			// guard account
			guardAccountTxData := "GuardAccount"
			guardAccountGasLimit := minGasLimit + 1500*len(guardAccountTxData) + guardAccountCost
			guardAccountTx := generateTransaction(sender.Bytes, senderNonce, sender.Bytes, big.NewInt(0), guardAccountTxData, uint64(guardAccountGasLimit))
			_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(guardAccountTx, maxNumOfBlocksToGenerateWhenExecutingTx)
			require.NoError(t, err)
			senderNonce++
		}

		senderBalanceBefore := getBalance(t, cs, sender)

		gasLimit := minGasLimit * 2
		extraGasLimit := 0
		if extraGas {
			extraGasLimit = minGasLimit
		}
		if guardedTx {
			gasLimit += extraGasLimitForGuarded
		}
		relayedTx := generateRelayedV3Transaction(sender.Bytes, senderNonce, receiver.Bytes, relayer.Bytes, oneEGLD, "", uint64(gasLimit+extraGasLimit))

		if guardedTx {
			relayedTx.GuardianAddr = guardian.Bytes
			relayedTx.GuardianSignature = []byte(mockTxSignature)
			relayedTx.Options = 2
		}

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		isIntraShard := relayerShard == destinationShard
		if isIntraShard {
			require.NoError(t, cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx))
		}

		// check fee fields
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, big.NewInt(0), true, guardedTx)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := getBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance
		senderBalanceAfter := getBalance(t, cs, sender)
		senderBalanceDiff := big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter)
		require.Equal(t, oneEGLD.String(), senderBalanceDiff.String())

		// check receiver balance
		receiverBalanceAfter := getBalance(t, cs, receiver)
		require.Equal(t, oneEGLD.String(), receiverBalanceAfter.String())

		// check scrs, should be none
		require.Zero(t, len(result.SmartContractResults))

		// check intra shard logs, should be none
		require.Nil(t, result.Logs)

		if extraGas && isIntraShard {
			require.NotNil(t, result.Receipt)
		}
	}
}

func testRelayedV3MoveBalanceLowerNonce(
	relayerShard uint32,
	receiverShard uint32,
) func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
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

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
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
	existingSenderWithBalance bool,
	relayedBySender bool,
) func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)
		relayerInitialBalance := initialBalance

		sender, senderInitialBalance := prepareSender(t, cs, existingSenderWithBalance, relayerShard, initialBalance)
		if relayedBySender {
			relayer = sender
			relayerInitialBalance = senderInitialBalance
		}

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
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := getBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(relayerInitialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance, only if the tx was not relayed by sender
		if !relayedBySender {
			senderBalanceAfter := getBalance(t, cs, sender)
			require.Equal(t, senderInitialBalance.String(), senderBalanceAfter.String())
		}

		// check owner balance
		_, feeDeploy, _ := computeTxGasAndFeeBasedOnRefund(resultDeploy, refundDeploy, false, false)
		ownerBalanceAfter := getBalance(t, cs, owner)
		ownerFee := big.NewInt(0).Sub(initialBalance, ownerBalanceAfter)
		require.Equal(t, feeDeploy.String(), ownerFee.String())

		// check scrs
		require.Equal(t, 1, len(result.SmartContractResults))
		for _, scr := range result.SmartContractResults {
			checkSCRSucceeded(t, cs, scr)
		}
	}
}

func testRelayedV3RelayedBySenderMoveBalance() func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))

		sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		senderNonce := uint64(0)
		senderBalanceBefore := getBalance(t, cs, sender)

		gasLimit := minGasLimit * 2
		relayedTx := generateRelayedV3Transaction(sender.Bytes, senderNonce, sender.Bytes, sender.Bytes, big.NewInt(0), "", uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check fee fields
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, big.NewInt(0), true, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check sender balance
		expectedFee := core.SafeMul(uint64(gasLimit), uint64(minGasPrice))
		senderBalanceAfter := getBalance(t, cs, sender)
		senderBalanceDiff := big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter)
		require.Equal(t, expectedFee.String(), senderBalanceDiff.String())

		// check scrs, should be none
		require.Zero(t, len(result.SmartContractResults))

		// check intra shard logs, should be none
		require.Nil(t, result.Logs)
	}
}

func testRelayedV3RelayedByReceiverMoveBalance() func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))

		sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		receiver, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		senderNonce := uint64(0)
		receiverBalanceBefore := getBalance(t, cs, receiver)

		gasLimit := minGasLimit * 2
		relayedTx := generateRelayedV3Transaction(sender.Bytes, senderNonce, receiver.Bytes, receiver.Bytes, big.NewInt(0), "", uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check fee fields
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, big.NewInt(0), true, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check sender balance
		senderBalanceAfter := getBalance(t, cs, sender)
		require.Equal(t, senderBalanceAfter.String(), initialBalance.String())

		// check receiver balance
		expectedFee := core.SafeMul(uint64(gasLimit), uint64(minGasPrice))
		receiverBalanceAfter := getBalance(t, cs, receiver)
		receiverBalanceDiff := big.NewInt(0).Sub(receiverBalanceBefore, receiverBalanceAfter)
		require.Equal(t, receiverBalanceDiff.String(), expectedFee.String())

		// check scrs, should be none
		require.Zero(t, len(result.SmartContractResults))

		// check intra shard logs, should be none
		require.Nil(t, result.Logs)
	}
}

func testRelayedV3ESDTTransfer(
	relayedByReceiver bool,
	senderIsIssuer bool,
) func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000))

		owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		sender := owner
		// if sender is not owner, make the sender a new address with no balance
		if !senderIsIssuer {
			sender, _ = prepareSender(t, cs, false, 0, big.NewInt(0))
		}

		receiver, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		relayer := receiver
		// if relayed tx won't be relayed by the receiver, generate the relayer
		if !relayedByReceiver {
			relayer, err = cs.GenerateAndMintWalletAddress(0, initialBalance)
			require.NoError(t, err)
		}

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		// issue new token
		initialSupply, _ := big.NewInt(0).SetString("1000000000", 10)
		ticker := issueToken(t, cs, owner, "TESTTOKEN", "TST", initialSupply)

		transferValue := big.NewInt(1000)
		txDataTransfer := "ESDTTransfer@" + hex.EncodeToString([]byte(ticker)) + "@" + hex.EncodeToString(transferValue.Bytes())

		// if sender is not owner, send some new tokens to the sender
		if !senderIsIssuer {
			gasLimit := minGasLimit + len(txDataTransfer)*1500 + extraGasESDTTransfer
			ownerNonce := getNonce(t, cs, owner)

			esdtTransferTx := generateTransaction(owner.Bytes, ownerNonce, sender.Bytes, big.NewInt(0), txDataTransfer, uint64(gasLimit))

			_, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(esdtTransferTx, maxNumOfBlocksToGenerateWhenExecutingTx)
			require.NoError(t, err)

			esdtBalanceSender := getESDTBalance(t, cs, sender, 0, ticker)
			require.Equal(t, transferValue.String(), esdtBalanceSender.String())
		}

		senderBalanceBefore := getBalance(t, cs, sender)

		// send relayed tx
		gasLimit := minGasLimit*2 + len(txDataTransfer)*1500 + extraGasESDTTransfer
		senderNonce := getNonce(t, cs, sender)
		relayedTx := generateRelayedV3Transaction(sender.Bytes, senderNonce, receiver.Bytes, relayer.Bytes, big.NewInt(0), txDataTransfer, uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check sender balance
		senderBalanceAfter := getBalance(t, cs, sender)
		require.Equal(t, senderBalanceAfter.String(), senderBalanceBefore.String())

		refundValue := getRefundValue(result.SmartContractResults)
		require.NotZero(t, refundValue.Uint64())

		// check fee fields
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := getBalance(t, cs, relayer)
		relayerBalanceDiff := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerBalanceDiff.String())

		// check receiver esdt balance
		expectedESDTBalanceSender := big.NewInt(0)
		if senderIsIssuer {
			expectedESDTBalanceSender = big.NewInt(0).Sub(initialSupply, transferValue)
		}
		esdtBalanceSender := getESDTBalance(t, cs, sender, 0, ticker)
		require.Equal(t, expectedESDTBalanceSender.String(), esdtBalanceSender.String())

		esdtBalanceReceiver := getESDTBalance(t, cs, receiver, 0, ticker)
		require.Equal(t, transferValue.String(), esdtBalanceReceiver.String())

		// check receiver egld balance unchanged if tx is relayed by third party
		if !relayedByReceiver {
			receiverBalanceAfter := getBalance(t, cs, receiver)
			require.Equal(t, initialBalance.String(), receiverBalanceAfter.String())
		}
	}
}

func testRelayedV3MultiESDTTransferWithEGLD(
	relayedByReceiver bool,
	senderIsIssuer bool,
) func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000))

		owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		sender := owner
		// if sender is not owner, make the sender a new address with no balance
		if !senderIsIssuer {
			sender, _ = prepareSender(t, cs, false, 0, big.NewInt(0))
		}

		receiver, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		relayer := receiver
		// if relayed tx won't be relayed by the receiver, generate the relayer
		if !relayedByReceiver {
			relayer, err = cs.GenerateAndMintWalletAddress(0, initialBalance)
			require.NoError(t, err)
		}

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		// issue new token
		initialSupply, _ := big.NewInt(0).SetString("1000000000", 10)
		ticker := issueToken(t, cs, owner, "TESTTOKEN", "TST", initialSupply)

		transferValue := big.NewInt(1000)
		egldTransferValue := oneEGLD

		// if sender is not owner, send some new tokens to the sender
		if !senderIsIssuer {
			txDataTransfer := "MultiESDTNFTTransfer@" +
				hex.EncodeToString(sender.Bytes) + "@02@" +
				hex.EncodeToString([]byte(ticker)) + "@00@" + hex.EncodeToString(transferValue.Bytes()) + "@" +
				hex.EncodeToString([]byte(egldTicker)) + "@00@" + hex.EncodeToString(egldTransferValue.Bytes())
			gasLimit := minGasLimit + len(txDataTransfer)*1500 + extraGasMultiESDTTransferPerToken*2
			ownerNonce := getNonce(t, cs, owner)

			esdtTransferTx := generateTransaction(owner.Bytes, ownerNonce, owner.Bytes, big.NewInt(0), txDataTransfer, uint64(gasLimit))

			_, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(esdtTransferTx, maxNumOfBlocksToGenerateWhenExecutingTx)
			require.NoError(t, err)

			esdtBalanceSender := getESDTBalance(t, cs, sender, 0, ticker)
			require.Equal(t, transferValue.String(), esdtBalanceSender.String())

			balanceSender := getBalance(t, cs, sender)
			require.Equal(t, egldTransferValue.String(), balanceSender.String())
		}

		senderBalanceBefore := getBalance(t, cs, sender)

		// send relayed tx
		txDataTransfer := "MultiESDTNFTTransfer@" +
			hex.EncodeToString(receiver.Bytes) + "@02@" +
			hex.EncodeToString([]byte(ticker)) + "@00@" + hex.EncodeToString(transferValue.Bytes()) + "@" +
			hex.EncodeToString([]byte(egldTicker)) + "@00@" + hex.EncodeToString(egldTransferValue.Bytes())
		gasLimit := minGasLimit*2 + len(txDataTransfer)*1500 + extraGasMultiESDTTransferPerToken*2
		senderNonce := getNonce(t, cs, sender)
		relayedTx := generateRelayedV3Transaction(sender.Bytes, senderNonce, sender.Bytes, relayer.Bytes, big.NewInt(0), txDataTransfer, uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check sender balance
		senderBalanceAfter := getBalance(t, cs, sender)
		expectedSenderBalanceDiff := egldTransferValue
		senderBalanceDiff := big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter)
		require.Equal(t, expectedSenderBalanceDiff.String(), senderBalanceDiff.String())

		refundValue := getRefundValue(result.SmartContractResults)
		require.NotZero(t, refundValue.Uint64())

		// check fee fields
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := getBalance(t, cs, relayer)
		relayerBalanceDiff := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		if relayedByReceiver {
			relayerBalanceDiff.Add(relayerBalanceDiff, egldTransferValue)
		}
		require.Equal(t, fee.String(), relayerBalanceDiff.String())

		// check receiver esdt balance
		expectedESDTBalanceSender := big.NewInt(0)
		if senderIsIssuer {
			expectedESDTBalanceSender = big.NewInt(0).Sub(initialSupply, transferValue)
		}
		esdtBalanceSender := getESDTBalance(t, cs, sender, 0, ticker)
		require.Equal(t, expectedESDTBalanceSender.String(), esdtBalanceSender.String())

		esdtBalanceReceiver := getESDTBalance(t, cs, receiver, 0, ticker)
		require.Equal(t, transferValue.String(), esdtBalanceReceiver.String())

		// check receiver egld balance
		if !relayedByReceiver {
			receiverBalanceAfter := getBalance(t, cs, receiver)
			expectedReceiverBalanceAfter := big.NewInt(0).Add(initialBalance, egldTransferValue)
			require.Equal(t, expectedReceiverBalanceAfter.String(), receiverBalanceAfter.String())
		}
	}
}

func issueToken(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	owner dtos.WalletAddress,
	tokenName string,
	tokenTicker string,
	initSupply *big.Int,
) string {
	issueTxData := "issue@" +
		hex.EncodeToString([]byte(tokenName)) + "@" +
		hex.EncodeToString([]byte(tokenTicker)) + "@" +
		hex.EncodeToString(initSupply.Bytes()) + "@" +
		"02"

	ownerNonce := getNonce(t, cs, owner)
	fiveEGLD := big.NewInt(0).Mul(oneEGLD, big.NewInt(5))
	issueTx := generateTransaction(owner.Bytes, ownerNonce, vm.ESDTSCAddress, fiveEGLD, issueTxData, 60000000)
	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(issueTx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	// generate few more blocks for refund to happen
	_ = cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx)

	require.NotNil(t, result.Logs)
	require.NotZero(t, len(result.Logs.Events))
	for _, log := range result.Logs.Events {
		if log.Identifier != "issue" {
			continue
		}

		ticker := log.Topics[0]
		balance := getESDTBalance(t, cs, owner, 0, string(ticker))
		require.Equal(t, initSupply.String(), balance.String())

		return string(ticker)
	}

	return ""
}

func prepareSender(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	existingSenderWithBalance bool,
	shard uint32,
	initialBalance *big.Int,
) (dtos.WalletAddress, *big.Int) {
	if existingSenderWithBalance {
		sender, err := cs.GenerateAndMintWalletAddress(shard, initialBalance)
		require.NoError(t, err)

		return sender, initialBalance
	}

	shardC := cs.GetNodeHandler(shard).GetShardCoordinator()
	pkConv := cs.GetNodeHandler(shard).GetCoreComponents().AddressPubKeyConverter()
	newAddress := generateAddressInShard(shardC, pkConv.Len())
	return dtos.WalletAddress{
		Bech32: pkConv.SilentEncode(newAddress, logger.GetOrCreate("tmp")),
		Bytes:  newAddress,
	}, big.NewInt(0)
}

func generateAddressInShard(shardCoordinator sharding.Coordinator, len int) []byte {
	for {
		buff := generateAddress(len)
		shardID := shardCoordinator.ComputeId(buff)
		if shardID == shardCoordinator.SelfId() {
			return buff
		}
	}
}

func generateAddress(len int) []byte {
	buff := make([]byte, len)
	_, _ = rand.Read(buff)

	return buff
}

func testRelayedV3ScCallInvalidGasLimit(
	relayerShard uint32,
	ownerShard uint32,
) func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
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

		require.NotNil(t, result.Logs)
		require.Equal(t, 2, len(result.Logs.Events))
		for _, event := range result.Logs.Events {
			if event.Identifier == core.SignalErrorOperation {
				continue
			}

			require.Equal(t, 1, len(event.AdditionalData))
			require.Contains(t, string(event.AdditionalData[0]), "[not enough gas]")
		}

		refundValue := getRefundValue(result.SmartContractResults)
		require.Zero(t, refundValue.Uint64())

		// check fee fields, should consume full gas
		initiallyPaidFee, fee, gasUsed := computeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
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

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
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

		require.NotNil(t, result.Logs)
		require.Equal(t, 2, len(result.Logs.Events))
		for _, event := range result.Logs.Events {
			if event.Identifier == core.SignalErrorOperation {
				continue
			}

			require.Equal(t, 1, len(event.AdditionalData))
			require.Contains(t, string(event.AdditionalData[0]), "[invalid function (not found)]")
		}

		refundValue := getRefundValue(result.SmartContractResults)
		require.Zero(t, refundValue.Uint64()) // no refund, tx failed

		// check fee fields, should consume full gas
		initiallyPaidFee, fee, _ := computeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
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

func testRelayedV3MetaInteraction() func(t *testing.T) {
	return func(t *testing.T) {

		providedActivationEpoch := uint32(1)
		alterConfigsFunc := func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3EnableEpoch = providedActivationEpoch
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV3FixESDTTransferEnableEpoch = providedActivationEpoch
		}

		cs := startChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		relayerShard := uint32(0)

		initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10000))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		// send createNewDelegationContract transaction
		txData := "createNewDelegationContract@00@00"
		gasLimit := uint64(60000000)
		value := big.NewInt(0).Mul(oneEGLD, big.NewInt(1250))
		relayedTx := generateRelayedV3Transaction(sender.Bytes, 0, vm.DelegationManagerSCAddress, relayer.Bytes, value, txData, gasLimit)

		relayerBefore := getBalance(t, cs, relayer)
		senderBefore := getBalance(t, cs, sender)

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, maxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		require.NoError(t, cs.GenerateBlocks(maxNumOfBlocksToGenerateWhenExecutingTx))

		relayerAfter := getBalance(t, cs, relayer)
		senderAfter := getBalance(t, cs, sender)

		// check consumed fees
		refund := getRefundValue(result.SmartContractResults)
		consumedFee := big.NewInt(0).Sub(relayerBefore, relayerAfter)

		gasForFullPrice := uint64(len(txData)*gasPerDataByte + minGasLimit + minGasLimit)
		gasForDeductedPrice := gasLimit - gasForFullPrice
		deductedGasPrice := uint64(minGasPrice / deductionFactor)
		initialFee := gasForFullPrice*minGasPrice + gasForDeductedPrice*deductedGasPrice
		initialFeeInt := big.NewInt(0).SetUint64(initialFee)
		expectedConsumedFee := big.NewInt(0).Sub(initialFeeInt, refund)

		gasUsed := gasForFullPrice + gasForDeductedPrice - refund.Uint64()/deductedGasPrice
		require.Equal(t, expectedConsumedFee.String(), consumedFee.String())
		require.Equal(t, value.String(), big.NewInt(0).Sub(senderBefore, senderAfter).String(), "sender should have consumed the value only")
		require.Equal(t, initialFeeInt.String(), result.InitiallyPaidFee)
		require.Equal(t, expectedConsumedFee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)
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

func TestRegularMoveBalanceWithRefundReceipt(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startChainSimulator(t, func(cfg *config.Configs) {})
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	// generate one block so the minting has effect
	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	senderNonce := uint64(0)

	extraGas := uint64(minGasLimit)
	gasLimit := minGasLimit + extraGas
	tx := generateTransaction(sender.Bytes, senderNonce, sender.Bytes, oneEGLD, "", gasLimit)

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	require.NotNil(t, result.Receipt)
	expectedGasRefunded := core.SafeMul(extraGas, minGasPrice/deductionFactor)
	require.Equal(t, expectedGasRefunded.String(), result.Receipt.Value.String())
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

func getESDTBalance(
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

func getNonce(
	t *testing.T,
	cs testsChainSimulator.ChainSimulator,
	address dtos.WalletAddress,
) uint64 {
	account, err := cs.GetAccount(address)
	require.NoError(t, err)

	return account.Nonce
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
	guardedTx bool,
) (*big.Int, *big.Int, uint64) {
	deductedGasPrice := uint64(minGasPrice / deductionFactor)

	initialTx := result.Tx
	gasForFullPrice := uint64(minGasLimit + gasPerDataByte*len(initialTx.GetData()))
	if guardedTx {
		gasForFullPrice += extraGasLimitForGuarded
	}

	if common.IsRelayedTxV3(initialTx) {
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
