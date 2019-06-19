package sharding_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func createGenesisOneShardOneNode() *sharding.Genesis {
	genesis := &sharding.Genesis{}
	genesis.InitialBalances = make([]*sharding.InitialBalance, 1)
	genesis.InitialBalances[0] = &sharding.InitialBalance{}
	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	genesis.InitialBalances[0].Balance = "11"

	err := genesis.ProcessConfig()
	if err != nil {
		return nil
	}

	return genesis
}

func createGenesisTwoShardTwoNodes() *sharding.Genesis {
	genesis := &sharding.Genesis{}
	genesis.InitialBalances = make([]*sharding.InitialBalance, 4)
	genesis.InitialBalances[0] = &sharding.InitialBalance{}
	genesis.InitialBalances[1] = &sharding.InitialBalance{}
	genesis.InitialBalances[2] = &sharding.InitialBalance{}
	genesis.InitialBalances[3] = &sharding.InitialBalance{}

	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"
	genesis.InitialBalances[1].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7418"
	genesis.InitialBalances[2].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7417"
	genesis.InitialBalances[3].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7416"

	genesis.InitialBalances[0].Balance = "999"
	genesis.InitialBalances[1].Balance = "999"
	genesis.InitialBalances[2].Balance = "999"
	genesis.InitialBalances[3].Balance = "999"

	err := genesis.ProcessConfig()
	if err != nil {
		return nil
	}

	return genesis
}

func createGenesisTwoShard6NodesMeta() *sharding.Genesis {
	genesis := &sharding.Genesis{}
	genesis.InitialBalances = make([]*sharding.InitialBalance, 6)
	genesis.InitialBalances[0] = &sharding.InitialBalance{}
	genesis.InitialBalances[1] = &sharding.InitialBalance{}
	genesis.InitialBalances[2] = &sharding.InitialBalance{}
	genesis.InitialBalances[3] = &sharding.InitialBalance{}
	genesis.InitialBalances[4] = &sharding.InitialBalance{}
	genesis.InitialBalances[5] = &sharding.InitialBalance{}

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

	err := genesis.ProcessConfig()
	if err != nil {
		return nil
	}

	return genesis
}

func TestGenesis_NewGenesisConfigWrongFile(t *testing.T) {
	genesis, err := sharding.NewGenesisConfig("")

	assert.Nil(t, genesis)
	assert.NotNil(t, err)
}

func TestNodes_NewGenesisConfigWrongDataInFile(t *testing.T) {
	genesis, err := sharding.NewGenesisConfig("mock/invalidGenesisMock.json")

	assert.Nil(t, genesis)
	assert.Equal(t, sharding.ErrCouldNotParsePubKey, err)
}

func TestNodes_NewGenesisShouldWork(t *testing.T) {
	genesis, err := sharding.NewGenesisConfig("mock/genesisMock.json")

	assert.NotNil(t, genesis)
	assert.Nil(t, err)
}

func TestGenesis_ProcessConfigGenesisWithIncompleteDataShouldErr(t *testing.T) {
	genesis := sharding.Genesis{}

	genesis.InitialBalances = make([]*sharding.InitialBalance, 2)
	genesis.InitialBalances[0] = &sharding.InitialBalance{}
	genesis.InitialBalances[1] = &sharding.InitialBalance{}

	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	err := genesis.ProcessConfig()

	assert.NotNil(t, genesis)
	assert.Equal(t, sharding.ErrCouldNotParsePubKey, err)
}

func TestGenesis_GenesisWithIncompleteBalance(t *testing.T) {
	genesis := sharding.Genesis{}

	genesis.InitialBalances = make([]*sharding.InitialBalance, 1)
	genesis.InitialBalances[0] = &sharding.InitialBalance{}

	genesis.InitialBalances[0].PubKey = "5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419"

	_ = genesis.ProcessConfig()

	shardCoordinator := mock.NewMultipleShardsCoordinatorFake(1, 0)
	adrConv := mock.NewAddressConverterFake(32, "")

	inBal, err := genesis.InitialNodesBalances(shardCoordinator, adrConv)

	assert.NotNil(t, genesis)
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
