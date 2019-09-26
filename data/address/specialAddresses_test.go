package address

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Args struct {
	ElrondCommunityAddress []byte
	BurnAddress            []byte
	AddrConv               state.AddressConverter
	ShardCoordinator       sharding.Coordinator
	NodesCoordiator        sharding.NodesCoordinator
}

func initDefaultArgs() *Args {
	args := &Args{
		ElrondCommunityAddress: []byte("community"),
		BurnAddress:            []byte("burn"),
		AddrConv:               &mock.AddressConverterMock{},
		ShardCoordinator:       mock.NewMultiShardsCoordinatorMock(1),
		NodesCoordiator:        mock.NewNodesCoordinatorMock(),
	}

	return args
}

func createSpecialAddressFromArgs(args *Args) (process.SpecialAddressHandler, error) {
	addr, err := NewSpecialAddressHolder(
		args.ElrondCommunityAddress,
		args.BurnAddress,
		args.AddrConv,
		args.ShardCoordinator,
		args.NodesCoordiator,
	)
	return addr, err
}

func createDefaultSpecialAddress() process.SpecialAddressHandler {
	args := initDefaultArgs()
	addr, _ := createSpecialAddressFromArgs(args)

	return addr
}

func TestNewSpecialAddressHolderNilCommunityAddressShouldErr(t *testing.T) {
	t.Parallel()

	args := initDefaultArgs()
	args.ElrondCommunityAddress = nil
	addr, err := createSpecialAddressFromArgs(args)

	assert.Nil(t, addr)
	assert.Equal(t, data.ErrNilElrondAddress, err)
}

func TestNewSpecialAddressHolderNilBurnAddressShouldErr(t *testing.T) {
	t.Parallel()

	args := initDefaultArgs()
	args.BurnAddress = nil
	addr, err := createSpecialAddressFromArgs(args)

	assert.Nil(t, addr)
	assert.Equal(t, data.ErrNilBurnAddress, err)
}

func TestNewSpecialAddressHolderNilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := initDefaultArgs()
	args.AddrConv = nil
	addr, err := createSpecialAddressFromArgs(args)

	assert.Nil(t, addr)
	assert.Equal(t, data.ErrNilAddressConverter, err)
}

func TestNewSpecialAddressHolderNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := initDefaultArgs()
	args.ShardCoordinator = nil
	addr, err := createSpecialAddressFromArgs(args)

	assert.Nil(t, addr)
	assert.Equal(t, data.ErrNilShardCoordinator, err)
}

func TestNewSpecialAddressHolderNilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := initDefaultArgs()
	args.NodesCoordiator = nil
	addr, err := createSpecialAddressFromArgs(args)

	assert.Nil(t, addr)
	assert.Equal(t, data.ErrNilNodesCoordinator, err)
}

func TestNewSpecialAddressHolderOK(t *testing.T) {
	t.Parallel()

	args := initDefaultArgs()
	addr, err := createSpecialAddressFromArgs(args)

	assert.NotNil(t, addr)
	assert.Nil(t, err)
}

func TestSpecialAddresses_ClearMetaConsensusDataOK(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()

	addr.ClearMetaConsensusData()
	metaConsensusData := addr.ConsensusMetaRewardData()

	assert.Equal(t, 0, len(metaConsensusData))
}

func TestSpecialAddresses_SetMetaConsensusDataSettingOnceOK(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()

	err := addr.SetMetaConsensusData([]byte("randomness"), 0, 0)
	assert.Nil(t, err)
}

func TestSpecialAddresses_SetMetaConsensusDataSettingMultipleOK(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	nConsensuses := 10

	for i := 0; i < nConsensuses; i++ {
		err := addr.SetMetaConsensusData([]byte("randomness"), uint64(i), 0)
		assert.Nil(t, err)
	}
}

func TestSpecialAddresses_ConsensusMetaRewardDataNoConsensusData(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	metaConsensusData := addr.ConsensusMetaRewardData()

	assert.Equal(t, 0, len(metaConsensusData))
}

