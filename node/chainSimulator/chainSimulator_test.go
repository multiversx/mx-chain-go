package chainSimulator

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorCommon "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

const (
	defaultPathToInitialConfig = "../../cmd/node/config/"
)

func TestNewChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 3,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              core.OptionalUint64{},
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	time.Sleep(time.Second)

	chainSimulator.Close()
}

func TestChainSimulator_GenerateBlocksShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    20,
		},
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		InitialRound:                200000000,
		InitialEpoch:                100,
		InitialNonce:                100,
		AlterConfigsFunction: func(cfg *config.Configs) {
			// we need to enable this as this test skips a lot of epoch activations events, and it will fail otherwise
			// because the owner of a BLS key coming from genesis is not set
			// (the owner is not set at genesis anymore because we do not enable the staking v2 in that phase)
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(50)
	require.Nil(t, err)
}

func TestChainSimulator_GenerateBlocksAndEpochChangeShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 3,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            100,
		MetaChainMinNodes:           100,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	facade, err := NewChainSimulatorFacade(chainSimulator)
	require.Nil(t, err)

	genesisBalances := make(map[string]*big.Int)
	for _, stakeWallet := range chainSimulator.initialWalletKeys.StakeWallets {
		initialAccount, errGet := facade.GetExistingAccountFromBech32AddressString(stakeWallet.Address.Bech32)
		require.Nil(t, errGet)

		genesisBalances[stakeWallet.Address.Bech32] = initialAccount.GetBalance()
	}

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(80)
	require.Nil(t, err)

	numAccountsWithIncreasedBalances := 0
	for _, stakeWallet := range chainSimulator.initialWalletKeys.StakeWallets {
		account, errGet := facade.GetExistingAccountFromBech32AddressString(stakeWallet.Address.Bech32)
		require.Nil(t, errGet)

		if account.GetBalance().Cmp(genesisBalances[stakeWallet.Address.Bech32]) > 0 {
			numAccountsWithIncreasedBalances++
		}
	}

	assert.True(t, numAccountsWithIncreasedBalances > 0)
}

func TestChainSimulator_SetState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 3,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	chainSimulatorCommon.CheckSetState(t, chainSimulator, chainSimulator.GetNodeHandler(0))
}

func TestChainSimulator_SetEntireState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 3,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	chainSimulatorCommon.CheckSetEntireState(t, chainSimulator, chainSimulator.GetNodeHandler(1))
}

func TestChainSimulator_SetEntireStateWithRemoval(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 3,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	chainSimulatorCommon.CheckSetEntireStateWithRemoval(t, chainSimulator, chainSimulator.GetNodeHandler(1))
}

func TestChainSimulator_GetAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 3,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	chainSimulatorCommon.CheckGetAccount(t, chainSimulator)
}

func TestSimulator_SendTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 3,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            1,
		MetaChainMinNodes:           1,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	oneEgld := big.NewInt(1000000000000000000)
	initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(100))
	transferValue := big.NewInt(0).Mul(oneEgld, big.NewInt(5))

	wallet0, err := chainSimulator.GenerateAndMintWalletAddress(0, initialMinting)
	require.Nil(t, err)

	wallet1, err := chainSimulator.GenerateAndMintWalletAddress(1, initialMinting)
	require.Nil(t, err)

	wallet2, err := chainSimulator.GenerateAndMintWalletAddress(2, initialMinting)
	require.Nil(t, err)

	wallet3, err := chainSimulator.GenerateAndMintWalletAddress(2, initialMinting)
	require.Nil(t, err)

	wallet4, err := chainSimulator.GenerateAndMintWalletAddress(2, initialMinting)
	require.Nil(t, err)

	gasLimit := uint64(50000)
	tx0 := chainSimulatorCommon.GenerateTransaction(wallet0.Bytes, 0, wallet2.Bytes, transferValue, "", gasLimit)
	tx1 := chainSimulatorCommon.GenerateTransaction(wallet1.Bytes, 0, wallet2.Bytes, transferValue, "", gasLimit)
	tx3 := chainSimulatorCommon.GenerateTransaction(wallet3.Bytes, 0, wallet4.Bytes, transferValue, "", gasLimit)

	maxNumOfBlockToGenerateWhenExecutingTx := 15

	t.Run("nil or empty slice of transactions should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted(nil, 1)
		assert.Equal(t, ErrEmptySliceOfTxs, errSend)
		assert.Nil(t, sentTxs)

		sentTxs, errSend = chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted(make([]*transaction.Transaction, 0), 1)
		assert.Equal(t, ErrEmptySliceOfTxs, errSend)
		assert.Nil(t, sentTxs)
	})
	t.Run("invalid max number of blocks to generate should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx0, tx1}, 0)
		assert.Equal(t, ErrInvalidMaxNumOfBlocks, errSend)
		assert.Nil(t, sentTxs)
	})
	t.Run("nil transaction in slice should error", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{nil}, 1)
		assert.ErrorIs(t, errSend, ErrNilTransaction)
		assert.Nil(t, sentTxs)
	})
	t.Run("2 transactions from different shard should call send correctly", func(t *testing.T) {
		sentTxs, errSend := chainSimulator.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx0, tx1}, maxNumOfBlockToGenerateWhenExecutingTx)
		assert.Equal(t, 2, len(sentTxs))
		assert.Nil(t, errSend)

		account, errGet := chainSimulator.GetAccount(wallet2)
		assert.Nil(t, errGet)
		expectedBalance := big.NewInt(0).Add(initialMinting, transferValue)
		expectedBalance.Add(expectedBalance, transferValue)
		assert.Equal(t, expectedBalance.String(), account.Balance)
	})
	t.Run("1 transaction should be sent correctly", func(t *testing.T) {
		_, errSend := chainSimulator.SendTxAndGenerateBlockTilTxIsExecuted(tx3, maxNumOfBlockToGenerateWhenExecutingTx)
		assert.Nil(t, errSend)

		account, errGet := chainSimulator.GetAccount(wallet4)
		assert.Nil(t, errGet)
		expectedBalance := big.NewInt(0).Add(initialMinting, transferValue)
		assert.Equal(t, expectedBalance.String(), account.Balance)
	})
}
