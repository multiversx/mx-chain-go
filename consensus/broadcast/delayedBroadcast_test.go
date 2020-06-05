package broadcast_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validatorDelayArgs struct {
	headerHash      []byte
	prevRandSeed    []byte
	round           uint64
	miniBlocks      map[uint32][]byte
	miniBlockHashes map[uint32]map[string]struct{}
	transactions    map[string][][]byte
	order           uint32
}

func createValidatorDelayArgs(index int) *validatorDelayArgs {
	iStr := strconv.Itoa(index)
	return &validatorDelayArgs{
		headerHash:      []byte("header hash" + iStr),
		prevRandSeed:    []byte("prev rand seed" + iStr),
		round:           uint64(0),
		miniBlocks:      map[uint32][]byte{0: []byte("miniblock data sh0" + iStr), 1: []byte("miniblock data sh1" + iStr)},
		miniBlockHashes: map[uint32]map[string]struct{}{0: {"miniBlockHash0" + iStr: struct{}{}}, 1: {"miniBlockHash1" + iStr: struct{}{}}},
		transactions:    map[string][][]byte{"topic": {[]byte("tx1" + iStr), []byte("tx2" + iStr)}},
		order:           uint32(1),
	}
}

func getDataToNotifyFromArgs(args *validatorDelayArgs) map[uint32][][]byte {
	mbHashesToNotify := make(map[uint32][][]byte, 0)

	for shardIdDest, hashesMap := range args.miniBlockHashes {
		for mbHash := range hashesMap {
			mbHashesToNotify[shardIdDest] = append(mbHashesToNotify[shardIdDest], []byte(mbHash))
		}
	}

	return mbHashesToNotify
}

func createMetaBlock() *block.MetaBlock {
	return &block.MetaBlock{
		Nonce: 0,
		Round: 0,
		ShardInfo: []block.ShardData{
			{
				HeaderHash: []byte("shard0 headerHash"),
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						Hash:            []byte("miniblock hash"),
						ReceiverShardID: 1,
						SenderShardID:   0,
						TxCount:         2,
					},
				},
				ShardID: 0,
			},
		},
	}
}

func createDefaultDelayedBroadcasterArgs() *broadcast.DelayedBlockBroadcasterArgs {
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptorsContainer := createInterceptorContainer()
	dbbArgs := &broadcast.DelayedBlockBroadcasterArgs{
		ShardCoordinator:      &mock.ShardCoordinatorMock{},
		InterceptorsContainer: interceptorsContainer,
		HeadersSubscriber:     headersSubscriber,
		LeaderCacheSize:       2,
		ValidatorCacheSize:    2,
	}

	return dbbArgs
}

func TestNewDelayedBlockBroadcaster_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	delayBroadcasterArgs.ShardCoordinator = nil
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Equal(t, spos.ErrNilShardCoordinator, err)
	require.Nil(t, dbb)
}

func TestNewDelayedBlockBroadcaster_NilHeadersSubscriberShouldErr(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	delayBroadcasterArgs.HeadersSubscriber = nil
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Equal(t, spos.ErrNilHeadersSubscriber, err)
	require.Nil(t, dbb)
}

func TestNewDelayedBlockBroadcaster_NilInterceptorsContainerShouldErr(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	delayBroadcasterArgs.InterceptorsContainer = nil
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Equal(t, spos.ErrNilInterceptorsContainer, err)
	require.Nil(t, dbb)
}

func TestNewDelayedBlockBroadcasterOK(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)
	require.NotNil(t, dbb)
}

