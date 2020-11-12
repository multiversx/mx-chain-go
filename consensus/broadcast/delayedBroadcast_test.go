package broadcast_test

import (
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/alarm"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validatorDelayArgs struct {
	headerHash       []byte
	header           data.HeaderHandler
	miniBlocks       map[uint32][]byte
	transactions     map[string][][]byte
	miniBlockHashes  map[string]map[string]struct{}
	metaMiniBlocks   map[uint32][]byte
	metaTransactions map[string][][]byte
	order            uint32
}

func createValidatorDelayArgs(index int) *validatorDelayArgs {
	iStr := strconv.Itoa(index)
	return &validatorDelayArgs{
		headerHash: []byte("header hash" + iStr),
		header: &block.Header{
			PrevRandSeed: []byte("prev rand seed" + iStr),
			Round:        uint64(0),
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash:            []byte("miniBlockHash0" + iStr),
					SenderShardID:   0,
					ReceiverShardID: 0,
				},
				{
					Hash:            []byte("miniBlockHash1" + iStr),
					SenderShardID:   0,
					ReceiverShardID: 1,
				},
			},
		},
		miniBlocks:       map[uint32][]byte{0: []byte("miniblock data sh0" + iStr), 1: []byte("miniblock data sh1" + iStr)},
		miniBlockHashes:  map[string]map[string]struct{}{"txBlockBodies_0": {"miniBlockHash0" + iStr: struct{}{}}, "txBlockBodies_0_1": {"miniBlockHash1" + iStr: struct{}{}}},
		transactions:     map[string][][]byte{"transactions_0_1": {[]byte("tx1" + iStr), []byte("tx2" + iStr)}},
		metaMiniBlocks:   map[uint32][]byte{0: []byte("meta miniblock data sh0" + iStr), 1: []byte("meta miniblock data sh1" + iStr)},
		metaTransactions: map[string][][]byte{"mtopic": {[]byte("mtx1" + iStr), []byte("mtx2" + iStr)}},
		order:            uint32(1),
	}
}

