package chainSimulator

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	apiBlock "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/node/external/transactionAPI"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	chainSimulatorCommon "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
)

const (
	defaultPathToInitialConfig            = "../../cmd/node/config/"
	defaultRoundDurationInMillis          = uint64(6000)
	defaultSupernovaRoundDurationInMillis = uint64(600)
	defaultRoundsPerEpochValue            = uint64(20)
	defaultSupernovaRoundsPerEpochValue   = uint64(40)
	defaultNumOfShards                    = uint32(3)
	defaultMinNodesPerShard               = uint32(1)
	defaultMetaChainMinNodes              = uint32(1)
)

var (
	defaultRoundsPerEpoch = core.OptionalUint64{
		HasValue: true,
		Value:    defaultRoundsPerEpochValue,
	}
	defaultSupernovaRoundsPerEpoch = core.OptionalUint64{
		HasValue: true,
		Value:    defaultSupernovaRoundsPerEpochValue,
	}
)

func TestChainSimulatorCheckSupernova(t *testing.T) {
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		AlterConfigsFunction: func(cfg *config.Configs) {

		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	err = chainSimulator.GenerateBlocksUntilEpochIsReached(2)
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(2)
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(1) // supernova round activation
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(50)
	require.Nil(t, err)

	time.Sleep(time.Second)

	chainSimulator.Close()
}

func TestNewChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	alterConfigsFunc := func(cfg *config.Configs) {
		cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 999999
		cfg.RoundConfig.RoundActivations = map[string]config.ActivationRoundByName{
			"DisableAsyncCallV1": {
				Round: "9999999",
			},
			"SupernovaEnableRound": {
				Round: "9999999",
			},
		}
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		AlterConfigsFunction:           alterConfigsFunc,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	for i := 0; i < 8; i++ {
		err = chainSimulator.ForceChangeOfEpoch()
		require.Nil(t, err)
	}

	err = chainSimulator.GenerateBlocks(50)
	require.Nil(t, err)

	time.Sleep(time.Second)

	chainSimulator.Close()
}

func TestChainSimulator_StartWithSupernova(t *testing.T) {
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		BypassBlockSignatureCheck:      true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
		InitialRound:                   20000,
		InitialEpoch:                   1000,
		InitialNonce:                   1000,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 1000
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)
	defer chainSimulator.Close()

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(200)
	require.Nil(t, err)
}

func TestChainSimulator_GenerateBlocksShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		BypassBlockSignatureCheck:      true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
		InitialRound:                   20000,
		InitialEpoch:                   100,
		InitialNonce:                   100,
		AlterConfigsFunction: func(cfg *config.Configs) {
			// we need to enable this as this test skips a lot of epoch activations events, and it will fail otherwise
			// because the owner of a BLS key coming from genesis is not set
			// (the owner is not set at genesis anymore because we do not enable the staking v2 in that phase)
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 99999
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)
	defer chainSimulator.Close()

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(50)
	require.Nil(t, err)

	heartBeats, err := chainSimulator.GetNodeHandler(0).GetFacadeHandler().GetHeartbeats()
	require.Nil(t, err)
	require.Equal(t, 4, len(heartBeats))

}

func TestChainSimulator_VerifyBlockTimestampSupernova(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	supernovaActivationRound := uint64(220)

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch: core.OptionalUint64{
			Value:    20,
			HasValue: true,
		},
		SupernovaRoundsPerEpoch: defaultSupernovaRoundsPerEpoch,
		ApiInterface:            api.NewNoApiInterface(),
		MinNodesPerShard:        defaultMinNodesPerShard,
		MetaChainMinNodes:       defaultMetaChainMinNodes,
		InitialRound:            200,
		InitialEpoch:            100,
		InitialNonce:            100,
		AlterConfigsFunction: func(cfg *config.Configs) {
			// we need to enable this as this test skips a lot of epoch activations events, and it will fail otherwise
			// because the owner of a BLS key coming from genesis is not set
			// (the owner is not set at genesis anymore because we do not enable the staking v2 in that phase)
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 100
			cfg.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)] = config.ActivationRoundByName{
				Round: fmt.Sprintf("%d", supernovaActivationRound),
			}
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)
	defer chainSimulator.Close()

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(30)
	require.Nil(t, err)

	blockBeforeSupernovaRound, err := chainSimulator.GetNodeHandler(0).GetFacadeHandler().GetBlockByRound(supernovaActivationRound-1, apiBlock.BlockQueryOptions{})
	require.Nil(t, err)

	blockS, err := chainSimulator.GetNodeHandler(0).GetFacadeHandler().GetBlockByRound(supernovaActivationRound, apiBlock.BlockQueryOptions{})
	require.Nil(t, err)

	blockAfterSupernovaRound, err := chainSimulator.GetNodeHandler(0).GetFacadeHandler().GetBlockByRound(supernovaActivationRound+1, apiBlock.BlockQueryOptions{})
	require.Nil(t, err)

	diff := blockS.TimestampMs - blockBeforeSupernovaRound.TimestampMs
	require.Equal(t, int64(6000), diff)
	diff = blockAfterSupernovaRound.TimestampMs - blockS.TimestampMs
	require.Equal(t, defaultSupernovaRoundDurationInMillis, uint64(diff))
}