func TestDelayedBlockBroadcaster_HeaderReceivedNoDelayedDataRegistered(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Flag{}
	txBroadcastCalled := atomic.Flag{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Set()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Set()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	metaBlock := createMetaBlock()

	dbb.HeaderReceived(metaBlock, []byte("metablock hash"))
	time.Sleep(10 * time.Millisecond)
	assert.False(t, mbBroadcastCalled.IsSet())
	assert.False(t, txBroadcastCalled.IsSet())
}

func TestDelayedBlockBroadcaster_HeaderReceivedForRegisteredDelayedDataShouldBroadcastTheData(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Flag{}
	txBroadcastCalled := atomic.Flag{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Set()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Set()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	headerHash, miniblocksData, transactionsData := createDelayData("1")
	delayedData := broadcast.CreateDelayBroadcastDataForLeader(headerHash, miniblocksData, transactionsData)
	err = dbb.SetLeaderData(delayedData)

	metaBlock := createMetaBlock()
	metaBlock.ShardInfo[0].HeaderHash = headerHash

	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, mbBroadcastCalled.IsSet())
	assert.False(t, txBroadcastCalled.IsSet())

	dbb.HeaderReceived(metaBlock, []byte("meta hash"))
	time.Sleep(core.ExtraDelayForBroadcastBlockInfo + 10*time.Millisecond)
	assert.True(t, mbBroadcastCalled.IsSet())
	assert.True(t, txBroadcastCalled.IsSet())
}

func TestDelayedBlockBroadcaster_HeaderReceivedForNotRegisteredDelayedDataShouldNotBroadcast(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Flag{}
	txBroadcastCalled := atomic.Flag{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Set()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Set()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	headerHash, miniblocksData, transactionsData := createDelayData("1")
	delayedData := broadcast.CreateDelayBroadcastDataForLeader(headerHash, miniblocksData, transactionsData)
	err = dbb.SetLeaderData(delayedData)
	require.Nil(t, err)

	metaBlock := createMetaBlock()
	metaBlock.ShardInfo[0].HeaderHash = headerHash[1:]

	err = dbb.SetLeaderData(delayedData)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, mbBroadcastCalled.IsSet())
	assert.False(t, txBroadcastCalled.IsSet())

	dbb.HeaderReceived(metaBlock, []byte("meta hash"))
	time.Sleep(10 * time.Millisecond)
	assert.False(t, mbBroadcastCalled.IsSet())
	assert.False(t, txBroadcastCalled.IsSet())
}

func TestDelayedBlockBroadcaster_HeaderReceivedForNextRegisteredDelayedDataShouldBroadcastBoth(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	headerHash, miniblocksData, transactionsData := createDelayData("1")
	delayedData := broadcast.CreateDelayBroadcastDataForLeader(headerHash, miniblocksData, transactionsData)
	err = dbb.SetLeaderData(delayedData)
	require.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(0), mbBroadcastCalled.Get())
	assert.Equal(t, int64(0), txBroadcastCalled.Get())

	headerHash2, miniBlockData2, transactionsData2 := createDelayData("2")
	delayedData = broadcast.CreateDelayBroadcastDataForLeader(headerHash2, miniBlockData2, transactionsData2)
	err = dbb.SetLeaderData(delayedData)
	require.Nil(t, err)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(0), mbBroadcastCalled.Get())
	assert.Equal(t, int64(0), txBroadcastCalled.Get())

	metaBlock := createMetaBlock()
	metaBlock.ShardInfo[0].HeaderHash = headerHash2

	dbb.HeaderReceived(metaBlock, []byte("meta hash"))
	time.Sleep(core.ExtraDelayForBroadcastBlockInfo + 10*time.Millisecond)
	assert.Equal(t, int64(2), mbBroadcastCalled.Get())
	assert.Equal(t, int64(2), txBroadcastCalled.Get())

	vbd := dbb.GetValidatorBroadcastData()
	assert.Equal(t, 0, len(vbd))
}

func TestDelayedBlockBroadcaster_SetLeaderDataNilDataShouldErr(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetLeaderData(nil)
	require.Equal(t, spos.ErrNilParameter, err)
}

func TestDelayedBlockBroadcaster_SetLeaderData(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	headerHash, miniBlockData, transactionsData := createDelayData("1")
	delayedData := broadcast.CreateDelayBroadcastDataForLeader(headerHash, miniBlockData, transactionsData)
	err = dbb.SetLeaderData(delayedData)
	require.Nil(t, err)

	vbb := dbb.GetLeaderBroadcastData()
	require.Equal(t, 1, len(vbb))
}

