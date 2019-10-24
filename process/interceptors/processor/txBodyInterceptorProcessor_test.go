package processor_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &mock.MarshalizerMock{}
var testHasher = mock.HasherMock{}

func createMockTxBodyArgument() *processor.ArgTxBodyInterceptorProcessor {
	return &processor.ArgTxBodyInterceptorProcessor{
		MiniblockCache:   &mock.CacherStub{},
		Marshalizer:      testMarshalizer,
		Hasher:           testHasher,
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
	}
}

func createInteceptedTxBlockBody(txBlockBody block.Body) *interceptedBlocks.InterceptedTxBlockBody {
	arg := &interceptedBlocks.ArgInterceptedTxBlockBody{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Hasher:           testHasher,
		Marshalizer:      testMarshalizer,
	}

	arg.TxBlockBodyBuff, _ = testMarshalizer.Marshal(txBlockBody)
	inTxBody, _ := interceptedBlocks.NewInterceptedTxBlockBody(arg)

	return inTxBody
}

func TestNewTxBodyInterceptorProcessor_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	tbip, err := processor.NewTxBodyInterceptorProcessor(nil)

	assert.Nil(t, tbip)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewTxBodyInterceptorProcessor_NilMiniblocksShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockTxBodyArgument()
	arg.MiniblockCache = nil
	tbip, err := processor.NewTxBodyInterceptorProcessor(arg)

	assert.Nil(t, tbip)
	assert.Equal(t, process.ErrNilMiniBlockPool, err)
}

func TestNewTxBodyInterceptorProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockTxBodyArgument()
	arg.Marshalizer = nil
	tbip, err := processor.NewTxBodyInterceptorProcessor(arg)

	assert.Nil(t, tbip)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewTxBodyInterceptorProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockTxBodyArgument()
	arg.Hasher = nil
	tbip, err := processor.NewTxBodyInterceptorProcessor(arg)

	assert.Nil(t, tbip)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewTxBodyInterceptorProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockTxBodyArgument()
	arg.ShardCoordinator = nil
	tbip, err := processor.NewTxBodyInterceptorProcessor(arg)

	assert.Nil(t, tbip)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTxBodyInterceptorProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	tbip, err := processor.NewTxBodyInterceptorProcessor(createMockTxBodyArgument())

	assert.False(t, check.IfNil(tbip))
	assert.Nil(t, err)
}

//------- Validate

func TestTxBodyInterceptorProcessor_ValidateShouldWork(t *testing.T) {
	t.Parallel()

	tbip, _ := processor.NewTxBodyInterceptorProcessor(createMockTxBodyArgument())

	assert.Nil(t, tbip.Validate(nil))
}

//------- Save

func TestTxBodyInterceptorProcessor_SaveWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	tbip, _ := processor.NewTxBodyInterceptorProcessor(createMockTxBodyArgument())

	err := tbip.Save(nil)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestTxBodyInterceptorProcessor_SaveEmptyBlockShouldNotAdd(t *testing.T) {
	t.Parallel()

	arg := createMockTxBodyArgument()
	cacher := arg.MiniblockCache.(*mock.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		assert.Fail(t, "hasOrAdd should have not been called")
		return
	}
	tbip, _ := processor.NewTxBodyInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedTxBlockBody(make([]*block.MiniBlock, 0))

	err := tbip.Save(inTxBlkBdy)

	assert.Nil(t, err)
}

func TestTxBodyInterceptorProcessor_SaveMiniblocksNotForCurrentShardShouldNotAdd(t *testing.T) {
	t.Parallel()

	txBlockBody := []*block.MiniBlock{
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: 1,
			SenderShardID:   2,
			Type:            0,
		},
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: 3,
			SenderShardID:   4,
			Type:            0,
		},
	}

	arg := createMockTxBodyArgument()
	cacher := arg.MiniblockCache.(*mock.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		assert.Fail(t, "hasOrAdd should have not been called")
		return
	}
	tbip, _ := processor.NewTxBodyInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedTxBlockBody(txBlockBody)

	err := tbip.Save(inTxBlkBdy)

	assert.Nil(t, err)
}

func TestTxBodyInterceptorProcessor_SaveMiniblocksWithSenderShouldAdd(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	txBlockBody := []*block.MiniBlock{
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: 1,
			SenderShardID:   2,
			Type:            0,
		},
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: 3,
			SenderShardID:   currentShard,
			Type:            0,
		},
	}

	arg := createMockTxBodyArgument()
	cacher := arg.MiniblockCache.(*mock.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		miniblock, ok := value.(*block.MiniBlock)
		if !ok {
			assert.Fail(t, "hasOrAdd called for an invalid type")
			return
		}
		if miniblock.SenderShardID != currentShard {
			assert.Fail(t, "hasOrAdd called for the wrong object")
		}

		return
	}
	tbip, _ := processor.NewTxBodyInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedTxBlockBody(txBlockBody)

	err := tbip.Save(inTxBlkBdy)

	assert.Nil(t, err)
}

func TestTxBodyInterceptorProcessor_SaveMiniblocksWithReceiverShouldAdd(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	txBlockBody := []*block.MiniBlock{
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: 1,
			SenderShardID:   2,
			Type:            0,
		},
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: currentShard,
			SenderShardID:   3,
			Type:            0,
		},
	}

	arg := createMockTxBodyArgument()
	cacher := arg.MiniblockCache.(*mock.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		miniblock, ok := value.(*block.MiniBlock)
		if !ok {
			assert.Fail(t, "hasOrAdd called for an invalid type")
			return
		}
		if miniblock.ReceiverShardID != currentShard {
			assert.Fail(t, "hasOrAdd called for the wrong object")
		}

		return
	}
	tbip, _ := processor.NewTxBodyInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedTxBlockBody(txBlockBody)

	err := tbip.Save(inTxBlkBdy)

	assert.Nil(t, err)
}

func TestTxBodyInterceptorProcessor_SaveMiniblocksMarshalizerFailShouldNotAdd(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	txBlockBody := []*block.MiniBlock{
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: currentShard,
			SenderShardID:   currentShard,
			Type:            0,
		},
		{
			TxHashes:        make([][]byte, 0),
			ReceiverShardID: currentShard,
			SenderShardID:   currentShard,
			Type:            0,
		},
	}

	errExpected := errors.New("expected error")
	arg := createMockTxBodyArgument()
	arg.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
			return nil, errExpected
		},
	}
	cacher := arg.MiniblockCache.(*mock.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		assert.Fail(t, "hasOrAdd should have not been called")
		return
	}
	tbip, _ := processor.NewTxBodyInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedTxBlockBody(txBlockBody)

	err := tbip.Save(inTxBlkBdy)

	assert.Equal(t, errExpected, err)
}

//------- IsInterfaceNil

func TestTxBodyInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var tbip *processor.TxBodyInterceptorProcessor

	assert.True(t, check.IfNil(tbip))
}
