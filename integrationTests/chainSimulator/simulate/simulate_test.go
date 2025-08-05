package simulate

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../cmd/node/config/"
)

var (
	oneEGLD = big.NewInt(1000000000000000000)
)

func TestCostScDeploy(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    20,
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
		AlterConfigsFunction: func(cfg *config.Configs) {

		},
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	receiver := "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu"
	receiverBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(receiver)

	scCode := wasm.GetSCCode("../../vm/wasm/testdata/misc/fib_wasm/output/fib_wasm.wasm")
	txDeployData := []byte(wasm.CreateDeployTxData(scCode))
	tx := generateTransaction(sender.Bytes, 0, receiverBytes, big.NewInt(0), string(txDeployData), 0)

	cost, err := cs.GetNodeHandler(0).GetFacadeHandler().ComputeTransactionGasLimit(tx)
	require.NoError(t, err)
	require.Equal(t, uint64(2_506_706), cost.GasUnits)
	require.Equal(t, "", cost.ReturnMessage)
}

func TestSimulateIntraShardTxWithGuardian(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    20,
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
		AlterConfigsFunction: func(cfg *config.Configs) {

		},
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))

	sender, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	receiver, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	guardian, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	senderNonce := uint64(0)
	// Set guardian for sender
	setGuardianTxData := "SetGuardian@" + hex.EncodeToString(guardian.Bytes) + "@" + hex.EncodeToString([]byte("uuid"))
	setGuardianGasLimit := 10_000_000
	setGuardianTx := generateTransaction(sender.Bytes, senderNonce, sender.Bytes, big.NewInt(0), setGuardianTxData, uint64(setGuardianGasLimit))
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(setGuardianTx, 10)
	require.NoError(t, err)
	senderNonce++

	// fast-forward until the guardian becomes active
	for i := 0; i < 21; i++ {
		err = cs.ForceChangeOfEpoch()
		require.NoError(t, err)
	}

	// guard account
	guardAccountTxData := "GuardAccount"
	guardAccountGasLimit := 10_000_000
	guardAccountTx := generateTransaction(sender.Bytes, senderNonce, sender.Bytes, big.NewInt(0), guardAccountTxData, uint64(guardAccountGasLimit))
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(guardAccountTx, 10)
	require.NoError(t, err)
	senderNonce++

	esdtTransferTx := generateTransaction(sender.Bytes, senderNonce, receiver.Bytes, big.NewInt(0), "ESDTTransfer@010101@01@01", 500_000)
	cost, err := cs.GetNodeHandler(0).GetFacadeHandler().ComputeTransactionGasLimit(esdtTransferTx)
	require.NoError(t, err)
	require.Equal(t, "transaction is not executable and gas will not be consumed, not allowed to bypass guardian: ", cost.ReturnMessage)

	esdtTransferTx.GuardianAddr = guardian.Bytes
	esdtTransferTx.Options = 2
	esdtTransferTx.Version = 2
	esdtTransferTx.GuardianSignature = []byte("aaa")
	cost, err = cs.GetNodeHandler(0).GetFacadeHandler().ComputeTransactionGasLimit(esdtTransferTx)
	require.NoError(t, err)
	require.Equal(t, "failed transaction, gas consumed: insufficient funds", cost.ReturnMessage)
}

func TestRelayedV3(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpochOpt := core.OptionalUint64{
		HasValue: true,
		Value:    20,
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
		AlterConfigsFunction: func(cfg *config.Configs) {

		},
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	for idx := 0; idx < 4; idx++ {
		err = cs.ForceChangeOfEpoch()
		require.NoError(t, err)
	}

	initialBalance := big.NewInt(0).Mul(oneEGLD, big.NewInt(10))

	sender, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: sender.Bech32,
			Balance: "0",
			Pairs: map[string]string{
				"454c524f4e446573647453484f572d633961633237": "12040002d820",
			},
		},
	})
	require.NoError(t, err)

	receiver, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	relayer, err := cs.GenerateAndMintWalletAddress(0, initialBalance)
	require.NoError(t, err)

	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	dataTx := "ESDTTransfer@53484f572d633961633237@3e80@5061796d656e7420746f20504f5334@317820564f444b41204d495820323530204d4c20782033322e3030202b20317820574849534b59204d495820323530204d4c20782033322e3030202b203178204a4147455220434f4c4120323530204d4c20782033322e3030202b2031782047494e20544f4e494320323530204d4c20782033322e3030202b2031782043554241204c49425245203235304d4c20782033322e3030"
	tx := generateTransaction(sender.Bytes, 0, receiver.Bytes, big.NewInt(0), dataTx, 0)
	tx.RelayerAddr = relayer.Bytes

	cost, err := cs.GetNodeHandler(0).GetFacadeHandler().ComputeTransactionGasLimit(tx)
	require.NoError(t, err)
	require.Equal(t, uint64(855001), cost.GasUnits)
	require.Equal(t, "", cost.ReturnMessage)
}

func generateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(data),
		GasLimit:  gasLimit,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
		Signature: []byte("ssig"),
	}
}
