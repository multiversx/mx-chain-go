package indexer

import (
	"sync"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	dataProcess "github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMetaBlock() *dataBlock.MetaBlock {
	shardData := dataBlock.ShardData{
		ShardID:               1,
		HeaderHash:            []byte{1},
		ShardMiniBlockHeaders: []dataBlock.MiniBlockHeader{},
		TxCount:               100,
	}
	return &dataBlock.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   100,
		ShardInfo: []dataBlock.ShardData{shardData},
	}
}

func NewDataIndexerArguments() ArgDataIndexer {
	return ArgDataIndexer{
		Marshalizer:        &mock.MarshalizerMock{},
		NodesCoordinator:   &mock.NodesCoordinatorMock{},
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		DataDispatcher:     &mock.DispatcherMock{},
		ElasticProcessor:   &testscommon.ElasticProcessorStub{},
		ShardCoordinator:   &mock.ShardCoordinatorMock{},
	}
}

func TestDataIndexer_NewIndexerWithNilNodesCoordinatorShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.NodesCoordinator = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilNodesCoordinator, err)
}

func TestDataIndexer_NewIndexerWithNilDataDispatcherShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, ErrNilDataDispatcher, err)
}

func TestDataIndexer_NewIndexerWithNilElasticProcessorShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.ElasticProcessor = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, ErrNilElasticProcessor, err)
}

func TestDataIndexer_NewIndexerWithNilMarshalizerShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.Marshalizer = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilMarshalizer, err)
}

func TestDataIndexer_NewIndexerWithNilEpochStartNotifierShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.EpochStartNotifier = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilEpochStartNotifier, err)
}

func TestDataIndexer_NewIndexerWithCorrectParamsShouldWork(t *testing.T) {
	arguments := NewDataIndexerArguments()

	ei, err := NewDataIndexer(arguments)

	require.Nil(t, err)
	require.False(t, check.IfNil(ei))
	require.False(t, ei.IsNilIndexer())
}

func TestDataIndexer_UpdateTPS(t *testing.T) {
	t.Parallel()

	called := false
	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, err := NewDataIndexer(arguments)
	require.Nil(t, err)
	_ = ei.Close()

	tpsBench := testscommon.TpsBenchmarkMock{}
	tpsBench.Update(newTestMetaBlock())

	ei.UpdateTPS(&tpsBench)
	require.True(t, called)
}

func TestDataIndexer_UpdateTPSNil(t *testing.T) {
	//TODO fix this test without logging subsystem

	_ = logger.SetLogLevel("core/indexer:TRACE")
	arguments := NewDataIndexerArguments()

	ei, err := NewDataIndexer(arguments)
	require.Nil(t, err)
	_ = ei.Close()

	ei.UpdateTPS(nil)
}

func TestDataIndexer_SaveBlock(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	args := &types.ArgsSaveBlockData{
		HeaderHash:             []byte("hash"),
		Body:                   &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{}},
		Header:                 nil,
		SignersIndexes:         nil,
		NotarizedHeadersHashes: nil,
		TransactionsPool:       nil,
	}
	ei.SaveBlock(args)
	require.True(t, called)
}

func TestDataIndexer_SaveRoundInfo(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}

	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	ei, _ := NewDataIndexer(arguments)
	_ = ei.Close()

	ei.SaveRoundsInfo([]*types.RoundInfo{})
	require.True(t, called)
}

func TestDataIndexer_SaveValidatorsPubKeys(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	valPubKey := make(map[uint32][][]byte)

	keys := [][]byte{[]byte("key")}
	valPubKey[0] = keys
	epoch := uint32(0)

	ei.SaveValidatorsPubKeys(valPubKey, epoch)
	require.True(t, called)
}

func TestDataIndexer_SaveValidatorsRating(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	ei.SaveValidatorsRating("ID", []*types.ValidatorRatingInfo{
		{Rating: 1}, {Rating: 2},
	})
	require.True(t, called)
}

func TestDataIndexer_RevertIndexedBlock(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	ei.RevertIndexedBlock(&dataBlock.Header{}, &dataBlock.Body{})
	require.True(t, called)
}

func TestDataIndexer_SetTxLogsProcessor(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.ElasticProcessor = &testscommon.ElasticProcessorStub{
		SetTxLogsProcessorCalled: func(txLogsProc dataProcess.TransactionLogProcessorDatabase) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	ei.SetTxLogsProcessor(disabled.NewNilTxLogsProcessor())
	require.True(t, called)
}

func TestDataIndexer_EpochChange(t *testing.T) {
	getEligibleValidatorsCalled := false

	_ = logger.SetLogLevel("core/indexer:TRACE")
	arguments := NewDataIndexerArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	arguments.ShardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}
	epochChangeNotifier := &mock.EpochStartNotifierStub{}
	arguments.EpochStartNotifier = epochChangeNotifier

	var wg sync.WaitGroup
	wg.Add(1)

	testEpoch := uint32(1)
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (m map[uint32][][]byte, err error) {
			defer wg.Done()
			if testEpoch == epoch {
				getEligibleValidatorsCalled = true
			}

			return nil, nil
		},
	}

	ei, _ := NewDataIndexer(arguments)
	assert.NotNil(t, ei)

	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 1, Epoch: testEpoch})
	wg.Wait()

	assert.True(t, getEligibleValidatorsCalled)
}

func TestDataIndexer_EpochChangeValidators(t *testing.T) {
	_ = logger.SetLogLevel("core/indexer:TRACE")
	arguments := NewDataIndexerArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	arguments.ShardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}
	epochChangeNotifier := &mock.EpochStartNotifierStub{}
	arguments.EpochStartNotifier = epochChangeNotifier

	var wg sync.WaitGroup

	val1PubKey := []byte("val1")
	val2PubKey := []byte("val2")
	val1MetaPubKey := []byte("val3")
	val2MetaPubKey := []byte("val4")

	validatorsEpoch1 := map[uint32][][]byte{
		0:                     {val1PubKey, val2PubKey},
		core.MetachainShardId: {val1MetaPubKey, val2MetaPubKey},
	}
	validatorsEpoch2 := map[uint32][][]byte{
		0:                     {val2PubKey, val1PubKey},
		core.MetachainShardId: {val2MetaPubKey, val1MetaPubKey},
	}
	var firstEpochCalled, secondEpochCalled bool
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (m map[uint32][][]byte, err error) {
			defer wg.Done()

			switch epoch {
			case 1:
				firstEpochCalled = true
				return validatorsEpoch1, nil
			case 2:
				secondEpochCalled = true
				return validatorsEpoch2, nil
			default:
				return nil, nil
			}
		},
	}

	ei, _ := NewDataIndexer(arguments)
	assert.NotNil(t, ei)

	wg.Add(1)
	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 1, Epoch: 1})
	wg.Wait()
	assert.True(t, firstEpochCalled)

	wg.Add(1)
	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 10, Epoch: 2})
	wg.Wait()
	assert.True(t, secondEpochCalled)
}
