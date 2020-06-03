package broadcast_test

import (
	"bytes"
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
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDelayData(prefix string) ([]byte, map[uint32][]byte, map[string][][]byte) {
	miniblocks := make(map[uint32][]byte)
	receiverShardID := uint32(1)
	miniblocks[receiverShardID] = []byte(prefix + "miniblock data")

	transactions := make(map[string][][]byte)
	topic := "txBlockBodies_0_1"
	transactions[topic] = [][]byte{
		[]byte(prefix + "tx0"),
		[]byte(prefix + "tx1"),
	}
	headerHash := []byte(prefix + "header hash")

	return headerHash, miniblocks, transactions
}

func createInterceptorContainer() process.InterceptorsContainer {
	return &mock.InterceptorsContainerStub{
		GetCalled: func(topic string) (process.Interceptor, error) {
			switch topic {
			default:
			}
			return &mock.InterceptorStub{
				ProcessReceivedMessageCalled: func(message p2p.MessageP2P) error {
					return nil
				},
			}, nil
		},
	}
}

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
		miniBlocks:      map[uint32][]byte{0: []byte("miniblock A" + iStr), 1: []byte("miniblock" + iStr)},
		miniBlockHashes: map[uint32]map[string]struct{}{0: {"miniBlockHash0" + iStr: struct{}{}}, 1: {"miniBlockHash1" + iStr: struct{}{}}},
		transactions:    map[string][][]byte{"topic": {[]byte("tx1" + iStr), []byte("tx2" + iStr)}},
		order:           uint32(1),
	}
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

// elements in who and where need to be unique
func isIncluded(who [][]byte, where [][]byte) bool {
	initialLenWho := len(who)
	initialLenWhere := len(where)
	if initialLenWho > initialLenWhere {
		return false
	}

	cntEqual := 0
	for _, elem := range who {
		for _, elem2 := range where {
			if bytes.Equal(elem, elem2) {
				cntEqual++
				break
			}
		}
	}

	return cntEqual == initialLenWho
}

func createDefaultShardChainArgs() broadcast.ShardChainMessengerArgs {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptorsContainer := createInterceptorContainer()

	return broadcast.ShardChainMessengerArgs{
		CommonMessengerArgs: broadcast.CommonMessengerArgs{
			Marshalizer:           marshalizerMock,
			Messenger:             messengerMock,
			PrivateKey:            privateKeyMock,
			ShardCoordinator:      shardCoordinatorMock,
			SingleSigner:          singleSignerMock,
			InterceptorsContainer: interceptorsContainer,
		},
		HeadersSubscriber:          headersSubscriber,
		MaxDelayCacheSize:          1,
		MaxValidatorDelayCacheSize: 1,
	}
}

func TestShardChainMessenger_NewShardChainMessengerNilMarshalizerShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.Marshalizer = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilMessengerShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.Messenger = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilMessenger, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilPrivateKeyShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.PrivateKey = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilPrivateKey, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilShardCoordinatorShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.ShardCoordinator = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilSingleSignerShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.SingleSigner = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilSingleSigner, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilInterceptorsContainerShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.InterceptorsContainer = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilInterceptorsContainer, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilHeadersSubscriberShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.HeadersSubscriber = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilHeadersSubscriber, err)
}

func TestShardChainMessenger_NewShardChainMessengerShouldWork(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.NotNil(t, scm)
	assert.Equal(t, nil, err)
	assert.False(t, scm.IsInterfaceNil())
}

func TestShardChainMessenger_BroadcastBlockShouldErrNilBody(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastBlock(nil, &block.Header{})
	assert.Equal(t, spos.ErrNilBody, err)
}

