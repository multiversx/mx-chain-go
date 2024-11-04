package relayedTx

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
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
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig              = "../../../cmd/node/config/"
	minGasPrice                             = 1_000_000_000
	minGasLimit                             = 50_000
	gasPerDataByte                          = 1_500
	txVersion                               = 2
	mockTxSignature                         = "sig"
	maxNumOfBlocksToGenerateWhenExecutingTx = 10
	roundsPerEpoch                          = 30
)

var (
	oneEGLD = big.NewInt(1000000000000000000)
)

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