func TestSpecialAddresses_ConsensusMetaRewardDataOneConsensusDataOK(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()

	_ = addr.SetMetaConsensusData([]byte("randomness"), 1, 2)
	metaConsensusData := addr.ConsensusMetaRewardData()

	assert.Equal(t, 1, len(metaConsensusData))
	assert.Equal(t, uint64(1), metaConsensusData[0].Round)
	assert.Equal(t, uint32(2), metaConsensusData[0].Epoch)
}

func TestSpecialAddresses_ConsensusMetaRewardDataMultipleConsensusesDataOK(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	nConsensuses := 10

	for i := 0; i < nConsensuses; i++ {
		_ = addr.SetMetaConsensusData([]byte("randomness"), uint64(i+1), uint32(i+2))
	}

	metaConsensusData := addr.ConsensusMetaRewardData()
	assert.Equal(t, nConsensuses, len(metaConsensusData))

	for i := 0; i < nConsensuses; i++ {
		assert.Equal(t, uint64(i+1), metaConsensusData[i].Round)
		assert.Equal(t, uint32(i+2), metaConsensusData[i].Epoch)
	}
}

func TestSpecialAddresses_ConsensusShardRewardDataNoData(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	shardRewardData := addr.ConsensusShardRewardData()

	assert.Nil(t, shardRewardData)
}

func TestSpecialAddresses_ConsensusShardRewardDataExistingData(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	_ = addr.SetShardConsensusData([]byte("randomness"), 1, 2, 0)
	shardRewardData := addr.ConsensusShardRewardData()

	assert.NotNil(t, shardRewardData)
	assert.Equal(t, uint64(1), shardRewardData.Round)
	assert.Equal(t, uint32(2), shardRewardData.Epoch)
}

func TestSpecialAddresses_SetShardConsensusData(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	err := addr.SetShardConsensusData([]byte("randomness"), 1, 2, 0)

	assert.Nil(t, err)
}

func TestSpecialAddresses_BurnAddress(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	burnAddr := addr.BurnAddress()

	assert.Equal(t, []byte("burn"), burnAddr)
}

func TestSpecialAddresses_ElrondCommunityAddress(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	communityAddr := addr.ElrondCommunityAddress()

	assert.Equal(t, []byte("community"), communityAddr)
}

func TestSpecialAddresses_LeaderAddressNoSetShardConsensusData(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	leaderAddress := addr.LeaderAddress()

	assert.Nil(t, leaderAddress)
}

func TestSpecialAddresses_LeaderAddressSetShardConsensusData(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	_ = addr.SetShardConsensusData([]byte("randomness"), 0, 0, 0)
	leaderAddress := addr.LeaderAddress()

	assert.Equal(t, "address00", string(leaderAddress))
}

func TestSpecialAddresses_Round(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	_ = addr.SetShardConsensusData([]byte("randomness"), 1, 2, 0)
	round := addr.Round()

	assert.Equal(t, uint64(0), round)
}

func TestSpecialAddresses_Epoch(t *testing.T) {
	t.Parallel()

	addr := createDefaultSpecialAddress()
	_ = addr.SetShardConsensusData([]byte("randomness"), 1, 2, 0)
	epoch := addr.Epoch()

	assert.Equal(t, uint32(2), epoch)
}

func TestSpecialAddresses_SetElrondCommunityAddress(t *testing.T) {
	addr := createDefaultSpecialAddress()
	communityAddress := addr.ElrondCommunityAddress()

	assert.Equal(t, []byte("community"), communityAddress)
}

func TestSpecialAddresses_ShardIdForAddress(t *testing.T) {
	args := initDefaultArgs()
	args.ShardCoordinator = &mock.MultipleShardsCoordinatorMock{
		NoShards: 4,
		ComputeIdCalled: func(address state.AddressContainer) uint32 {
			return uint32(address.Bytes()[0])
		},
		CurrentShard: 0,
	}
	addr, _ := createSpecialAddressFromArgs(args)
	shardId, err := addr.ShardIdForAddress([]byte{3})

	assert.Nil(t, err)
	assert.Equal(t, uint32(3), shardId)
}

func TestSpecialAddresses_IsInterfaceNil(t *testing.T) {
	addr := &specialAddresses{}

	addr = nil
	isNil := addr.IsInterfaceNil()

	assert.True(t, isNil)
}