func TestDelayedBlockBroadcaster_SetValidatorDataNilDataShouldErr(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetValidatorData(nil)
	require.Equal(t, spos.ErrNilParameter, err)

	vbb := dbb.GetLeaderBroadcastData()
	require.Equal(t, 0, len(vbb))
}

func TestDelayedBlockBroadcaster_SetValidatorData(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)
	require.Nil(t, err)

	vbb := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbb))
}

func TestDelayedBlockBroadcaster_SetValidatorDelayBroadcastAccumulatedDataBounded(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	vbd := dbb.GetValidatorBroadcastData()
	expectedLen := 0
	require.Equal(t, expectedLen, len(vbd))

	for i := 1; i < 100; i++ {
		vArgs := createValidatorDelayArgs(i)
		delayedData := broadcast.CreateDelayBroadcastDataForValidator(
			vArgs.headerHash,
			vArgs.prevRandSeed,
			vArgs.round,
			vArgs.miniBlocks,
			vArgs.miniBlockHashes,
			vArgs.transactions,
			vArgs.order,
		)

		err = dbb.SetValidatorData(delayedData)
		require.Nil(t, err)

		vbd = dbb.GetValidatorBroadcastData()
		expectedLen = i
		if i > int(delayBroadcasterArgs.ValidatorCacheSize) {
			expectedLen = int(delayBroadcasterArgs.ValidatorCacheSize)
		}
		require.Equal(t, expectedLen, len(vbd))
	}
}

func TestDelayedBlockBroadcaster_ScheduleValidatorBroadcastDifferentHeaderRoundShouldDoNothing(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)

	require.Nil(t, err)
	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round + 1,
		PrevRandSeed: vArgs.prevRandSeed,
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was no broadcast and validator delay data still present
	require.Equal(t, int64(0), mbBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
}

func TestDelayedBlockBroadcaster_ScheduleValidatorBroadcastDifferentPrevRandShouldDoNothing(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)

	require.Nil(t, err)
	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	differentPrevRandSeed := make([]byte, len(vArgs.prevRandSeed))
	copy(differentPrevRandSeed, vArgs.prevRandSeed)
	differentPrevRandSeed[0] = ^differentPrevRandSeed[0]
	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round,
		PrevRandSeed: differentPrevRandSeed,
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was no broadcast and validator delay data still present
	require.Equal(t, int64(0), mbBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
}

func TestDelayedBlockBroadcaster_ScheduleValidatorBroadcastSameRoundAndPrevRandShouldBroadcast(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)

	require.Nil(t, err)
	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round,
		PrevRandSeed: vArgs.prevRandSeed,
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was a broadcast and validator delay data empty
	require.Equal(t, int64(1), mbBroadcastCalled.Get())
	require.Equal(t, int64(1), txBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 0, len(vbd))
}

func TestDelayedBlockBroadcaster_AlarmExpiredShouldBroadcastTheDataForRegisteredDelayedData(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)

	require.Nil(t, err)
	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	dbb.AlarmExpired(string(vArgs.headerHash))
	time.Sleep(time.Millisecond * 100)

	// check there was a broadcast and validator delay data empty
	require.Equal(t, int64(1), mbBroadcastCalled.Get())
	require.Equal(t, int64(1), txBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 0, len(vbd))
}

func TestDelayedBlockBroadcaster_AlarmExpiredShouldDoNothingForNotRegisteredData(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)

	require.Nil(t, err)
	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	differentHeaderHash := make([]byte, len(vArgs.headerHash))
	copy(differentHeaderHash, vArgs.headerHash)
	differentHeaderHash[0] = ^differentHeaderHash[0]
	dbb.AlarmExpired(string(differentHeaderHash))
	time.Sleep(time.Millisecond * 100)

	// check there was no broadcast and validator delay data still present
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
}