func TestShardChainMessenger_BroadcastBlockShouldErrNilHeader(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastBlock(newTestBlockBody(), nil)
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestShardChainMessenger_BroadcastBlockShouldErrMockMarshalizer(t *testing.T) {
	marshalizer := mock.MarshalizerMock{
		Fail: true,
	}
	args := createDefaultShardChainArgs()
	args.Marshalizer = marshalizer
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastBlock(newTestBlockBody(), &block.Header{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestShardChainMessenger_BroadcastBlockShouldWork(t *testing.T) {
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastBlock(newTestBlockBody(), &block.Header{})
	assert.Nil(t, err)
}

func TestShardChainMessenger_BroadcastMiniBlocksShouldBeDone(t *testing.T) {
	channelCalled := make(chan bool, 100)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	miniBlocks := make(map[uint32][]byte)
	miniBlocks[0] = make([]byte, 0)
	miniBlocks[1] = make([]byte, 0)
	miniBlocks[2] = make([]byte, 0)
	miniBlocks[3] = make([]byte, 0)
	err := scm.BroadcastMiniBlocks(miniBlocks)

	called := 0
	for i := 0; i < 4; i++ {
		select {
		case <-channelCalled:
			called++
		case <-time.After(time.Millisecond * 100):
			break
		}
	}

	assert.Nil(t, err)
	assert.Equal(t, 4, called)
}

func TestShardChainMessenger_BroadcastTransactionsShouldNotBeCalled(t *testing.T) {
	channelCalled := make(chan bool)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	transactions := make(map[string][][]byte)
	err := scm.BroadcastTransactions(transactions)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.False(t, wasCalled)

	transactions[factory.TransactionTopic] = make([][]byte, 0)
	err = scm.BroadcastTransactions(transactions)

	wasCalled = false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.False(t, wasCalled)
}

func TestShardChainMessenger_BroadcastTransactionsShouldBeCalled(t *testing.T) {
	channelCalled := make(chan bool)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}

	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	transactions := make(map[string][][]byte)
	txs := make([][]byte, 0)
	txs = append(txs, []byte(""))
	transactions[factory.TransactionTopic] = txs
	err := scm.BroadcastTransactions(transactions)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestShardChainMessenger_BroadcastHeaderNilHeaderShouldErr(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastHeader(nil)
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestShardChainMessenger_BroadcastHeaderShouldWork(t *testing.T) {
	channelCalled := make(chan bool)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	hdr := block.MetaBlock{Nonce: 10}
	err := scm.BroadcastHeader(&hdr)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestShardChainMessenger_SetDataForDelayBroadcastNilHeaderHashShouldErr(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	_, miniblocks, transactions := createDelayData("1")

	err := scm.SetLeaderDelayBroadcast(nil, miniblocks, transactions)
	assert.Equal(t, spos.ErrNilHeaderHash, err)
}

func TestShardChainMessenger_SetDataForDelayBroadcastNilMiniblocksShouldReturnNil(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	headerHash, _, transactions := createDelayData("1")

	err := scm.SetLeaderDelayBroadcast(headerHash, nil, transactions)
	assert.Nil(t, err)
}

func TestShardChainMessenger_SetDataForDelayBroadcastShouldTriggerWaitingDelayedMessage(t *testing.T) {
	wasCalled := atomic.Flag{}
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			wasCalled.Set()
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	headerHash, miniBlocksMarshalled, transactions := createDelayData("1")
	err := scm.SetLeaderDelayBroadcast(headerHash, miniBlocksMarshalled, transactions)
	time.Sleep(10 * time.Millisecond)
	assert.Nil(t, err)
	assert.False(t, wasCalled.IsSet())

	wasCalled.Unset()
	headerHash2, miniBlocksMarshalled2, transactions2 := createDelayData("2")
	err = scm.SetLeaderDelayBroadcast(headerHash2, miniBlocksMarshalled2, transactions2)
	time.Sleep(10 * time.Millisecond)
	assert.Nil(t, err)
	assert.True(t, wasCalled.IsSet())
}

func TestShardChainMessenger_HeaderReceivedNoDelayedDataRegistered(t *testing.T) {
	wasCalled := atomic.Flag{}

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			wasCalled.Set()
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)
	metaBlock := createMetaBlock()

	scm.HeaderReceived(metaBlock, []byte("metablock hash"))
	time.Sleep(10 * time.Millisecond)
	assert.False(t, wasCalled.IsSet())
}

func TestShardChainMessenger_HeaderReceivedForRegisteredDelayedDataShouldBroadcastTheData(t *testing.T) {
	wasCalled := false
	mutData := &sync.Mutex{}
	broadcastBuffer := make([][]byte, 0)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutData.Lock()
			wasCalled = true
			broadcastBuffer = append(broadcastBuffer, buff)
			mutData.Unlock()
		},
	}

	args := createDefaultShardChainArgs()
	args.MaxDelayCacheSize = 2
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)
	headerHash, miniBlocksMarshalled, transactions := createDelayData("1")
	metaBlock := createMetaBlock()
	metaBlock.ShardInfo[0].HeaderHash = headerHash

	err := scm.SetLeaderDelayBroadcast(headerHash, miniBlocksMarshalled, transactions)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	mutData.Lock()
	assert.False(t, wasCalled)
	mutData.Unlock()

	scm.HeaderReceived(metaBlock, []byte("meta hash"))
	time.Sleep(core.ExtraDelayForBroadcastBlockInfo + 10*time.Millisecond)
	mutData.Lock()
	assert.True(t, wasCalled)
	assert.Contains(t, broadcastBuffer, miniBlocksMarshalled[1])
	mutData.Unlock()
}

func TestShardChainMessenger_HeaderReceivedForNotRegisteredDelayedDataShouldNotBroadcast(t *testing.T) {
	wasCalled := false
	mutData := &sync.Mutex{}
	broadcastBuffer := make([][]byte, 0)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutData.Lock()
			wasCalled = true
			broadcastBuffer = append(broadcastBuffer, buff)
			mutData.Unlock()
		},
	}

	args := createDefaultShardChainArgs()
	args.MaxDelayCacheSize = 2
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)
	headerHash, miniBlocksMarshalled, transactions := createDelayData("1")
	metaBlock := createMetaBlock()
	metaBlock.ShardInfo[0].HeaderHash = headerHash[1:]

	err := scm.SetLeaderDelayBroadcast(headerHash, miniBlocksMarshalled, transactions)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	mutData.Lock()
	assert.False(t, wasCalled)
	mutData.Unlock()

	var expectedSent [][]byte
	expectedSent = append(expectedSent, miniBlocksMarshalled[1])

	scm.HeaderReceived(metaBlock, []byte("meta hash"))
	time.Sleep(10 * time.Millisecond)
	mutData.Lock()
	assert.False(t, wasCalled)
	assert.False(t, isIncluded(expectedSent, broadcastBuffer))
	mutData.Unlock()
}