func getDataToNotifyFromArgs(args *validatorDelayArgs) map[string][][]byte {
	mbHashesToNotify := make(map[string][][]byte, 0)

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

func createDefaultDelayedBroadcasterArgs() *broadcast.ArgsDelayedBlockBroadcaster {
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptorsContainer := createInterceptorContainer()
	dbbArgs := &broadcast.ArgsDelayedBlockBroadcaster{
		ShardCoordinator:      &mock.ShardCoordinatorMock{},
		InterceptorsContainer: interceptorsContainer,
		HeadersSubscriber:     headersSubscriber,
		LeaderCacheSize:       2,
		ValidatorCacheSize:    2,
		AlarmScheduler:        alarm.NewAlarmScheduler(),
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

func TestNewDelayedBlockBroadcaster_NilAlarmSchedulerShouldErr(t *testing.T) {
	t.Parallel()

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	delayBroadcasterArgs.AlarmScheduler = nil
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Equal(t, spos.ErrNilAlarmScheduler, err)
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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	headerHash, _, miniblocksData, transactionsData := createDelayData("1")
	delayedData := broadcast.CreateDelayBroadcastDataForLeader(headerHash, miniblocksData, transactionsData)
	err = dbb.SetLeaderData(delayedData)

	metaBlock := createMetaBlock()
	metaBlock.ShardInfo[0].HeaderHash = headerHash

	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, mbBroadcastCalled.IsSet())
	assert.False(t, txBroadcastCalled.IsSet())

	dbb.HeaderReceived(metaBlock, []byte("meta hash"))
	sleepTime := core.ExtraDelayForBroadcastBlockInfo +
		core.ExtraDelayBetweenBroadcastMbsAndTxs +
		10*time.Millisecond
	time.Sleep(sleepTime)
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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	headerHash, _, miniblocksData, transactionsData := createDelayData("1")
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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	headerHash, _, miniblocksData, transactionsData := createDelayData("1")
	delayedData := broadcast.CreateDelayBroadcastDataForLeader(headerHash, miniblocksData, transactionsData)
	err = dbb.SetLeaderData(delayedData)
	require.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(0), mbBroadcastCalled.Get())
	assert.Equal(t, int64(0), txBroadcastCalled.Get())

	headerHash2, _, miniBlockData2, transactionsData2 := createDelayData("2")
	delayedData = broadcast.CreateDelayBroadcastDataForLeader(headerHash2, miniBlockData2, transactionsData2)
	err = dbb.SetLeaderData(delayedData)
	require.Nil(t, err)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(0), mbBroadcastCalled.Get())
	assert.Equal(t, int64(0), txBroadcastCalled.Get())

	metaBlock := createMetaBlock()
	metaBlock.ShardInfo[0].HeaderHash = headerHash2

	dbb.HeaderReceived(metaBlock, []byte("meta hash"))
	sleepTime := core.ExtraDelayForBroadcastBlockInfo +
		core.ExtraDelayBetweenBroadcastMbsAndTxs +
		10*time.Millisecond
	time.Sleep(sleepTime)
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

	headerHash, _, miniBlockData, transactionsData := createDelayData("1")
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
		vArgs.header,
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

func TestDelayedBlockBroadcaster_SetHeaderForValidatorShouldSetAlarmAndBroadcastHeader(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}
	headerBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}
	broadcastHeader := func(header data.HeaderHandler) error {
		headerBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	vArgs.header.SetSignature([]byte("agg sig"))
	valHeaderData := broadcast.CreateValidatorHeaderBroadcastData(
		vArgs.headerHash,
		vArgs.header,
		vArgs.metaMiniBlocks,
		vArgs.metaTransactions,
		vArgs.order,
	)
	err = dbb.SetHeaderForValidator(valHeaderData)
	require.Nil(t, err)

	vbb := dbb.GetValidatorHeaderBroadcastData()
	require.Equal(t, 1, len(vbb))
	require.Equal(t, int64(0), headerBroadcastCalled.Get())
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	sleepTime := broadcast.ValidatorDelayPerOrder()*time.Duration(vArgs.order) +
		time.Millisecond*100
	time.Sleep(sleepTime)

	// alarm expired and sent header
	require.Equal(t, int64(1), headerBroadcastCalled.Get())
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	vbb = dbb.GetValidatorHeaderBroadcastData()
	require.Equal(t, 0, len(vbb))
}

func TestDelayedBlockBroadcaster_SetValidatorDataFinalizedMetaHeaderShouldSetAlarmAndBroadcastHeaderAndData(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}
	headerBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}
	broadcastHeader := func(header data.HeaderHandler) error {
		headerBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	delayBroadcasterArgs.ShardCoordinator = mock.ShardCoordinatorMock{
		ShardID: core.MetachainShardId,
	}
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	vArgs.header.SetSignature([]byte("agg sig"))
	vArgs.header.SetShardID(core.MetachainShardId)
	valData := broadcast.CreateValidatorHeaderBroadcastData(
		vArgs.headerHash,
		vArgs.header,
		vArgs.metaMiniBlocks,
		vArgs.metaTransactions,
		vArgs.order,
	)
	err = dbb.SetHeaderForValidator(valData)
	require.Nil(t, err)

	vhbd := dbb.GetValidatorHeaderBroadcastData()
	require.Equal(t, 1, len(vhbd))
	require.Equal(t, int64(0), headerBroadcastCalled.Get())
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	sleepTime := broadcast.ValidatorDelayPerOrder()*time.Duration(vArgs.order) +
		core.ExtraDelayForBroadcastBlockInfo +
		core.ExtraDelayBetweenBroadcastMbsAndTxs +
		time.Millisecond*100
	time.Sleep(sleepTime)

	// alarm expired and sent header
	require.Equal(t, int64(1), headerBroadcastCalled.Get())
	require.Equal(t, int64(1), mbBroadcastCalled.Get())
	require.Equal(t, int64(1), txBroadcastCalled.Get())

	vhbd = dbb.GetValidatorHeaderBroadcastData()
	require.Equal(t, 0, len(vhbd))
}