func TestDelayedBlockBroadcaster_RegisterInterceptorCallback(t *testing.T) {
	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	var cbs []func(toShard uint32, data []byte)
	mutCbs := &sync.Mutex{}

	registerHandler := func(handler func(toShard uint32, data []byte)) {
		mutCbs.Lock()
		cbs = append(cbs, handler)
		mutCbs.Unlock()
	}

	delayBroadcasterArgs.InterceptorsContainer = &mock.InterceptorsContainerStub{
		GetCalled: func(topic string) (process.Interceptor, error) {
			return &mock.InterceptorStub{
				RegisterHandlerCalled: registerHandler,
			}, nil
		},
	}

	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)

	mutCbs.Lock()
	nbRegistered := len(cbs)
	mutCbs.Unlock()
	require.Equal(t, 1, nbRegistered)

	err = dbb.SetValidatorData(delayedData)
	require.Nil(t, err)

	cb := func(toShard uint32, data []byte) {}
	err = dbb.RegisterInterceptorCallback(cb)
	require.Nil(t, err)

	mutCbs.Lock()
	nbRegistered = len(cbs)
	mutCbs.Unlock()
	require.Equal(t, 2, nbRegistered)
}

func TestDelayedBlockBroadcaster_GetShardDataFromMetaChainBlock(t *testing.T) {
	metaHeader := createMetaBlock()
	shardID := uint32(0)

	expHeaderHashes := make([][]byte, 0)
	valData := make([]*broadcast.HeaderDataForValidator, 0)

	for _, shInfo := range metaHeader.ShardInfo {
		if shInfo.ShardID != shardID {
			continue
		}
		valData = append(valData, &broadcast.HeaderDataForValidator{
			Round:        shInfo.Round,
			PrevRandSeed: shInfo.PrevRandSeed,
		})
		expHeaderHashes = append(expHeaderHashes, shInfo.HeaderHash)
	}

	headerHashes, dfv, err := broadcast.GetShardDataFromMetaChainBlock(metaHeader, shardID)
	require.Nil(t, err)
	require.Equal(t, expHeaderHashes, headerHashes)
	require.Equal(t, valData, dfv)
}

func TestDelayedBlockBroadcaster_InterceptedMiniBlockForNotSetValDataShouldBroadcast(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)
	require.Nil(t, err)

	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round,
		PrevRandSeed: vArgs.prevRandSeed,
	}

	dbb.InterceptedMiniBlockData(1, []byte("some other miniBlock hash"))
	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was a broadcast and validator delay data empty
	require.Equal(t, int64(1), mbBroadcastCalled.Get())
	require.Equal(t, int64(1), txBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 0, len(vbd))
}

func TestDelayedBlockBroadcaster_InterceptedMiniBlockOutOfManyForSetValDataShouldBroadcast(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	miniBlockHashToNotify := []byte("miniBlockHash to notify")
	destShardID := uint32(1)
	vArgs.miniBlockHashes[destShardID][string(miniBlockHashToNotify)] = struct{}{}
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)
	require.Nil(t, err)

	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round,
		PrevRandSeed: vArgs.prevRandSeed,
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	dbb.InterceptedMiniBlockData(destShardID, miniBlockHashToNotify)
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was a broadcast and validator delay data empty
	require.Equal(t, int64(1), mbBroadcastCalled.Get())
	require.Equal(t, int64(1), txBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 0, len(vbd))
}

func TestDelayedBlockBroadcaster_InterceptedMiniBlockFinalForSetValDataShouldNotBroadcast(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}
	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	mbHahsesToNotify := getDataToNotifyFromArgs(vArgs)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)
	require.Nil(t, err)

	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round,
		PrevRandSeed: vArgs.prevRandSeed,
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	for destShardID, hashes := range mbHahsesToNotify {
		for _, hash := range hashes {
			dbb.InterceptedMiniBlockData(destShardID, hash)
		}
	}
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was no broadcast and validator delay data empty
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 0, len(vbd))
}

func TestDelayedBlockBroadcaster_Close(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}
	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)
	require.Nil(t, err)

	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round,
		PrevRandSeed: vArgs.prevRandSeed,
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	dbb.Close()

	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was no broadcast
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
}