func TestChainSimulator_EpochStartMetaBlockV3ShouldUseTimestampMs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	initialEpoch := uint32(100)
	initialRound := int64(200)

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
		InitialRound:                   initialRound,
		InitialEpoch:                   initialEpoch,
		InitialNonce:                   100,
		AlterConfigsFunction: func(cfg *config.Configs) {
			// we need to enable this as this test skips a lot of epoch activations events, and it will fail otherwise
			// because the owner of a BLS key coming from genesis is not set
			// (the owner is not set at genesis anymore because we do not enable the staking v2 in that phase)
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = initialEpoch
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)
	defer chainSimulator.Close()

	shardNode := chainSimulator.GetNodeHandler(0)
	expectedTimestampMs := uint64(shardNode.GetCoreComponents().RoundHandler().TimeStamp().UnixMilli())
	expectedTimestampSec := uint64(shardNode.GetCoreComponents().RoundHandler().TimeStamp().Unix())

	epochStartTimestampMs := shardNode.GetProcessComponents().BlockchainHook().EpochStartBlockTimeStampMs()

	require.Equal(t, expectedTimestampMs, epochStartTimestampMs)
	require.NotEqual(t, expectedTimestampSec, epochStartTimestampMs)
}

func TestChainSimulator_GenerateBlocksAndEpochChangeShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               100,
		MetaChainMinNodes:              100,
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

func TestSimulator_TriggerChangeOfEpoch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    15000,
	}
	supernovaRoundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    150000,
	}
	alterConfigsFunc := func(cfg *config.Configs) {
		cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 999999
		cfg.RoundConfig.RoundActivations = map[string]config.ActivationRoundByName{
			"DisableAsyncCallV1": {
				Round: "9999999",
			},
			"SupernovaEnableRound": {
				Round: "9999999",
			},
		}
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 roundsPerEpoch,
		SupernovaRoundsPerEpoch:        supernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               100,
		MetaChainMinNodes:              100,
		AlterConfigsFunction:           alterConfigsFunc,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	err = chainSimulator.ForceChangeOfEpoch()
	require.Nil(t, err)

	err = chainSimulator.ForceChangeOfEpoch()
	require.Nil(t, err)

	err = chainSimulator.ForceChangeOfEpoch()
	require.Nil(t, err)

	err = chainSimulator.ForceChangeOfEpoch()
	require.Nil(t, err)

	metaNode := chainSimulator.GetNodeHandler(core.MetachainShardId)
	currentEpoch := metaNode.GetProcessComponents().EpochStartTrigger().Epoch()
	require.Equal(t, uint32(4), currentEpoch)
}

func TestChainSimulator_ChangeRoundsPerEpoch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	supernovaRoundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 roundsPerEpoch,
		SupernovaRoundsPerEpoch:        supernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[0].EnableEpoch = 0
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[0].RoundsPerEpoch = 10
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[0].MinRoundsBetweenEpochs = 10

			cfg.EpochConfig.EnableEpochs.AndromedaEnableEpoch = 3
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[1].EnableEpoch = 3
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[1].RoundsPerEpoch = 20
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[1].MinRoundsBetweenEpochs = 10

			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 5
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[2].EnableEpoch = 5
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[2].RoundsPerEpoch = 30
			cfg.GeneralConfig.GeneralSettings.ChainParametersByEpoch[2].MinRoundsBetweenEpochs = 10
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	err = chainSimulator.GenerateBlocks(140)
	require.Nil(t, err)

	expectedEpoch := uint32(7)

	metaNode := chainSimulator.GetNodeHandler(core.MetachainShardId)
	currentEpoch := metaNode.GetProcessComponents().EpochStartTrigger().Epoch()
	require.Equal(t, expectedEpoch, currentEpoch)

	defer chainSimulator.Close()

}

