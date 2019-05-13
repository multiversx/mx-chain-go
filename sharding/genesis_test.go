package sharding_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func createGenesisOneShardOneNode() *sharding.Genesis {
	g := &sharding.Genesis{}
	g.InitialBalances = make([]*sharding.InitialBalance, 1)
	g.InitialBalances[0] = &sharding.InitialBalance{}
	g.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialBalances[0].Balance = "11"

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}

	return g
}

func createGenesisTwoShardTwoNodes() *sharding.Genesis {
	g := &sharding.Genesis{}
	g.InitialBalances = make([]*sharding.InitialBalance, 4)
	g.InitialBalances[0] = &sharding.InitialBalance{}
	g.InitialBalances[1] = &sharding.InitialBalance{}
	g.InitialBalances[2] = &sharding.InitialBalance{}
	g.InitialBalances[3] = &sharding.InitialBalance{}

	g.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialBalances[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	g.InitialBalances[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	g.InitialBalances[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"

	g.InitialBalances[0].Balance = "999"
	g.InitialBalances[1].Balance = "999"
	g.InitialBalances[2].Balance = "999"
	g.InitialBalances[3].Balance = "999"

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}

	return g
}

func createGenesisTwoShard6NodesMeta() *sharding.Genesis {
	g := &sharding.Genesis{}
	g.InitialBalances = make([]*sharding.InitialBalance, 6)
	g.InitialBalances[0] = &sharding.InitialBalance{}
	g.InitialBalances[1] = &sharding.InitialBalance{}
	g.InitialBalances[2] = &sharding.InitialBalance{}
	g.InitialBalances[3] = &sharding.InitialBalance{}
	g.InitialBalances[4] = &sharding.InitialBalance{}
	g.InitialBalances[5] = &sharding.InitialBalance{}

	g.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	g.InitialBalances[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	g.InitialBalances[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	g.InitialBalances[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"
	g.InitialBalances[4].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7411"
	g.InitialBalances[5].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7410"

	g.InitialBalances[0].Balance = "999"
	g.InitialBalances[1].Balance = "999"
	g.InitialBalances[2].Balance = "999"
	g.InitialBalances[3].Balance = "999"
	g.InitialBalances[4].Balance = "999"
	g.InitialBalances[5].Balance = "999"

	err := g.ProcessConfig()
	if err != nil {
		return nil
	}

	return g
}

func TestGenesis_NewGenesisConfigWrongFile(t *testing.T) {
	genesis, err := sharding.NewGenesisConfig("")

	assert.Nil(t, genesis)
	assert.NotNil(t, err)
}

func TestGenesis_ProcessConfigGenesisWithIncompleteDataShouldErr(t *testing.T) {
	g := sharding.Genesis{}

	g.InitialBalances = make([]*sharding.InitialBalance, 2)
	g.InitialBalances[0] = &sharding.InitialBalance{}
	g.InitialBalances[1] = &sharding.InitialBalance{}

	g.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := g.ProcessConfig()

	assert.NotNil(t, g)
	assert.Equal(t, sharding.ErrCouldNotParsePubKey, err)
}

func TestGenesis_GenesisWithIncompleteBalance(t *testing.T) {
	g := sharding.Genesis{}

	g.InitialBalances = make([]*sharding.InitialBalance, 1)
	g.InitialBalances[0] = &sharding.InitialBalance{}

	g.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	_ = g.ProcessConfig()

	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)
	adrConv := mock.NewAddressConverterFake(32, "")

	inBal, err := g.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, g)
	assert.Nil(t, err)
	for _, val := range inBal {
		assert.Equal(t, big.NewInt(0), val)
	}
}

func TestGenesis_InitialNodesBalancesNil(t *testing.T) {
	genesis := sharding.Genesis{}
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)
	adrConv := mock.NewAddressConverterFake(32, "")
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, genesis)
	assert.Equal(t, 0, len(inBalance))
	assert.Nil(t, err)
}

func TestGenesis_InitialNodesBalancesNilShardCoordinatorShouldErr(t *testing.T) {
	genesis := createGenesisOneShardOneNode()
	adrConv := mock.NewAddressConverterFake(32, "")
	inBalance, err := genesis.InitialNodesBalances(nil, adrConv)

	assert.NotNil(t, genesis)
	assert.Nil(t, inBalance)
	assert.Equal(t, sharding.ErrNilShardCoordinator, err)
}

func TestGenesis_InitialNodesBalancesNilAddrConverterShouldErr(t *testing.T) {
	genesis := createGenesisOneShardOneNode()
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator, nil)

	assert.NotNil(t, genesis)
	assert.Nil(t, inBalance)
	assert.Equal(t, sharding.ErrNilAddressConverter, err)
}

func TestGenesis_InitialNodesBalancesGood(t *testing.T) {
	genesis := createGenesisTwoShardTwoNodes()
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(2, 1)
	adrConv := mock.NewAddressConverterFake(32, "")
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, genesis)
	assert.Equal(t, 2, len(inBalance))
	assert.Nil(t, err)
}

func TestGenesis_Initial5NodesBalancesGood(t *testing.T) {
	genesis := createGenesisTwoShard6NodesMeta()
	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(2, 1)
	adrConv := mock.NewAddressConverterFake(32, "")
	inBalance, err := genesis.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, genesis)
	assert.Equal(t, 3, len(inBalance))
	assert.Nil(t, err)
}
