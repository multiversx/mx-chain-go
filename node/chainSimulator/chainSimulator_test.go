package chainSimulator

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../cmd/node/config/"
)

func TestNewChainSimulator(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(t.TempDir(), 3, defaultPathToInitialConfig, startTime, roundDurationInMillis, core.OptionalUint64{}, api.NewNoApiInterface())
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	time.Sleep(time.Second)

	err = chainSimulator.Close()
	assert.Nil(t, err)
}

func TestChainSimulator_GenerateBlocksShouldWork(t *testing.T) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	chainSimulator, err := NewChainSimulator(t.TempDir(), 3, defaultPathToInitialConfig, startTime, roundDurationInMillis, core.OptionalUint64{}, api.NewNoApiInterface())
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
	chainSimulator, err := NewChainSimulator(t.TempDir(), 3, defaultPathToInitialConfig, startTime, roundDurationInMillis, roundsPerEpoch, api.NewNoApiInterface())
	require.Nil(t, err)
	require.NotNil(t, chainSimulator)

	facade, err := NewChainSimulatorFacade(chainSimulator)
	require.Nil(t, err)

	initialAccount, err := facade.GetExistingAccountFromBech32AddressString(testdata.GenesisAddressWithStake)
	require.Nil(t, err)

	time.Sleep(time.Second)

	err = chainSimulator.GenerateBlocks(80)
	require.Nil(t, err)

	accountAfterRewards, err := facade.GetExistingAccountFromBech32AddressString(testdata.GenesisAddressWithStake)
	require.Nil(t, err)

	assert.True(t, accountAfterRewards.GetBalance().Cmp(initialAccount.GetBalance()) > 0,
		fmt.Sprintf("initial balance %s, balance after rewards %s", initialAccount.GetBalance().String(), accountAfterRewards.GetBalance().String()))

	fmt.Println(chainSimulator.GetRestAPIInterfaces())

	err = chainSimulator.Close()
	assert.Nil(t, err)
}