func TestDelayedBlockBroadcaster_InterceptedHeaderShouldCancelAlarm(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}
	headerBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}
	broadcastHeader := func(header data.HeaderHandler) error {
		headerBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	delayBroadcasterArgs.ShardCoordinator = mock.ShardCoordinatorMock{
		ShardID: core.MetachainShardId,
	}
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	vArgs.header.SetSignature([]byte("agg sig"))
	vArgs.header.SetShardID(core.MetachainShardId)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)
	require.Nil(t, err)

	vbb := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbb))
	require.Equal(t, int64(0), headerBroadcastCalled.Get())
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	sleepTime := broadcast.ValidatorDelayPerOrder()*time.Duration(vArgs.order) +
		time.Second

	// should cancel alarm
	dbb.InterceptedHeaderData("headerTopic", vArgs.headerHash, vArgs.header)
	time.Sleep(sleepTime)

	require.Equal(t, int64(0), headerBroadcastCalled.Get())
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	vbb = dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbb))
}

func TestDelayedBlockBroadcaster_InterceptedHeaderShouldCancelAlarmForHeaderBroadcast(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}
	headerBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}
	broadcastHeader := func(header data.HeaderHandler) error {
		headerBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	delayBroadcasterArgs.ShardCoordinator = mock.ShardCoordinatorMock{
		ShardID: core.MetachainShardId,
	}
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	vArgs.header.SetSignature([]byte("agg sig"))
	vArgs.header.SetShardID(core.MetachainShardId)
	validatorHeaderBroadcastData := broadcast.CreateValidatorHeaderBroadcastData(
		vArgs.headerHash,
		vArgs.header,
		vArgs.metaMiniBlocks,
		vArgs.metaTransactions,
		vArgs.order,
	)
	err = dbb.SetHeaderForValidator(validatorHeaderBroadcastData)
	require.Nil(t, err)

	vhbd := dbb.GetValidatorHeaderBroadcastData()
	require.Equal(t, 1, len(vhbd))
	require.Equal(t, int64(0), headerBroadcastCalled.Get())
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	sleepTime := broadcast.ValidatorDelayPerOrder()*time.Duration(vArgs.order) +
		time.Second

	// should cancel alarm
	dbb.InterceptedHeaderData("headerTopic", vArgs.headerHash, vArgs.header)
	time.Sleep(sleepTime)

	require.Equal(t, int64(0), headerBroadcastCalled.Get())
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	vhbd = dbb.GetValidatorHeaderBroadcastData()
	require.Equal(t, 0, len(vhbd))
}

