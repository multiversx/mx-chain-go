package process

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
)

func TestNewSovereignOutportDataProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil input", func(t *testing.T) {
		sodp, err := NewSovereignOutportDataProvider(nil)
		require.Nil(t, sodp)
		require.Equal(t, process.ErrNilOutportDataProvider, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createArgOutportDataProvider()
		odp, _ := NewOutportDataProvider(args)
		sodp, err := NewSovereignOutportDataProvider(odp)
		require.Nil(t, err)
		require.False(t, odp.IsInterfaceNil())
		require.Equal(t, "*process.sovereignOutportDataProvider", fmt.Sprintf("%T", sodp.shardRewardsCreator))
	})
}

func TestSovereignOutportDataProvider_PrepareOutportSaveBlockData(t *testing.T) {
	t.Parallel()

	reward1 := &rewardTx.RewardTx{
		Round: 4,
	}
	reward2 := &rewardTx.RewardTx{
		Round: 5,
	}

	hash1 := "hash1"
	hash2 := "hash2"

	rewardTxs := map[string]data.TransactionHandler{
		hash1: reward1,
		hash2: reward2,
	}

	args := createArgOutportDataProvider()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{}
	args.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		GetAllCurrentUsedTxsCalled: func(blockType block.Type) map[string]data.TransactionHandler {
			require.NotEqual(t, block.RewardsBlock, blockType)
			return nil
		},
	}
	odp, _ := NewOutportDataProvider(args)
	sodp, _ := NewSovereignOutportDataProvider(odp)

	res, err := sodp.PrepareOutportSaveBlockData(ArgPrepareOutportSaveBlockData{
		Header: &block.Header{},
		Body: &block.Body{
			MiniBlocks: []*block.MiniBlock{},
		},
		RewardsTxs: rewardTxs,
		HeaderHash: []byte("hdrHash"),
	})
	require.Nil(t, err)
	require.Equal(t, map[string]*outportcore.RewardInfo{
		hex.EncodeToString([]byte(hash1)): {
			Reward: reward1,
		},
		hex.EncodeToString([]byte(hash2)): {
			Reward: reward2,
		},
	}, res.TransactionPool.Rewards)
}