func TestChainSimulator_SetState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
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

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	balance := "431271308732096033771131"
	contractAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	accountState := &dtos.AddressState{
		Address:          contractAddress,
		Nonce:            new(uint64),
		Balance:          balance,
		Code:             "0061736d010000000129086000006000017f60027f7f017f60027f7f0060017f0060037f7f7f017f60037f7f7f0060017f017f0290020b03656e7619626967496e74476574556e7369676e6564417267756d656e74000303656e760f6765744e756d417267756d656e7473000103656e760b7369676e616c4572726f72000303656e76126d42756666657253746f726167654c6f6164000203656e76176d427566666572546f426967496e74556e7369676e6564000203656e76196d42756666657246726f6d426967496e74556e7369676e6564000203656e76136d42756666657253746f7261676553746f7265000203656e760f6d4275666665725365744279746573000503656e760e636865636b4e6f5061796d656e74000003656e7614626967496e7446696e697368556e7369676e6564000403656e7609626967496e744164640006030b0a010104070301000000000503010003060f027f0041a080080b7f0041a080080b074607066d656d6f7279020004696e697400110667657453756d00120361646400130863616c6c4261636b00140a5f5f646174615f656e6403000b5f5f686561705f6261736503010aca010a0e01017f4100100c2200100020000b1901017f419c8008419c800828020041016b220036020020000b1400100120004604400f0b4180800841191002000b16002000100c220010031a2000100c220010041a20000b1401017f100c2202200110051a2000200210061a0b1301017f100c220041998008410310071a20000b1401017f10084101100d100b210010102000100f0b0e0010084100100d1010100e10090b2201037f10084101100d100b210110102202100e220020002001100a20022000100f0b0300010b0b2f0200418080080b1c77726f6e67206e756d626572206f6620617267756d656e747373756d00419c80080b049cffffff",
		CodeHash:         "n9EviPlHS6EV+3Xp0YqP28T0IUfeAFRFBIRC1Jw6pyU=",
		RootHash:         "76cr5Jhn6HmBcDUMIzikEpqFgZxIrOzgNkTHNatXzC4=",
		CodeMetadata:     "BQY=",
		Owner:            "erd1ss6u80ruas2phpmr82r42xnkd6rxy40g9jl69frppl4qez9w2jpsqj8x97",
		DeveloperRewards: "5401004999998",
		Pairs: map[string]string{
			"73756d": "0a",
		},
	}

	chainSimulatorCommon.CheckSetEntireState(t, chainSimulator, chainSimulator.GetNodeHandler(1), accountState)
}

func TestChainSimulator_SetEntireStateWithRemoval(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	balance := "431271308732096033771131"
	contractAddress := "erd1qqqqqqqqqqqqqpgqmzzm05jeav6d5qvna0q2pmcllelkz8xddz3syjszx5"
	accountState := &dtos.AddressState{
		Address:          contractAddress,
		Nonce:            new(uint64),
		Balance:          balance,
		Code:             "0061736d010000000129086000006000017f60027f7f017f60027f7f0060017f0060037f7f7f017f60037f7f7f0060017f017f0290020b03656e7619626967496e74476574556e7369676e6564417267756d656e74000303656e760f6765744e756d417267756d656e7473000103656e760b7369676e616c4572726f72000303656e76126d42756666657253746f726167654c6f6164000203656e76176d427566666572546f426967496e74556e7369676e6564000203656e76196d42756666657246726f6d426967496e74556e7369676e6564000203656e76136d42756666657253746f7261676553746f7265000203656e760f6d4275666665725365744279746573000503656e760e636865636b4e6f5061796d656e74000003656e7614626967496e7446696e697368556e7369676e6564000403656e7609626967496e744164640006030b0a010104070301000000000503010003060f027f0041a080080b7f0041a080080b074607066d656d6f7279020004696e697400110667657453756d00120361646400130863616c6c4261636b00140a5f5f646174615f656e6403000b5f5f686561705f6261736503010aca010a0e01017f4100100c2200100020000b1901017f419c8008419c800828020041016b220036020020000b1400100120004604400f0b4180800841191002000b16002000100c220010031a2000100c220010041a20000b1401017f100c2202200110051a2000200210061a0b1301017f100c220041998008410310071a20000b1401017f10084101100d100b210010102000100f0b0e0010084100100d1010100e10090b2201037f10084101100d100b210110102202100e220020002001100a20022000100f0b0300010b0b2f0200418080080b1c77726f6e67206e756d626572206f6620617267756d656e747373756d00419c80080b049cffffff",
		CodeHash:         "n9EviPlHS6EV+3Xp0YqP28T0IUfeAFRFBIRC1Jw6pyU=",
		RootHash:         "eqIumOaMn7G5cNSViK3XHZIW/C392ehfHxOZkHGp+Gc=", // root hash with auto balancing enabled
		CodeMetadata:     "BQY=",
		Owner:            "erd1ss6u80ruas2phpmr82r42xnkd6rxy40g9jl69frppl4qez9w2jpsqj8x97",
		DeveloperRewards: "5401004999998",
		Pairs: map[string]string{
			"73756d": "0a",
		},
	}
	chainSimulatorCommon.CheckSetEntireStateWithRemoval(t, chainSimulator, chainSimulator.GetNodeHandler(1), accountState)
}