func TestShardChainMessenger_HeaderReceivedForNextRegisteredDelayedDataShouldBroadcastBoth(t *testing.T) {
	wasCalled := false
	mutData := &sync.Mutex{}
	broadcastBuffer := make([][]byte, 0)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutData.Lock()
			wasCalled = true
			broadcastBuffer = append(broadcastBuffer, buff)
			mutData.Unlock()
		},
	}

	args := createDefaultShardChainArgs()
	args.MaxDelayCacheSize = 2
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	headerHash, miniBlocksMarshalled, transactions := createDelayData("1")
	err := scm.SetLeaderDelayBroadcast(headerHash, miniBlocksMarshalled, transactions)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	mutData.Lock()
	assert.False(t, wasCalled)
	mutData.Unlock()

	headerHash2, miniBlocksMarshalled2, transactions2 := createDelayData("2")
	err = scm.SetLeaderDelayBroadcast(headerHash2, miniBlocksMarshalled2, transactions2)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	mutData.Lock()
	assert.False(t, wasCalled)
	mutData.Unlock()

	metaBlock := createMetaBlock()
	metaBlock.ShardInfo[0].HeaderHash = headerHash2

	var expectedSent [][]byte
	expectedSent = append(expectedSent, miniBlocksMarshalled[1])
	expectedSent = append(expectedSent, miniBlocksMarshalled2[1])

	scm.HeaderReceived(metaBlock, []byte("meta hash"))
	time.Sleep(core.ExtraDelayForBroadcastBlockInfo + 10*time.Millisecond)
	mutData.Lock()
	assert.True(t, wasCalled)
	assert.True(t, isIncluded(expectedSent, broadcastBuffer))
	mutData.Unlock()
}

