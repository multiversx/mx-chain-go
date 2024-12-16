package epochStartTrigger

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

func TestCheckNilArgs(t *testing.T) {
	t.Parallel()

	t.Run("nil data comps", func(t *testing.T) {
		args := createArgs(0)
		args.DataComps = nil
		err := checkNilArgs(args)
		require.Equal(t, process.ErrNilDataComponentsHolder, err)
	})
	t.Run("nil data pool", func(t *testing.T) {
		args := createArgs(0)
		dataComps := createDataCompsMock()
		dataComps.DataPool = nil
		args.DataComps = dataComps
		err := checkNilArgs(args)
		require.Equal(t, process.ErrNilDataPoolHolder, err)
	})
	t.Run("nil blockchain", func(t *testing.T) {
		args := createArgs(0)
		dataComps := createDataCompsMock()
		dataComps.BlockChain = nil
		args.DataComps = dataComps
		err := checkNilArgs(args)
		require.Equal(t, process.ErrNilBlockChain, err)
	})
	t.Run("nil mb pool", func(t *testing.T) {
		args := createArgs(0)
		dataComps := createDataCompsMock()
		dataPool := createDataPoolMock()
		dataPool.MiniBlocksCalled = func() storage.Cacher {
			return nil
		}
		dataComps.DataPool = dataPool
		args.DataComps = dataComps
		err := checkNilArgs(args)
		require.Equal(t, dataRetriever.ErrNilMiniblocksPool, err)
	})
	t.Run("nil validator pool", func(t *testing.T) {
		args := createArgs(0)
		dataComps := createDataCompsMock()
		dataPool := createDataPoolMock()
		dataPool.ValidatorsInfoCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return nil
		}
		dataComps.DataPool = dataPool
		args.DataComps = dataComps
		err := checkNilArgs(args)
		require.Equal(t, process.ErrNilValidatorInfoPool, err)
	})
	t.Run("nil bootstrap comps", func(t *testing.T) {
		args := createArgs(0)
		args.BootstrapComponents = nil
		err := checkNilArgs(args)
		require.Equal(t, process.ErrNilBootstrapComponentsHolder, err)
	})
	t.Run("nil shard coordinator", func(t *testing.T) {
		args := createArgs(0)
		bootStrapComps := createBootstrapComps(0)
		bootStrapComps.ShardCoordinatorCalled = nil
		args.BootstrapComponents = bootStrapComps
		err := checkNilArgs(args)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("nil request handler", func(t *testing.T) {
		args := createArgs(0)
		args.RequestHandler = nil
		err := checkNilArgs(args)
		require.Equal(t, process.ErrNilRequestHandler, err)
	})
}