func TestChainSimulator_GetAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	// the facade's GetAccount method requires that at least one block was produced over the genesis block
	_ = chainSimulator.GenerateBlocks(1)

	defer chainSimulator.Close()

	chainSimulatorCommon.CheckGetAccount(t, chainSimulator)
}

func TestSimulator_SendTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	chainSimulatorCommon.CheckGenerateTransactions(t, chainSimulator)
}

func TestSimulator_MoveBalanceCheckReceipt(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = uint32(2)
			cfg.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)] = config.ActivationRoundByName{
				Round: "46",
			}
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	wallet0, err := chainSimulator.GenerateAndMintWalletAddress(0, chainSimulatorCommon.OneEGLD)
	require.Nil(t, err)
	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	ftx := &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(1),
		SndAddr:   wallet0.Bytes,
		RcvAddr:   wallet0.Bytes,
		Data:      []byte(""),
		GasLimit:  100_000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
		Signature: []byte("010101"),
	}

	checkReceipts := func(te *testing.T, aB *apiBlock.Block, value string) {
		called := false
		for _, mb := range aB.MiniBlocks {
			if mb.Type == block.ReceiptBlock.String() {
				called = true
				require.Equal(te, 1, len(mb.Receipts))
				require.Equal(te, value, mb.Receipts[0].Value.String())
			}
		}
		require.True(te, called)
	}

	apiTx, err := chainSimulator.SendTxAndGenerateBlockTilTxIsExecuted(ftx, 10)
	require.Nil(t, err)
	require.NotNil(t, apiTx)

	blockWithTxs, err := chainSimulator.GetNodeHandler(0).GetFacadeHandler().GetBlockByNonce(apiTx.BlockNonce, apiBlock.BlockQueryOptions{
		WithTransactions: true,
		WithLogs:         true,
	})
	require.Nil(t, err)
	require.Equal(t, 2, len(blockWithTxs.MiniBlocks))
	checkReceipts(t, blockWithTxs, "50000000000000")

	err = chainSimulator.GenerateBlocks(50)
	require.Nil(t, err)

	ftx.Nonce++
	apiTx, err = chainSimulator.SendTxAndGenerateBlockTilTxIsExecuted(ftx, 10)
	require.Nil(t, err)
	require.NotNil(t, apiTx)

	blockWithTxs, err = chainSimulator.GetNodeHandler(0).GetFacadeHandler().GetBlockByNonce(apiTx.BlockNonce, apiBlock.BlockQueryOptions{
		WithTransactions: true,
		WithLogs:         true,
	})
	require.Nil(t, err)
	require.Equal(t, 2, len(blockWithTxs.MiniBlocks))
	checkReceipts(t, blockWithTxs, "500000000000")
}

func TestSimulator_SentMoveBalanceNoGasForFee(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	wallet0, err := chainSimulator.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.Nil(t, err)

	ftx := &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(0),
		SndAddr:   wallet0.Bytes,
		RcvAddr:   wallet0.Bytes,
		Data:      []byte(""),
		GasLimit:  50_000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
		Signature: []byte("010101"),
	}
	_, err = chainSimulator.sendTx(ftx)
	require.True(t, strings.Contains(err.Error(), errors.ErrInsufficientFunds.Error()))
}

func TestSimulator_SendMoveBalanceTxBeforeAndAfterSupernovaWithMoreGasLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
		CreateBlockMaxTimePercent:      0.25,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 2
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	defer chainSimulator.Close()

	chainSimulatorCommon.GenerateMoveBalanceTxsInShardsWithMoreGasLimit(t, chainSimulator)

	err = chainSimulator.GenerateBlocksUntilEpochIsReached(3)
	require.Nil(t, err)

	chainSimulatorCommon.GenerateMoveBalanceTxsInShardsWithMoreGasLimit(t, chainSimulator)
}

