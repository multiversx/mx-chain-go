package relayedTxSet1

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	relayedTx2 "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/relayedTx"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
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

			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 0
			cfg.RoundConfig.RoundActivations = map[string]config.ActivationRoundByName{
				"DisableAsyncCallV1": {
					Round: "9999999",
				},
				"SupernovaEnableRound": {
					Round: "0",
				},
			}
		}

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
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
			setGuardianGasLimit := relayedTx2.MinGasLimit + 1500*len(setGuardianTxData) + relayedTx2.GuardAccountCost
			setGuardianTx := relayedTx2.GenerateTransaction(sender.Bytes, senderNonce, sender.Bytes, big.NewInt(0), setGuardianTxData, uint64(setGuardianGasLimit))
			_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(setGuardianTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
			require.NoError(t, err)
			senderNonce++

			// fast-forward until the guardian becomes active
			err = cs.GenerateBlocks(relayedTx2.RoundsPerEpoch * 20)
			require.NoError(t, err)

			// guard account
			guardAccountTxData := "GuardAccount"
			guardAccountGasLimit := relayedTx2.MinGasLimit + 1500*len(guardAccountTxData) + relayedTx2.GuardAccountCost
			guardAccountTx := relayedTx2.GenerateTransaction(sender.Bytes, senderNonce, sender.Bytes, big.NewInt(0), guardAccountTxData, uint64(guardAccountGasLimit))
			_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(guardAccountTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
			require.NoError(t, err)
			senderNonce++
		}

		senderBalanceBefore := relayedTx2.GetBalance(t, cs, sender)

		gasLimit := relayedTx2.MinGasLimit * 2
		extraGasLimit := 0
		if extraGas {
			extraGasLimit = relayedTx2.MinGasLimit
		}
		if guardedTx {
			gasLimit += relayedTx2.ExtraGasLimitForGuarded
		}
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, senderNonce, receiver.Bytes, relayer.Bytes, relayedTx2.OneEGLD, "", uint64(gasLimit+extraGasLimit))

		if guardedTx {
			relayedTx.GuardianAddr = guardian.Bytes
			relayedTx.GuardianSignature = []byte(relayedTx2.MockTxSignature)
			relayedTx.Options = 2
		}

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		isIntraShard := relayerShard == destinationShard
		if isIntraShard {
			require.NoError(t, cs.GenerateBlocks(relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx))
		}

		// check fee fields
		initiallyPaidFee, fee, gasUsed := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(result, big.NewInt(0), true, guardedTx)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := relayedTx2.GetBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance
		senderBalanceAfter := relayedTx2.GetBalance(t, cs, sender)
		senderBalanceDiff := big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter)
		require.Equal(t, relayedTx2.OneEGLD.String(), senderBalanceDiff.String())

		// check receiver balance
		receiverBalanceAfter := relayedTx2.GetBalance(t, cs, receiver)
		require.Equal(t, relayedTx2.OneEGLD.String(), receiverBalanceAfter.String())

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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		receiver, err := cs.GenerateAndMintWalletAddress(receiverShard, big.NewInt(0))
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		gasLimit := relayedTx2.MinGasLimit * 2
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, 0, receiver.Bytes, relayer.Bytes, relayedTx2.OneEGLD, "", uint64(gasLimit))

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// send same tx again, lower nonce
		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		receiver, err := cs.GenerateAndMintWalletAddress(receiverShard, big.NewInt(0))
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		gasLimit := relayedTx2.MinGasLimit
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, 0, receiver.Bytes, relayer.Bytes, relayedTx2.OneEGLD, "", uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
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

		resultDeploy, scAddressBytes := relayedTx2.DeployAdder(t, cs, owner, 0)
		refundDeploy := relayedTx2.GetRefundValue(resultDeploy.SmartContractResults)

		// send relayed tx
		txDataAdd := "add@" + hex.EncodeToString(big.NewInt(1).Bytes())
		gasLimit := uint64(2000000)
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, 0, scAddressBytes, relayer.Bytes, big.NewInt(0), txDataAdd, gasLimit)

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// if cross shard, generate few more blocks for eventual refunds to be executed
		if relayerShard != ownerShard {
			require.NoError(t, cs.GenerateBlocks(relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx))
		}

		relayedTx2.CheckSum(t, cs.GetNodeHandler(ownerShard), scAddressBytes, owner.Bytes, 1)

		refundValue := relayedTx2.GetRefundValue(result.SmartContractResults)
		require.NotZero(t, refundValue.Uint64())

		// check fee fields
		initiallyPaidFee, fee, gasUsed := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := relayedTx2.GetBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(relayerInitialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance, only if the tx was not relayed by sender
		if !relayedBySender {
			senderBalanceAfter := relayedTx2.GetBalance(t, cs, sender)
			require.Equal(t, senderInitialBalance.String(), senderBalanceAfter.String())
		}

		// check owner balance
		_, feeDeploy, _ := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(resultDeploy, refundDeploy, false, false)
		ownerBalanceAfter := relayedTx2.GetBalance(t, cs, owner)
		ownerFee := big.NewInt(0).Sub(initialBalance, ownerBalanceAfter)
		require.Equal(t, feeDeploy.String(), ownerFee.String())

		// check scrs
		require.Equal(t, 1, len(result.SmartContractResults))
		for _, scr := range result.SmartContractResults {
			relayedTx2.CheckSCRSucceeded(t, cs, scr)
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))

		sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		senderNonce := uint64(0)
		senderBalanceBefore := relayedTx2.GetBalance(t, cs, sender)

		gasLimit := relayedTx2.MinGasLimit * 2
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, senderNonce, sender.Bytes, sender.Bytes, big.NewInt(0), "", uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check fee fields
		initiallyPaidFee, fee, gasUsed := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(result, big.NewInt(0), true, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check sender balance
		expectedFee := core.SafeMul(uint64(gasLimit), uint64(relayedTx2.MinGasPrice))
		senderBalanceAfter := relayedTx2.GetBalance(t, cs, sender)
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))

		sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		receiver, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		senderNonce := uint64(0)
		receiverBalanceBefore := relayedTx2.GetBalance(t, cs, receiver)

		gasLimit := relayedTx2.MinGasLimit * 2
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, senderNonce, receiver.Bytes, receiver.Bytes, big.NewInt(0), "", uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check fee fields
		initiallyPaidFee, fee, gasUsed := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(result, big.NewInt(0), true, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check sender balance
		senderBalanceAfter := relayedTx2.GetBalance(t, cs, sender)
		require.Equal(t, senderBalanceAfter.String(), initialBalance.String())

		// check receiver balance
		expectedFee := core.SafeMul(uint64(gasLimit), uint64(relayedTx2.MinGasPrice))
		receiverBalanceAfter := relayedTx2.GetBalance(t, cs, receiver)
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10000))

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
			gasLimit := relayedTx2.MinGasLimit + len(txDataTransfer)*1500 + relayedTx2.ExtraGasESDTTransfer
			ownerNonce := relayedTx2.GetNonce(t, cs, owner)

			esdtTransferTx := relayedTx2.GenerateTransaction(owner.Bytes, ownerNonce, sender.Bytes, big.NewInt(0), txDataTransfer, uint64(gasLimit))

			_, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(esdtTransferTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
			require.NoError(t, err)

			esdtBalanceSender := relayedTx2.GetESDTBalance(t, cs, sender, 0, ticker)
			require.Equal(t, transferValue.String(), esdtBalanceSender.String())
		}

		senderBalanceBefore := relayedTx2.GetBalance(t, cs, sender)

		// send relayed tx
		gasLimit := relayedTx2.MinGasLimit*2 + len(txDataTransfer)*1500 + relayedTx2.ExtraGasESDTTransfer
		senderNonce := relayedTx2.GetNonce(t, cs, sender)
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, senderNonce, receiver.Bytes, relayer.Bytes, big.NewInt(0), txDataTransfer, uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check sender balance
		senderBalanceAfter := relayedTx2.GetBalance(t, cs, sender)
		require.Equal(t, senderBalanceAfter.String(), senderBalanceBefore.String())

		refundValue := relayedTx2.GetRefundValue(result.SmartContractResults)
		require.NotZero(t, refundValue.Uint64())

		// check fee fields
		initiallyPaidFee, fee, gasUsed := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := relayedTx2.GetBalance(t, cs, relayer)
		relayerBalanceDiff := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerBalanceDiff.String())

		// check receiver esdt balance
		expectedESDTBalanceSender := big.NewInt(0)
		if senderIsIssuer {
			expectedESDTBalanceSender = big.NewInt(0).Sub(initialSupply, transferValue)
		}
		esdtBalanceSender := relayedTx2.GetESDTBalance(t, cs, sender, 0, ticker)
		require.Equal(t, expectedESDTBalanceSender.String(), esdtBalanceSender.String())

		esdtBalanceReceiver := relayedTx2.GetESDTBalance(t, cs, receiver, 0, ticker)
		require.Equal(t, transferValue.String(), esdtBalanceReceiver.String())

		// check receiver egld balance unchanged if tx is relayed by third party
		if !relayedByReceiver {
			receiverBalanceAfter := relayedTx2.GetBalance(t, cs, receiver)
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10000))

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
		egldTransferValue := relayedTx2.OneEGLD

		// if sender is not owner, send some new tokens to the sender
		if !senderIsIssuer {
			txDataTransfer := "MultiESDTNFTTransfer@" +
				hex.EncodeToString(sender.Bytes) + "@02@" +
				hex.EncodeToString([]byte(ticker)) + "@00@" + hex.EncodeToString(transferValue.Bytes()) + "@" +
				hex.EncodeToString([]byte(relayedTx2.EgldTicker)) + "@00@" + hex.EncodeToString(egldTransferValue.Bytes())
			gasLimit := relayedTx2.MinGasLimit + len(txDataTransfer)*1500 + relayedTx2.ExtraGasMultiESDTTransfer
			ownerNonce := relayedTx2.GetNonce(t, cs, owner)

			esdtTransferTx := relayedTx2.GenerateTransaction(owner.Bytes, ownerNonce, owner.Bytes, big.NewInt(0), txDataTransfer, uint64(gasLimit))

			_, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(esdtTransferTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
			require.NoError(t, err)

			esdtBalanceSender := relayedTx2.GetESDTBalance(t, cs, sender, 0, ticker)
			require.Equal(t, transferValue.String(), esdtBalanceSender.String())

			balanceSender := relayedTx2.GetBalance(t, cs, sender)
			require.Equal(t, egldTransferValue.String(), balanceSender.String())
		}

		senderBalanceBefore := relayedTx2.GetBalance(t, cs, sender)

		// send relayed tx
		txDataTransfer := "MultiESDTNFTTransfer@" +
			hex.EncodeToString(receiver.Bytes) + "@02@" +
			hex.EncodeToString([]byte(ticker)) + "@00@" + hex.EncodeToString(transferValue.Bytes()) + "@" +
			hex.EncodeToString([]byte(relayedTx2.EgldTicker)) + "@00@" + hex.EncodeToString(egldTransferValue.Bytes())
		gasLimit := relayedTx2.MinGasLimit*2 + len(txDataTransfer)*1500 + relayedTx2.ExtraGasMultiESDTTransfer
		senderNonce := relayedTx2.GetNonce(t, cs, sender)
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, senderNonce, sender.Bytes, relayer.Bytes, big.NewInt(0), txDataTransfer, uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// check sender balance
		senderBalanceAfter := relayedTx2.GetBalance(t, cs, sender)
		expectedSenderBalanceDiff := egldTransferValue
		senderBalanceDiff := big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter)
		require.Equal(t, expectedSenderBalanceDiff.String(), senderBalanceDiff.String())

		refundValue := relayedTx2.GetRefundValue(result.SmartContractResults)
		require.NotZero(t, refundValue.Uint64())

		// check fee fields
		initiallyPaidFee, fee, gasUsed := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := relayedTx2.GetBalance(t, cs, relayer)
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
		esdtBalanceSender := relayedTx2.GetESDTBalance(t, cs, sender, 0, ticker)
		require.Equal(t, expectedESDTBalanceSender.String(), esdtBalanceSender.String())

		esdtBalanceReceiver := relayedTx2.GetESDTBalance(t, cs, receiver, 0, ticker)
		require.Equal(t, transferValue.String(), esdtBalanceReceiver.String())

		// check receiver egld balance
		if !relayedByReceiver {
			receiverBalanceAfter := relayedTx2.GetBalance(t, cs, receiver)
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

	ownerNonce := relayedTx2.GetNonce(t, cs, owner)
	fiveEGLD := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(5))
	issueTx := relayedTx2.GenerateTransaction(owner.Bytes, ownerNonce, vm.ESDTSCAddress, fiveEGLD, issueTxData, 60000000)
	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(issueTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	// generate few more blocks for refund to happen
	_ = cs.GenerateBlocks(relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)

	require.NotNil(t, result.Logs)
	require.NotZero(t, len(result.Logs.Events))
	for _, log := range result.Logs.Events {
		if log.Identifier != "issue" {
			continue
		}

		ticker := log.Topics[0]
		balance := relayedTx2.GetESDTBalance(t, cs, owner, 0, string(ticker))
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		owner, err := cs.GenerateAndMintWalletAddress(ownerShard, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		_, scAddressBytes := relayedTx2.DeployAdder(t, cs, owner, 0)

		// send relayed tx with less gas limit
		txDataAdd := "add@" + hex.EncodeToString(big.NewInt(1).Bytes())
		gasLimit := relayedTx2.GasPerDataByte*len(txDataAdd) + relayedTx2.MinGasLimit + relayedTx2.MinGasLimit
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, 0, scAddressBytes, relayer.Bytes, big.NewInt(0), txDataAdd, uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
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

		refundValue := relayedTx2.GetRefundValue(result.SmartContractResults)
		require.Zero(t, refundValue.Uint64())

		// check fee fields, should consume full gas
		initiallyPaidFee, fee, gasUsed := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)
		require.Equal(t, fee.String(), result.Fee)
		require.Equal(t, result.InitiallyPaidFee, result.Fee)
		require.Equal(t, gasUsed, result.GasUsed)
		require.Equal(t, relayedTx.GasLimit, result.GasUsed)

		// check relayer balance
		relayerBalanceAfter := relayedTx2.GetBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance
		senderBalanceAfter := relayedTx2.GetBalance(t, cs, sender)
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		sender, err := cs.GenerateAndMintWalletAddress(relayerShard, initialBalance)
		require.NoError(t, err)

		owner, err := cs.GenerateAndMintWalletAddress(ownerShard, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		_, scAddressBytes := relayedTx2.DeployAdder(t, cs, owner, 0)

		// send relayed tx with invalid value
		txDataAdd := "invalid"
		gasLimit := uint64(3000000)
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, 0, scAddressBytes, relayer.Bytes, big.NewInt(0), txDataAdd, uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
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

		refundValue := relayedTx2.GetRefundValue(result.SmartContractResults)
		require.Zero(t, refundValue.Uint64()) // no refund, tx failed

		// check fee fields, should consume full gas
		initiallyPaidFee, fee, _ := relayedTx2.ComputeTxGasAndFeeBasedOnRefund(result, refundValue, false, false)
		require.Equal(t, initiallyPaidFee.String(), result.InitiallyPaidFee)

		// check relayer balance
		relayerBalanceAfter := relayedTx2.GetBalance(t, cs, relayer)
		relayerFee := big.NewInt(0).Sub(initialBalance, relayerBalanceAfter)
		require.Equal(t, fee.String(), relayerFee.String())

		// check sender balance
		senderBalanceAfter := relayedTx2.GetBalance(t, cs, sender)
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

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		relayerShard := uint32(0)

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10000))
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
		value := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(1250))
		relayedTx := relayedTx2.GenerateRelayedV3Transaction(sender.Bytes, 0, vm.DelegationManagerSCAddress, relayer.Bytes, value, txData, gasLimit)

		relayerBefore := relayedTx2.GetBalance(t, cs, relayer)
		senderBefore := relayedTx2.GetBalance(t, cs, sender)

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		require.NoError(t, cs.GenerateBlocks(relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx))

		relayerAfter := relayedTx2.GetBalance(t, cs, relayer)
		senderAfter := relayedTx2.GetBalance(t, cs, sender)

		// check consumed fees
		refund := relayedTx2.GetRefundValue(result.SmartContractResults)
		consumedFee := big.NewInt(0).Sub(relayerBefore, relayerAfter)

		gasForFullPrice := uint64(len(txData)*relayedTx2.GasPerDataByte + relayedTx2.MinGasLimit + relayedTx2.MinGasLimit)
		gasForDeductedPrice := gasLimit - gasForFullPrice
		deductedGasPrice := uint64(relayedTx2.MinGasPrice / relayedTx2.DeductionFactor)
		initialFee := gasForFullPrice*relayedTx2.MinGasPrice + gasForDeductedPrice*deductedGasPrice
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