func TestShardChainMessenger_SetValidatorDelayBroadcastNilHeaderHashShouldErr(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)
	vArgs.headerHash = nil

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)

	require.Equal(t, spos.ErrNilHeaderHash, err)
}

func TestShardChainMessenger_SetValidatorDelayBroadcastInvalidMiniBlockDataShouldErr(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)
	vArgs.miniBlocks = nil

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)

	require.Equal(t, spos.ErrInvalidDataToBroadcast, err)
}

func TestShardChainMessenger_SetValidatorDelayBroadcastNoMiniBlocksShouldReturn(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)
	vArgs.miniBlocks = nil
	vArgs.miniBlockHashes = nil

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)

	require.Equal(t, nil, err)
}

func TestShardChainMessenger_SetValidatorDelayBroadcastAccumulatedDataBounded(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)

	vbd := scm.GetValidatorBroadcastData()
	expectedLen := 0
	require.Equal(t, expectedLen, len(vbd))

	for i := 1; i < 100; i++ {
		err := scm.SetValidatorDelayBroadcast(
			vArgs.headerHash,
			vArgs.prevRandSeed,
			vArgs.round,
			vArgs.miniBlocks,
			vArgs.miniBlockHashes,
			vArgs.transactions,
			vArgs.order,
		)
		require.Nil(t, err)
		vbd = scm.GetValidatorBroadcastData()
		expectedLen = i
		if i > int(args.MaxValidatorDelayCacheSize) {
			expectedLen = int(args.MaxValidatorDelayCacheSize)
		}
		require.Equal(t, expectedLen, len(vbd))
	}
}

func TestSharadChainMessenger_ScheduleValidatorBroadcastDifferentHeaderRoundShouldDoNothing(t *testing.T) {
	wasCalled := false
	mutData := &sync.Mutex{}
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutData.Lock()
			wasCalled = true
			mutData.Unlock()
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	require.Nil(t, err)
	vbd := scm.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round + 1,
		PrevRandSeed: vArgs.prevRandSeed,
	}

	scm.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was no broadcast and validator delay data still present
	mutData.Lock()
	called := wasCalled
	mutData.Unlock()
	require.False(t, called)

	vbd = scm.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
}

func TestShardChainMessenger_ScheduleValidatorBroadcastDifferentPrevRandShouldDoNothing(t *testing.T) {
	wasCalled := false
	mutData := &sync.Mutex{}
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutData.Lock()
			wasCalled = true
			mutData.Unlock()
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	require.Nil(t, err)
	vbd := scm.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	prevRandSeed := make([]byte, len(vArgs.prevRandSeed))
	copy(prevRandSeed, vArgs.prevRandSeed)
	prevRandSeed[0] = ^prevRandSeed[0]

	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round,
		PrevRandSeed: prevRandSeed,
	}

	scm.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was no broadcast and validator delay data still present
	mutData.Lock()
	called := wasCalled
	mutData.Unlock()
	require.False(t, called)

	vbd = scm.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
}