// TestRemoveSCRFromPoolAndDestinationShouldBeRequested checks that, after an
// ESDT issue SCR is manually removed from the pools, the destination shard
// requests it again and the SCR becomes available through the API.
func TestRemoveSCRFromPoolAndDestinationShouldBeRequested(t *testing.T) {
	activationEpoch := uint32(4)

	baseIssuingCost := "1000"

	cs, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
			cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = uint32(2)
			cfg.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)] = config.ActivationRoundByName{
				Round: "46",
			}

		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	wallet0, err := cs.GenerateAndMintWalletAddress(0, chainSimulatorCommon.OneEGLD)
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(int32(activationEpoch))
	require.Nil(t, err)

	nftTicker := []byte("NFTTICKER")
	nonce := uint64(0)

	callValue, _ := big.NewInt(0).SetString(baseIssuingCost, 10)

	txDataField := bytes.Join(
		[][]byte{
			[]byte("issueNonFungible"),
			[]byte(hex.EncodeToString([]byte("asdname"))),
			[]byte(hex.EncodeToString(nftTicker)),
		},
		[]byte("@"),
	)

	tx := &transaction.Transaction{
		Nonce:     nonce,
		SndAddr:   wallet0.Bytes,
		RcvAddr:   core.ESDTSCAddress,
		GasLimit:  100_000_000,
		GasPrice:  1_000_000_000,
		Signature: []byte("dummySig"),
		Data:      txDataField,
		Value:     callValue,
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, 10)
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, "success", txResult.Status.String())

	//  SCRS remove from pool
	keys := cs.GetNodeHandler(core.MetachainShardId).GetDataComponents().Datapool().UnsignedTransactions().Keys()
	scrsForShardZero := cs.GetNodeHandler(0).GetDataComponents().Datapool().UnsignedTransactions().Keys()
	for _, key := range keys {
		cs.GetNodeHandler(core.MetachainShardId).GetDataComponents().Datapool().UnsignedTransactions().RemoveDataFromAllShards(key)
		cs.GetNodeHandler(0).GetDataComponents().Datapool().UnsignedTransactions().RemoveDataFromAllShards(key)
	}

	scrHash := scrsForShardZero[0]
	res, err := cs.GetNodeHandler(0).GetFacadeHandler().GetTransaction(hex.EncodeToString(scrHash), true)
	require.Nil(t, res)
	require.True(t, strings.Contains(err.Error(), transactionAPI.ErrTransactionNotFound.Error()))

	called := false
	count := 0
	for {
		count++
		err = cs.GenerateBlocks(1)
		require.Nil(t, err)

		res, _ = cs.GetNodeHandler(0).GetFacadeHandler().GetTransaction(hex.EncodeToString(scrHash), true)
		if res != nil {
			called = true
			break
		}
		if count == 100 {
			require.FailNow(t, "cannot find SCR on the destination shard")
		}
	}
	require.True(t, called)
}

func TestChainSimulator_VerifyEconomicsMetricsSupernova(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	supernovaActivationRound := uint64(46)
	supernovaActivationEpoch := uint64(2)

	cs, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch: core.OptionalUint64{
			Value:    20,
			HasValue: true,
		},
		SupernovaRoundsPerEpoch: defaultSupernovaRoundsPerEpoch,
		ApiInterface:            api.NewNoApiInterface(),
		MinNodesPerShard:        defaultMinNodesPerShard,
		MetaChainMinNodes:       defaultMetaChainMinNodes,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = uint32(supernovaActivationEpoch)
			cfg.RoundConfig.RoundActivations[string(common.SupernovaRoundFlag)] = config.ActivationRoundByName{
				Round: fmt.Sprintf("%d", supernovaActivationRound),
			}
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	require.Nil(t, cs.GenerateBlocksUntilEpochIsReached(int32(supernovaActivationEpoch)))

	mintValue := big.NewInt(0).Mul(chainSimulatorCommon.OneEGLD, big.NewInt(3000*5))
	wallet1, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)

	_, blsKeys, err := GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nonce := uint64(0)
	for currentEpoch := supernovaActivationEpoch + 1; currentEpoch < supernovaActivationEpoch+4; currentEpoch++ {
		dataFieldTx1 := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
		tx1Value := big.NewInt(0).Mul(big.NewInt(2501), chainSimulatorCommon.OneEGLD)
		tx1 := chainSimulatorCommon.GenerateTransaction(wallet1.Bytes, nonce, vm.ValidatorSCAddress, tx1Value, dataFieldTx1, staking.GasLimitForStakeOperation)

		results, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx1}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.Equal(t, 1, len(results))
		require.NotNil(t, results)

		require.Nil(t, cs.GenerateBlocksUntilEpochIsReached(int32(currentEpoch)))
		checkMetrics(t, cs, core.MetachainShardId, currentEpoch)
		checkMetrics(t, cs, 0, currentEpoch)

		nonce++
	}
}

