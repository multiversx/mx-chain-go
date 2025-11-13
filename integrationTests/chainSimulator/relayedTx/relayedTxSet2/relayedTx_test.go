package relayedTxSet2

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	relayedTx2 "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/relayedTx"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

func TestFixRelayedMoveBalanceWithChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	expectedFeeScCallBefore := "827295420000000"
	expectedFeeScCallAfter := "885705420000000"
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
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV1V2DisableEpoch = integrationTests.UnreachableEpoch
		}

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
		relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// deploy adder contract
		owner, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
		require.NoError(t, err)

		// generate one block so the minting has effect
		err = cs.GenerateBlocks(1)
		require.NoError(t, err)

		ownerNonce := uint64(0)
		_, scAddressBytes := relayedTx2.DeployAdder(t, cs, owner, ownerNonce)

		// fast-forward until epoch 4
		err = cs.GenerateBlocksUntilEpochIsReached(int32(4))
		require.NoError(t, err)

		// send relayed tx
		txDataAdd := "add@" + hex.EncodeToString(big.NewInt(1).Bytes())
		innerTx := relayedTx2.GenerateTransaction(owner.Bytes, 1, scAddressBytes, big.NewInt(0), txDataAdd, 3000000)
		marshalledTx, err := json.Marshal(innerTx)
		require.NoError(t, err)
		txData := []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit := 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx := relayedTx2.GenerateTransaction(relayer.Bytes, 0, owner.Bytes, big.NewInt(0), string(txData), gasLimit)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		// send relayed tx, fix still not active
		innerTx = relayedTx2.GenerateTransaction(owner.Bytes, 2, scAddressBytes, big.NewInt(0), txDataAdd, 1230000)
		marshalledTx, err = json.Marshal(innerTx)
		require.NoError(t, err)
		txData = []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit = 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx = relayedTx2.GenerateTransaction(relayer.Bytes, 1, owner.Bytes, big.NewInt(0), string(txData), gasLimit)

		relayerBalanceBefore := relayedTx2.GetBalance(t, cs, relayer)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)
		relayerBalanceAfter := relayedTx2.GetBalance(t, cs, relayer)

		feeConsumed := big.NewInt(0).Sub(relayerBalanceBefore, relayerBalanceAfter)

		require.Equal(t, expectedFeeBeforeFix, feeConsumed.String())

		// fast-forward until the fix is active
		err = cs.GenerateBlocksUntilEpochIsReached(int32(providedActivationEpoch))
		require.NoError(t, err)

		// send relayed tx after fix
		innerTx = relayedTx2.GenerateTransaction(owner.Bytes, 3, scAddressBytes, big.NewInt(0), txDataAdd, 1500000)
		marshalledTx, err = json.Marshal(innerTx)
		require.NoError(t, err)
		txData = []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit = 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx = relayedTx2.GenerateTransaction(relayer.Bytes, 2, owner.Bytes, big.NewInt(0), string(txData), gasLimit)

		relayerBalanceBefore = relayedTx2.GetBalance(t, cs, relayer)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		relayerBalanceAfter = relayedTx2.GetBalance(t, cs, relayer)

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
			cfg.EpochConfig.EnableEpochs.RelayedTransactionsV1V2DisableEpoch = integrationTests.UnreachableEpoch
		}

		cs := relayedTx2.StartChainSimulator(t, alterConfigsFunc)
		defer cs.Close()

		initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
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
		innerTx := relayedTx2.GenerateTransaction(sender.Bytes, 0, receiver.Bytes, relayedTx2.OneEGLD, "", 50000)
		marshalledTx, err := json.Marshal(innerTx)
		require.NoError(t, err)
		txData := []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit := 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx := relayedTx2.GenerateTransaction(relayer.Bytes, 0, sender.Bytes, big.NewInt(0), string(txData), gasLimit)

		relayerBalanceBefore := relayedTx2.GetBalance(t, cs, relayer)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		relayerBalanceAfter := relayedTx2.GetBalance(t, cs, relayer)

		feeConsumed := big.NewInt(0).Sub(relayerBalanceBefore, relayerBalanceAfter)

		require.Equal(t, expectedFeeBeforeFix, feeConsumed.String())

		// fast-forward until the fix is active
		err = cs.GenerateBlocksUntilEpochIsReached(int32(providedActivationEpoch))
		require.NoError(t, err)

		// send relayed tx
		innerTx = relayedTx2.GenerateTransaction(sender.Bytes, 1, receiver.Bytes, relayedTx2.OneEGLD, "", 50000)
		marshalledTx, err = json.Marshal(innerTx)
		require.NoError(t, err)
		txData = []byte("relayedTx@" + hex.EncodeToString(marshalledTx))
		gasLimit = 50000 + uint64(len(txData))*1500 + innerTx.GasLimit

		relayedTx = relayedTx2.GenerateTransaction(relayer.Bytes, 1, sender.Bytes, big.NewInt(0), string(txData), gasLimit)

		relayerBalanceBefore = relayedTx2.GetBalance(t, cs, relayer)

		_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		relayerBalanceAfter = relayedTx2.GetBalance(t, cs, relayer)

		feeConsumed = big.NewInt(0).Sub(relayerBalanceBefore, relayerBalanceAfter)

		require.Equal(t, expectedFeeAfterFix, feeConsumed.String())
	}
}

