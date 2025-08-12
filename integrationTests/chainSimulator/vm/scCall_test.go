package vm

import (
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	api2 "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

func TestMoveBalance(t *testing.T) {
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
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	pubKeyConverter := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()
	sndBech := "erd1v3undjd5mrvqq3tg3re8u9z55tc3ysxaw7ladqrvq9m69r4tv0yqrkfryq"
	snd, _ := pubKeyConverter.Decode(sndBech)
	rcv, _ := pubKeyConverter.Decode("erd1vj3efd5czwearu0gr3vjct8ef53lvtl7vs42vts2kh2qn3cucrnsj7ymqx")

	value := big.NewInt(0).Mul(oneEGLD, big.NewInt(100))

	tx := &transaction.Transaction{
		Nonce:     35,
		Value:     value,
		SndAddr:   snd,
		RcvAddr:   rcv,
		Data:      []byte(""),
		GasLimit:  50_000,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
		Signature: []byte("dummy"),
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	sndShardID := cs.GetNodeHandler(core.MetachainShardId).GetShardCoordinator().ComputeId(snd)
	senderAccount, _, err := cs.GetNodeHandler(sndShardID).GetFacadeHandler().GetAccount(sndBech, api2.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, uint64(36), senderAccount.Nonce)
	fmt.Println(senderAccount.Balance)
}

func TestSCCallMainnetState(t *testing.T) {
	t.Skip()

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
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	pubKeyConverter := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()
	snd, _ := pubKeyConverter.Decode("erd1khl4mll00m6tgvcpmhf9tu05s9cfszruw77fk4dfmzyn7r2h6kgqdtsj7u")
	rcv, _ := pubKeyConverter.Decode("erd1qqqqqqqqqqqqqpgqd9azvcr2a9c953a2gfdm5f7ty49ycqa7gmus63aepv")

	tx := &transaction.Transaction{
		Nonce:     839,
		Value:     big.NewInt(0),
		SndAddr:   snd,
		RcvAddr:   rcv,
		Data:      []byte("cancelGame@0b7b"),
		GasLimit:  7_000_000,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
		Signature: []byte("dummy"),
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
}

func TestSCCallSwapMainnetState(t *testing.T) {
	t.Skip()

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
		FetchStateConfig: chainSimulator.FetchStateFromProvidedChainConfig{
			Enabled:    true,
			FetchNonce: true,
			FetchRound: true,
			FetchEpoch: true,
			GatewayURL: "https://gateway.multiversx.com",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, cs)

	err = cs.GenerateBlocks(1)
	require.NoError(t, err)

	pubKeyConverter := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter()
	sndBech := "erd1djem6d5st695x36kqzcgeatxahs0fhyy7c5j5p4neaj997yufvcs6cskhq"
	snd, _ := pubKeyConverter.Decode(sndBech)
	rcv, _ := pubKeyConverter.Decode("erd1qqqqqqqqqqqqqpgqlt8sksgnhk98pm2chnjwhz5cat7s5wy72jpsgdrmac")

	senderNonce, err := cs.ChainFetcher().GetAddressNonce(sndBech)
	require.NoError(t, err)

	tx := &transaction.Transaction{
		Nonce:     senderNonce,
		Value:     big.NewInt(0),
		SndAddr:   snd,
		RcvAddr:   rcv,
		Data:      []byte("ESDTTransfer@55544b2d326638306539@97add92ed868e27d34@73776170546f6b656e734669786564496e707574@5553482d313131653039@01"),
		GasLimit:  30_000_000,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
		Signature: []byte("dummy"),
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())
}