func checkMetrics(t *testing.T, cs ChainSimulator, shardID uint32, expectedEpoch uint64) {
	res, err := cs.GetNodeHandler(shardID).GetFacadeHandler().StatusMetrics().EconomicsMetrics()
	require.Nil(t, err)

	expectedMetrics := map[string]struct{}{
		common.MetricTotalSupply:           {},
		common.MetricInflation:             {},
		common.MetricEpochForEconomicsData: {},
		common.MetricTotalFees:             {},
		common.MetricDevRewardsInEpoch:     {},
	}

	for foundMetric, metricValue := range res {
		require.Contains(t, expectedMetrics, foundMetric)

		switch metricVal := metricValue.(type) {
		case string:
			require.Greater(t, len(metricVal), 1)
		case uint64:
			require.Equal(t, expectedEpoch, metricValue)
		default:
			require.Fail(t, "metric value is not a string or uint64")
		}

		delete(expectedMetrics, foundMetric)

	}

	require.Empty(t, expectedMetrics, "should've found all expected metrics in the result from facade")
}

func TestChainSimulator_VMQueryShardContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	initialEpoch := uint32(100)
	initialRound := int64(100)
	initialNonce := uint64(100)

	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    defaultNumOfShards,
		RoundDurationInMillis:          defaultRoundDurationInMillis,
		SupernovaRoundDurationInMillis: defaultSupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 defaultRoundsPerEpoch,
		SupernovaRoundsPerEpoch:        defaultSupernovaRoundsPerEpoch,
		ApiInterface:                   api.NewNoApiInterface(),
		MinNodesPerShard:               defaultMinNodesPerShard,
		MetaChainMinNodes:              defaultMetaChainMinNodes,
		InitialEpoch:                   initialEpoch,
		InitialRound:                   initialRound,
		InitialNonce:                   initialNonce,
		VmQueryDelayAfterStartInMs:     0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV2EnableEpoch = 0
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = 150
		},
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)
	defer chainSimulator.Close()

	err = chainSimulator.SetStateMultiple([]*dtos.AddressState{
		{
			Address:          "erd1qqqqqqqqqqqqqpgqhe8t5jewej70zupmh44jurgn29psua5l2jps3ntjj3",
			Nonce:            nil,
			Balance:          "251280789604421862657421",
			Code:             "0061736d0100000001761360000060017f017f60017f0060027f7f017f60027f7f006000017f60037f7f7f0060037f7f7f017f60057f7f7e7f7f017f60047f7f7f7f0060067e7f7f7f7f7f017f60057f7f7f7e7f0060027f7e0060017e006000017e60047f7f7f7f017f60057f7f7f7f7f0060037e7f7f0060057f7f7e7f7f0002a3072703656e760b7369676e616c4572726f72000403656e760a6d4275666665724e6577000503656e760d6d427566666572417070656e64000303656e76096d4275666665724571000303656e76106d616e61676564534341646472657373000203656e761b6d616e61676564457865637574654f6e44657374436f6e74657874000a03656e760f636c65616e52657475726e44617461000003656e760d6d616e6167656443616c6c6572000203656e76136d616e616765644f776e657241646472657373000203656e7612626967496e7447657443616c6c56616c7565000203656e76136765744e756d455344545472616e7366657273000503656e760f6765744e756d417267756d656e7473000503656e76126d427566666572417070656e644279746573000703656e760f6d4275666665725365744279746573000703656e76196d42756666657246726f6d426967496e74556e7369676e6564000303656e760a626967496e745369676e000103656e76106d4275666665724765744c656e677468000103656e76126d42756666657253746f726167654c6f6164000303656e76136d42756666657253746f7261676553746f7265000303656e76126d427566666572476574417267756d656e74000303656e76126d616e616765645369676e616c4572726f72000203656e760f6973536d617274436f6e7472616374000103656e760f6d4275666665724765744279746573000303656e761c626967496e744765744553445445787465726e616c42616c616e6365000b03656e7618626967496e7447657445787465726e616c42616c616e6365000403656e760e626967496e74536574496e743634000c03656e7609626967496e74416464000603656e76226d616e616765644d756c74695472616e73666572455344544e465445786563757465000803656e760e636865636b4e6f5061796d656e74000003656e7614626967496e7446696e697368556e7369676e6564000203656e760d6d42756666657246696e697368000103656e760666696e697368000403656e7614736d616c6c496e7446696e6973685369676e6564000d03656e7616626967496e744765744553445443616c6c56616c7565000203656e761067657445534454546f6b656e4e616d65000103656e7609626967496e74436d70000303656e760a6765744761734c656674000e03656e761b6d616e616765645472616e7366657256616c756545786563757465000803656e76136d42756666657247657442797465536c696365000f0338371000010300041105020500000401010202000404050104010301010604040601030205120909050105050200000000000000000000000005030100110619037f01418080c0000b7f0041e483c0000b7f0041f083c0000b07a7010d066d656d6f7279020004696e69740052146765744c6f636b656445676c6442616c616e63650053156765745772617070656445676c64546f6b656e496400540869735061757365640055057061757365005609726562616c616e6365005707756e706175736500580a756e7772617045676c640059087772617045676c64005a0863616c6c4261636b005b0a5f5f646174615f656e6403010b5f5f686561705f6261736503020a9f16372e000240200120024d0440200220044d0d011028000b1028000b2000200220016b3602042000200120036a3602000b0500105d000b0f01017f10012201200010021a20010b0b0020002001100341004a0b0c0041fa82c00041121000000b0900200020011000000b2101017f41671004102e2203102f20004167200320012002102e10051a100610060b1b01017f419c83c000419c83c00028020041016b220036020020000b08002000420010190b0c01017f102e2200100720000b1e01017f102e2200100820001030102a04400f0b419a80c00041241000000b1700100a450440417510090f0b41be80c00041251000000b3e01027f02402001280200220341a083c0002802004e0440410121020c010b2001200341016a3602002003103421010b20002001360204200020023602000b0d002000102e220010131a20000b4e01017f230041106b220124000240200010104104470d002001410036020c20002001410c6a41041042200128020c41c58eb1a204470d002000418c83c0004100100d1a0b200141106a240020000b1a00200041a083c00028020048044041f480c00041121000000b0b1500100b20004604400f0b418681c00041191000000b1b0041a083c00028020041004e04400f0b41e380c00041111000000b4701017f230041106b2202240020022001410874418080fc077120014118747220014108764180fe03712001411876727236020c20002002410c6a4104100c1a200241106a24000b2d01017f103b210202402001103c4504402001102921020c010b200241ad81c0004104100d1a0b2000200210390b1401017f102e2200418c83c0004100100d1a20000b070020001010450b1601017f103b1a102e22022001100e1a2000200210390b1300417f2000100f220041004720004100481b0b1101017f102e220220002001100d1a20020bad0102027f017e230041106b2201240020014200370308200010412200101022024109490440200141002002200141086a41081027200020012802002202200128020422001042027f41002000450d001a034020000440200041016b210020023100002003420886842103200241016a21020c010b0b02402003420158044041002003a741016b0d021a0c010b41b181c00041121043000b41010b200141106a24000f0b419f81c000410e1043000b0d002000102e220010111a20000b0d00200041002002200110261a0b1b01017f41c381c0004116103f220220002001100c1a20021014000b9c0101037f230041106b2203240020032001ad4238863703080240200141ff0171450440418c83c00021010c010b4100210103400240024020014108470440200341086a20016a2d00002202450d022002411874411f75220220016a220441094f0d01200341086a20026a20016a2101410820046b21020c040b105c000b105c000b200141016a21010c000b000b2000200120021045200341106a24000b0d00200020012002103f10121a0b08002000104110350b1d002000104841a483c000101541004c044041000f0b2001103c4101730b0c00200041a483c00010161a0b5401047f103b2202103c102e2200100420002103102e210045044020021010210120031048200241c483c00010161a41a483c00041c483c000200142002000101720000f0b2003104841a483c0002000101820000b9f0201047f230041306b22052400103b2107200110292106102e2201102f200120014175101a2001103e41ff017104402006103c1a0b200541206a42003703002005420037031820052006410874418080fc077120064118747220064108764180fe037120064118767272360228200541106a200541186a220641004104104b20052802102005280214200541286a22084104104c20054200370328200541086a20064104410c104b2005280208200528020c20084108104c20052001410874418080fc077120014118747220014108764180fe03712001411876727236022820052006410c4110104b2005280200200528020420084104104c200720064110100c1a20002802002007200220032802002004280200101b1a200541306a24000b3b01017f230041106b22042400200441086a20022003200141101027200428020c21012000200428020836020020002001360204200441106a24000bb50201067f2001200346044020012203410f4b04402000410020006b41037122046a210520040440200221010340200020012d00003a0000200141016a2101200041016a22002005490d000b0b2005200320046b2203417c7122066a21000240200220046a22044103710440200641004c0d012004410374220141187121072004417c71220841046a2102410020016b4118712109200828020021010340200520012007762002280200220120097472360200200241046a2102200541046a22052000490d000b0c010b200641004c0d0020042102034020052002280200360200200241046a2102200541046a22052000490d000b0b20034103712103200420066a21020b20030440200020036a21010340200020022d00003a0000200241016a2102200041016a22002001490d000b0b0f0b105d000b100041e782c0004113103f10404101730b0d002000103e41ff01714101460b0b0041d582c0004112103f0b0b00418c83c0004100103f0b0f0041e782c0004113103f200010440b4201027f101c410110374100103410352100104f210102402000103c4504402001200010121a0c010b200141ad81c000410410450b41e782c0004113103f410110440b0c00101c410010371049101d0b2501017f101c41001037104f10462200103c4504402000101e1a0f0b41ad81c0004104101f0b1600101c4100103741e782c0004113103f1040ad10200b0e00101c103141001037410110510b0a0010321031410010370b0e00101c103141001037410010510bfc0201067f230041306b2200240002400240024002400240024002400240100a45044041751009417521010c010b4173210141731021100a0d010b103b21020c010b200041286a4200370300200041206a4200370300200041186a4200370300200042003703100240200041106a10222201450440102e2202418c83c0004100100d1a0c010b200141214f0d02200041106a2001103f21020b417321010b41a083c000100b360200103820004100360210200041086a200041106a1033200028020c21032000280208210420002802101036104d450d012002104f10462205102a450d022001104e450d0320011049102341004a0d04103b22022005103a20022001103d1024418080c000410d103f2002102d02401030220220040440103b21030b200310474504402002200142001050100110251a0c010b2002200110244290ce007d2003103b10251a0b200041306a24000f0b1028000b102b000b41d981c0004110102c000b41e981c000411c102c000b418582c0004123102c000bae0202057f017e230041206b220024001032103b210141a083c000100b36020010382000410036021c200041086a2000411c6a1033200028020c210220002802082104200028021c103602400240104d04402001103c450d014175104e450d02104f10462101103b22032001103a20034175103d1024418d80c000410d103f2003102d20001030220336021020040440103b21020b2000200236021402402003200210474504402000105022023602182000103b220436021c2001103c04402003417542002002200410251a0c020b200041106a20014200200041186a2000411c6a104a0c010b102421052000103b36021c200041106a200120054290ce007d200041146a2000411c6a104a0b200041206a24000f0b102b000b41a882c0004112102c000b41ba82c000411b102c000b0300010b0c00418c83c000410e1000000b0500105c000b0bb0030200418080c0000b9a03455344544c6f63616c4275726e455344544c6f63616c4d696e74456e64706f696e742063616e206f6e6c792062652063616c6c6564206279206f776e657266756e6374696f6e20646f6573206e6f74206163636570742045534454207061796d656e74746f6f2066657720617267756d656e7473746f6f206d616e7920617267756d656e747377726f6e67206e756d626572206f6620617267756d656e7473696e70757420746f6f206c6f6e6745474c44696e707574206f7574206f662072616e676573746f72616765206465636f6465206572726f723a2057726f6e67206573647420746f6b656e4d75737420706179206d6f7265207468616e203020746f6b656e7321436f6e747261637420646f6573206e6f74206861766520656e6f7567682066756e64734f6e6c792045474c442061636365707465645061796d656e74206d757374206265206d6f7265207468616e20307772617070656445676c64546f6b656e496470617573655f6d6f64756c653a706175736564636f6e74726163742069732070617573656470616e6963206f6363757272656400419c83c0000b049cffffff",
			CodeMetadata:     "BQA=",
			CodeHash:         "tqi9WZfp6eHSWSRS6/yMuc0cFLdyhmFrQuKpz09f4ac=",
			DeveloperRewards: "",
			Owner:            "",
			Pairs: map[string]string{
				"454c524f4e44657364745745474c442d626434643739":         "120900035f99906af3d080",
				"454c524f4e44726f6c65657364745745474c442d626434643739": "0a1145534454526f6c654c6f63616c4d696e740a1145534454526f6c654c6f63616c4275726e",
				"7772617070656445676c64546f6b656e4964":                 "5745474c442d626434643739",
			},
		},
		{
			Address: "erd1lllllllllllllllllllllllllllllllllllllllllllllllllllsckry7t",
			Pairs: map[string]string{
				"454c524f4e44657364745745474c442d626434643739": "0000",
			},
		},
	})
	require.NoError(t, err)

	err = chainSimulator.GenerateBlocks(2)

	addressBytes, err := chainSimulator.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqpgqhe8t5jewej70zupmh44jurgn29psua5l2jps3ntjj3")
	require.NoError(t, err)

	scQuery := &process.SCQuery{
		ScAddress:  addressBytes,
		FuncName:   "isPaused",
		CallerAddr: addressBytes,
		CallValue:  big.NewInt(0),
	}

	res, _, err := chainSimulator.GetNodeHandler(1).GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, "ok", res.ReturnCode)
}