func TestRelayedTransactionFeeField(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := relayedTx2.StartChainSimulator(t, func(cfg *config.Configs) {
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsEnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.FixRelayedBaseCostEnableEpoch = 1
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsV1V2DisableEpoch = integrationTests.UnreachableEpoch
	})
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	receiver, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	t.Run("relayed v1", func(t *testing.T) {
		innerTx := relayedTx2.GenerateTransaction(sender.Bytes, 0, receiver.Bytes, relayedTx2.OneEGLD, "", relayedTx2.MinGasLimit)
		buff, err := json.Marshal(innerTx)
		require.NoError(t, err)

		txData := []byte("relayedTx@" + hex.EncodeToString(buff))
		gasLimit := relayedTx2.MinGasLimit + len(txData)*relayedTx2.GasPerDataByte + int(innerTx.GasLimit)
		relayedTx := relayedTx2.GenerateTransaction(relayer.Bytes, 0, sender.Bytes, big.NewInt(0), string(txData), uint64(gasLimit))

		result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
		require.NoError(t, err)

		expectedFee := core.SafeMul(uint64(gasLimit), relayedTx2.MinGasPrice)
		require.Equal(t, expectedFee.String(), result.Fee)
		require.Equal(t, expectedFee.String(), result.InitiallyPaidFee)
		require.Equal(t, uint64(gasLimit), result.GasUsed)
	})
}

func TestRelayedTxCheckTxProcessingTypeAfterRelayedV1Disabled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	const RelayedV1DisableEpoch = 3

	cs := relayedTx2.StartChainSimulator(t, func(cfg *config.Configs) {
		cfg.EpochConfig.EnableEpochs.RelayedTransactionsV1V2DisableEpoch = RelayedV1DisableEpoch
	})
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))
	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	snd, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	rcv, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	userTx := relayedTx2.GenerateTransaction(snd.Bytes, 0, rcv.Bytes, relayedTx2.OneEGLD, "d", relayedTx2.MinGasLimit+relayedTx2.GasPerDataByte)
	buff, err := json.Marshal(userTx)
	require.NoError(t, err)

	txData := []byte("relayedTx@" + hex.EncodeToString(buff))
	gasLimit := relayedTx2.MinGasLimit + len(txData)*relayedTx2.GasPerDataByte + int(userTx.GasLimit)
	relayedTx := relayedTx2.GenerateTransaction(relayer.Bytes, 0, snd.Bytes, big.NewInt(0), string(txData), uint64(gasLimit))

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(relayedTx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)
	require.Equal(t, process.RelayedTx.String(), result.ProcessingTypeOnSource)
	require.Equal(t, process.RelayedTx.String(), result.ProcessingTypeOnSource)

	err = cs.GenerateBlocksUntilEpochIsReached(RelayedV1DisableEpoch)
	require.NoError(t, err)

	result, err = cs.GetNodeHandler(0).GetFacadeHandler().GetTransaction(result.Hash, true)
	require.NoError(t, err)
	require.Equal(t, process.RelayedTx.String(), result.ProcessingTypeOnSource)
	require.Equal(t, process.RelayedTx.String(), result.ProcessingTypeOnSource)
}

func TestRegularMoveBalanceWithRefundReceipt(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := relayedTx2.StartChainSimulator(t, func(cfg *config.Configs) {})
	defer cs.Close()

	initialBalance := big.NewInt(0).Mul(relayedTx2.OneEGLD, big.NewInt(10))

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	// generate one block so the minting has effect
	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	senderNonce := uint64(0)

	extraGas := uint64(relayedTx2.MinGasLimit)
	gasLimit := relayedTx2.MinGasLimit + extraGas
	tx := relayedTx2.GenerateTransaction(sender.Bytes, senderNonce, sender.Bytes, relayedTx2.OneEGLD, "", gasLimit)

	result, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, relayedTx2.MaxNumOfBlocksToGenerateWhenExecutingTx)
	require.NoError(t, err)

	require.NotNil(t, result.Receipt)
	expectedGasRefunded := core.SafeMul(extraGas, relayedTx2.MinGasPrice/relayedTx2.DeductionFactor)
	require.Equal(t, expectedGasRefunded.String(), result.Receipt.Value.String())
}