func TestShardChainMessenger_ScheduleValidatorBroadcastSameRoundAndPrevRandShouldScheduleAlarm(t *testing.T) {
	wasCalled := false
	mutData := &sync.Mutex{}
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutData.Lock()
			wasCalled = true
			mutData.Unlock()
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	require.Nil(t, err)
	vbd := scm.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	hdfv := &broadcast.HeaderDataForValidator{
		Round:        vArgs.round,
		PrevRandSeed: vArgs.prevRandSeed,
	}

	scm.ScheduleValidatorBroadcast([]*broadcast.HeaderDataForValidator{hdfv})
	timeToWait := time.Duration(vArgs.order) * broadcast.ValidatorDelayPerOrder()
	time.Sleep(timeToWait + 100*time.Millisecond)

	// check there was a broadcast and validator delay data empty
	mutData.Lock()
	called := wasCalled
	mutData.Unlock()
	require.True(t, called)

	vbd = scm.GetValidatorBroadcastData()
	require.Equal(t, 0, len(vbd))
}

func TestShardChainMessenger_AlarmExpiredShouldBroadcastTheDataForRegisteredDelayedData(t *testing.T) {
	wasCalled := false
	mutData := &sync.Mutex{}
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutData.Lock()
			wasCalled = true
			mutData.Unlock()
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	require.Nil(t, err)
	vbd := scm.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	scm.AlarmExpired(string(vArgs.headerHash))
	time.Sleep(time.Millisecond * 100)

	// check there was a broadcast and validator delay data empty
	mutData.Lock()
	called := wasCalled
	mutData.Unlock()
	require.True(t, called)

	vbd = scm.GetValidatorBroadcastData()
	require.Equal(t, 0, len(vbd))
}

func TestShardChainMessenger_AlarmExpiredShouldDoNothingForNotRegisteredData(t *testing.T) {
	wasCalled := false
	mutData := &sync.Mutex{}
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutData.Lock()
			wasCalled = true
			mutData.Unlock()
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	require.Nil(t, err)
	vbd := scm.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))

	differentHeaderHash := make([]byte, len(vArgs.headerHash))
	copy(differentHeaderHash, vArgs.headerHash)
	differentHeaderHash[0] = ^differentHeaderHash[0]
	scm.AlarmExpired(string(differentHeaderHash))
	time.Sleep(time.Millisecond * 100)

	// check there was no broadcast and validator delay data still present
	mutData.Lock()
	called := wasCalled
	mutData.Unlock()
	require.False(t, called)

	vbd = scm.GetValidatorBroadcastData()
	require.Equal(t, 1, len(vbd))
}

func TestShardChainMessenger_RegisterInterceptorCallback(t *testing.T) {
	args := createDefaultShardChainArgs()
	var cbs []func(toShard uint32, data []byte)
	mutCbs := &sync.Mutex{}

	registerHandler := func(handler func(toShard uint32, data []byte)) {
		mutCbs.Lock()
		cbs = append(cbs, handler)
		mutCbs.Unlock()
	}

	args.InterceptorsContainer = &mock.InterceptorsContainerStub{
		GetCalled: func(topic string) (process.Interceptor, error) {
			return &mock.InterceptorStub{
				RegisterHandlerCalled: registerHandler,
			}, nil
		},
	}
	scm, _ := broadcast.NewShardChainMessenger(args)
	vArgs := createValidatorDelayArgs(0)
	mutCbs.Lock()
	nbRegistered := len(cbs)
	mutCbs.Unlock()
	require.Equal(t, 1, nbRegistered)

	err := scm.SetValidatorDelayBroadcast(
		vArgs.headerHash,
		vArgs.prevRandSeed,
		vArgs.round,
		vArgs.miniBlocks,
		vArgs.miniBlockHashes,
		vArgs.transactions,
		vArgs.order,
	)
	require.Nil(t, err)

	cb := func(toShard uint32, data []byte) {}
	err = scm.RegisterInterceptorCallback(cb)
	require.Nil(t, err)

	mutCbs.Lock()
	nbRegistered = len(cbs)
	mutCbs.Unlock()
	require.Equal(t, 2, nbRegistered)
}
