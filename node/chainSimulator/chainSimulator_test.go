package chainSimulator

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../cmd/node/config/"
)

func TestNewChainSimulator(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         core.OptionalUint64{},
		ApiInterface:           api.NewNoApiInterface(),
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	time.Sleep(time.Second)

	err = chainSimulator.Close()
	assert.Nil(t, err)
}

func TestChainSimulator_GenerateBlocksShouldWork(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         core.OptionalUint64{},
		ApiInterface:           api.NewNoApiInterface(),
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(10)
	require.Nil(t, err)

	err = chainSimulator.Close()
	assert.Nil(t, err)
}

func TestChainSimulator_GenerateBlocksAndEpochChangeShouldWork(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	facade, err := NewChainSimulatorFacade(chainSimulator)
	require.Nil(t, err)

	genesisAddressWithStake := chainSimulator.initialWalletKeys.InitialWalletWithStake.Address
	initialAccount, err := facade.GetExistingAccountFromBech32AddressString(genesisAddressWithStake)
	require.Nil(t, err)

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(80)
	require.Nil(t, err)

	accountAfterRewards, err := facade.GetExistingAccountFromBech32AddressString(genesisAddressWithStake)
	require.Nil(t, err)

	assert.True(t, accountAfterRewards.GetBalance().Cmp(initialAccount.GetBalance()) > 0,
		fmt.Sprintf("initial balance %s, balance after rewards %s", initialAccount.GetBalance().String(), accountAfterRewards.GetBalance().String()))

	fmt.Println(chainSimulator.GetRestAPIInterfaces())

	err = chainSimulator.Close()
	assert.Nil(t, err)
}

func TestChainSimulator_SetState(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	chainSimulator, err := NewChainSimulator(ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
	})
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	keyValueMap := map[string]string{
		"01": "01",
		"02": "02",
	}

	address := "erd1qtc600lryvytxuy4h7vn7xmsy5tw6vuw3tskr75cwnmv4mnyjgsq6e5zgj"
	err = chainSimulator.SetKeyValueForAddress(address, keyValueMap)
	require.Nil(t, err)

	err = chainSimulator.GenerateBlocks(1)
	require.Nil(t, err)

	nodeHandler := chainSimulator.GetNodeHandler(0)
	keyValuePairs, _, err := nodeHandler.GetFacadeHandler().GetKeyValuePairs(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.Equal(t, keyValueMap, keyValuePairs)
}