func TestDelayedBlockBroadcaster_InterceptedHeaderInvalidOrDifferentShouldIgnore(t *testing.T) {
	t.Parallel()

	mbBroadcastCalled := atomic.Counter{}
	txBroadcastCalled := atomic.Counter{}
	headerBroadcastCalled := atomic.Counter{}

	broadcastMiniBlocks := func(mbData map[uint32][]byte) error {
		mbBroadcastCalled.Increment()
		return nil
	}
	broadcastTransactions := func(txData map[string][][]byte) error {
		txBroadcastCalled.Increment()
		return nil
	}
	broadcastHeader := func(header data.HeaderHandler) error {
		headerBroadcastCalled.Increment()
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	delayBroadcasterArgs.ShardCoordinator = mock.ShardCoordinatorMock{
		ShardID: core.MetachainShardId,
	}
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	vArgs.header.SetSignature([]byte("agg sig"))
	vArgs.header.SetShardID(core.MetachainShardId)
	valHeaderData := broadcast.CreateValidatorHeaderBroadcastData(
		vArgs.headerHash,
		vArgs.header,
		vArgs.metaMiniBlocks,
		vArgs.metaTransactions,
		vArgs.order,
	)
	err = dbb.SetHeaderForValidator(valHeaderData)
	require.Nil(t, err)

	vhbd := dbb.GetValidatorHeaderBroadcastData()
	require.Equal(t, 1, len(vhbd))
	require.Equal(t, int64(0), headerBroadcastCalled.Get())
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	invalidHeader := []byte("not a header")
	prevRandSeed := append([]byte(nil), vArgs.header.GetPrevRandSeed()...)
	prevRandSeed[0] = ^prevRandSeed[0]
	differentHeader := block.Header{
		PrevRandSeed: prevRandSeed,
		Round:        vArgs.header.GetRound() + 1,
	}
	headerHash := append([]byte(nil), vArgs.headerHash...)
	headerHash[0] = ^headerHash[0]

	// should not cancel alarm
	dbb.InterceptedHeaderData("headerTopic", headerHash, differentHeader)
	dbb.InterceptedMiniBlockData("headerTopic", headerHash, invalidHeader)
	sleepTime := broadcast.ValidatorDelayPerOrder()*time.Duration(vArgs.order) +
		core.ExtraDelayForBroadcastBlockInfo +
		core.ExtraDelayBetweenBroadcastMbsAndTxs +
		time.Millisecond*100
	time.Sleep(sleepTime)

	// alarm expired and sent header
	require.Equal(t, int64(1), headerBroadcastCalled.Get())
	require.Equal(t, int64(1), mbBroadcastCalled.Get())
	require.Equal(t, int64(1), txBroadcastCalled.Get())

	vhbd = dbb.GetValidatorHeaderBroadcastData()
	require.Equal(t, 0, len(vhbd))
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
			vArgs.header,
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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
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
		Round:        vArgs.header.GetRound() + 1,
		PrevRandSeed: vArgs.header.GetPrevRandSeed(),
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	sleepTime := time.Duration(vArgs.order)*broadcast.ValidatorDelayPerOrder() +
		100*time.Millisecond
	time.Sleep(sleepTime)

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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)

	require.Nil(t, err)
	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	differentPrevRandSeed := make([]byte, len(vArgs.header.GetPrevRandSeed()))
	copy(differentPrevRandSeed, vArgs.header.GetPrevRandSeed())
	differentPrevRandSeed[0] = ^differentPrevRandSeed[0]
	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.header.GetRound(),
		PrevRandSeed: differentPrevRandSeed,
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	sleepTime := time.Duration(vArgs.order)*broadcast.ValidatorDelayPerOrder() +
		100*time.Millisecond
	time.Sleep(sleepTime)

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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
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
		Round:        vArgs.header.GetRound(),
		PrevRandSeed: vArgs.header.GetPrevRandSeed(),
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	sleepTime := time.Duration(vArgs.order)*broadcast.ValidatorDelayPerOrder() +
		core.ExtraDelayForBroadcastBlockInfo +
		core.ExtraDelayBetweenBroadcastMbsAndTxs +
		100*time.Millisecond
	time.Sleep(sleepTime)

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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	err = dbb.SetValidatorData(delayedData)

	require.Nil(t, err)
	vbd := dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	dbb.AlarmExpired(hex.EncodeToString(vArgs.headerHash))
	sleepTime := core.ExtraDelayBetweenBroadcastMbsAndTxs +
		time.Millisecond*100
	time.Sleep(sleepTime)

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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
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
	var cbsHeader []func(topic string, hash []byte, data interface{})
	var cbsMiniblock []func(topic string, hash []byte, data interface{})
	mutCbs := &sync.Mutex{}

	registerHandlerHeaders := func(handler func(topic string, hash []byte, data interface{})) {
		mutCbs.Lock()
		cbsHeader = append(cbsHeader, handler)
		mutCbs.Unlock()
	}

	registerHandlerMiniblocks := func(handler func(topic string, hash []byte, data interface{})) {
		mutCbs.Lock()
		cbsMiniblock = append(cbsMiniblock, handler)
		mutCbs.Unlock()
	}

	delayBroadcasterArgs.InterceptorsContainer = &mock.InterceptorsContainerStub{
		GetCalled: func(topic string) (process.Interceptor, error) {
			var hdl func(handler func(topic string, hash []byte, data interface{}))
			switch topic {
			case "shardBlocks_0_META":
				hdl = registerHandlerHeaders
			case "txBlockBodies_0_1":
				hdl = registerHandlerMiniblocks
			default:
				return nil, errors.New("unexpected topic")
			}

			return &mock.InterceptorStub{
				RegisterHandlerCalled: hdl,
			}, nil
		},
	}

	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)

	mutCbs.Lock()
	nbRegisteredMbsHandlers := len(cbsMiniblock)
	nbRegisteredHeaderHandlers := len(cbsHeader)
	mutCbs.Unlock()
	require.Equal(t, 1, nbRegisteredMbsHandlers)
	require.Equal(t, 1, nbRegisteredHeaderHandlers)

	err = dbb.SetValidatorData(delayedData)
	require.Nil(t, err)

	cb := func(topic string, hash []byte, data interface{}) {}
	err = dbb.RegisterInterceptorCallback(cb)
	require.Nil(t, err)

	mutCbs.Lock()
	nbRegisteredMbsHandlers = len(cbsMiniblock)
	mutCbs.Unlock()
	require.Equal(t, 2, nbRegisteredMbsHandlers)
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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
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
		Round:        vArgs.header.GetRound(),
		PrevRandSeed: vArgs.header.GetPrevRandSeed(),
	}

	dbb.InterceptedMiniBlockData("txBlockBodies_0_1", []byte("some other miniBlock hash"), &block.MiniBlock{})
	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	sleepTime := time.Duration(vArgs.order)*broadcast.ValidatorDelayPerOrder() +
		core.ExtraDelayForBroadcastBlockInfo +
		core.ExtraDelayBetweenBroadcastMbsAndTxs +
		100*time.Millisecond
	time.Sleep(sleepTime)

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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	miniBlockHashToNotify := []byte("miniBlockHash to notify")
	destShardID := 1
	vArgs.miniBlockHashes["txBlockBodies_0_"+strconv.Itoa(destShardID)][string(miniBlockHashToNotify)] = struct{}{}
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
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
		Round:        vArgs.header.GetRound(),
		PrevRandSeed: vArgs.header.GetPrevRandSeed(),
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	dbb.InterceptedMiniBlockData("txBlockBodies_0_"+strconv.Itoa(destShardID), miniBlockHashToNotify, &block.MiniBlock{})
	sleepTime := time.Duration(vArgs.order)*broadcast.ValidatorDelayPerOrder() +
		core.ExtraDelayForBroadcastBlockInfo +
		core.ExtraDelayBetweenBroadcastMbsAndTxs +
		100*time.Millisecond
	time.Sleep(sleepTime)

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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	mbHahsesToNotify := getDataToNotifyFromArgs(vArgs)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
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
		Round:        vArgs.header.GetRound(),
		PrevRandSeed: vArgs.header.GetPrevRandSeed(),
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	for destShardID, hashes := range mbHahsesToNotify {
		for _, hash := range hashes {
			dbb.InterceptedMiniBlockData(destShardID, hash, &block.MiniBlock{})
		}
	}
	sleepTime := time.Duration(vArgs.order)*broadcast.ValidatorDelayPerOrder() +
		100*time.Millisecond
	time.Sleep(sleepTime)

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
	broadcastHeader := func(header data.HeaderHandler) error {
		return nil
	}

	delayBroadcasterArgs := createDefaultDelayedBroadcasterArgs()
	dbb, err := broadcast.NewDelayedBlockBroadcaster(delayBroadcasterArgs)
	require.Nil(t, err)

	err = dbb.SetBroadcastHandlers(broadcastMiniBlocks, broadcastTransactions, broadcastHeader)
	require.Nil(t, err)

	vArgs := createValidatorDelayArgs(0)
	delayedData := broadcast.CreateDelayBroadcastDataForValidator(
		vArgs.headerHash,
		vArgs.header,
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
		Round:        vArgs.header.GetRound(),
		PrevRandSeed: vArgs.header.GetPrevRandSeed(),
	}

	dbb.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	dbb.Close()

	sleepTime := time.Duration(vArgs.order)*broadcast.ValidatorDelayPerOrder() +
		100*time.Millisecond
	time.Sleep(sleepTime)

	// check there was no broadcast
	require.Equal(t, int64(0), mbBroadcastCalled.Get())
	require.Equal(t, int64(0), txBroadcastCalled.Get())

	vbd = dbb.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
}
