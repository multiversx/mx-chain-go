package sharding

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func createGenesisOneShardOneNode() *Genesis {
	genesis := &Genesis{
		pubkeyConverter: createMockPubkeyConverter(),
	}
	genesis.InitialBalances = make([]*InitialBalance, 1)
	genesis.InitialBalances[0] = &InitialBalance{}
	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	genesis.InitialBalances[0].Balance = "11"

	err := genesis.processConfig()
	if err != nil {
		return nil
	}

	return genesis
}

func createGenesisTwoShardTwoNodes() *Genesis {
	genesis := &Genesis{
		pubkeyConverter: createMockPubkeyConverter(),
	}
	genesis.InitialBalances = make([]*InitialBalance, 4)
	genesis.InitialBalances[0] = &InitialBalance{}
	genesis.InitialBalances[1] = &InitialBalance{}
	genesis.InitialBalances[2] = &InitialBalance{}
	genesis.InitialBalances[3] = &InitialBalance{}

	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	genesis.InitialBalances[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	genesis.InitialBalances[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	genesis.InitialBalances[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"

	genesis.InitialBalances[0].Balance = "999"
	genesis.InitialBalances[1].Balance = "999"
	genesis.InitialBalances[2].Balance = "999"
	genesis.InitialBalances[3].Balance = "999"

	err := genesis.processConfig()
	if err != nil {
		return nil
	}

	return genesis
}

func createGenesisTwoShard6NodesMeta() *Genesis {
	genesis := &Genesis{
		pubkeyConverter: createMockPubkeyConverter(),
	}
	genesis.InitialBalances = make([]*InitialBalance, 6)
	genesis.InitialBalances[0] = &InitialBalance{}
	genesis.InitialBalances[1] = &InitialBalance{}
	genesis.InitialBalances[2] = &InitialBalance{}
	genesis.InitialBalances[3] = &InitialBalance{}
	genesis.InitialBalances[4] = &InitialBalance{}
	genesis.InitialBalances[5] = &InitialBalance{}

	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	genesis.InitialBalances[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	genesis.InitialBalances[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	genesis.InitialBalances[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"
	genesis.InitialBalances[4].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7411"
	genesis.InitialBalances[5].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7410"

	genesis.InitialBalances[0].Balance = "999"
	genesis.InitialBalances[1].Balance = "999"
	genesis.InitialBalances[2].Balance = "999"
	genesis.InitialBalances[3].Balance = "999"
	genesis.InitialBalances[4].Balance = "999"
	genesis.InitialBalances[5].Balance = "999"

	err := genesis.processConfig()
	if err != nil {
		return nil
	}

	return genesis
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func TestGenesis_NewGenesisConfigWrongFile(t *testing.T) {
	genesis, err := NewGenesisConfig("", createMockPubkeyConverter())

	assert.Nil(t, genesis)
	assert.NotNil(t, err)
}

func TestNodes_NewGenesisConfigWrongDataInFile(t *testing.T) {
	genesis, err := NewGenesisConfig("mock/invalidGenesisMock.json", createMockPubkeyConverter())

	assert.Nil(t, genesis)
	assert.True(t, errors.Is(err, ErrCouldNotParsePubKey))
}

func TestNodes_NewGenesisShouldWork(t *testing.T) {
	genesis, err := NewGenesisConfig("mock/genesisMock.json", createMockPubkeyConverter())

	assert.NotNil(t, genesis)
	assert.Nil(t, err)
}

func TestGenesis_ProcessConfigGenesisWithIncompleteDataShouldErr(t *testing.T) {
	genesis := Genesis{
		pubkeyConverter: createMockPubkeyConverter(),
	}

	genesis.InitialBalances = make([]*InitialBalance, 2)
	genesis.InitialBalances[0] = &InitialBalance{}
	genesis.InitialBalances[1] = &InitialBalance{}

	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := genesis.processConfig()

	assert.NotNil(t, genesis)
	assert.True(t, errors.Is(err, ErrCouldNotParsePubKey))
}

func TestGenesis_GenesisWithIncompleteBalance(t *testing.T) {
	genesis := Genesis{
		pubkeyConverter: createMockPubkeyConverter(),
	}

	genesis.InitialBalances = make([]*InitialBalance, 1)
	genesis.InitialBalances[0] = &InitialBalance{}

	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	_ = genesis.processConfig()

	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)

	inBal, err := genesis.InitialNodesBalances(shardCoordinator)

	assert.NotNil(t, genesis)
	assert.Nil(t, err)
	for _, val := range inBal {
		assert.Equal(t, big.NewInt(0), val)
	}
}

func TestGenesis_InitialNodesBalancesNil(t *testing.T) {
	genesis := Genesis{
		pubkeyConverter: createMockPubkeyConverter(),
	}
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator)

	assert.NotNil(t, genesis)
	assert.Equal(t, 0, len(inBalance))
	assert.Nil(t, err)
}

func TestGenesis_InitialNodesBalancesNilShardCoordinatorShouldErr(t *testing.T) {
	genesis := createGenesisOneShardOneNode()
	inBalance, err := genesis.InitialNodesBalances(nil)

	assert.NotNil(t, genesis)
	assert.Nil(t, inBalance)
	assert.Equal(t, ErrNilShardCoordinator, err)
}

func TestGenesis_InitialNodesBalancesGood(t *testing.T) {
	genesis := createGenesisTwoShardTwoNodes()
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(2, 1)
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator)

	assert.NotNil(t, genesis)
	assert.Equal(t, 2, len(inBalance))
	assert.Nil(t, err)
}

func TestGenesis_Initial5NodesBalancesGood(t *testing.T) {
	genesis := createGenesisTwoShard6NodesMeta()
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(2, 1)
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator)

	assert.NotNil(t, genesis)
	assert.Equal(t, 3, len(inBalance))
	assert.Nil(t, err)
}
