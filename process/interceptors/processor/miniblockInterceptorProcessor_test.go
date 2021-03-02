package processor_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &mock.MarshalizerMock{}
var testHasher = mock.HasherMock{}

func createMockMiniblockArgument() *processor.ArgMiniblockInterceptorProcessor {
	return &processor.ArgMiniblockInterceptorProcessor{
		MiniblockCache:   testscommon.NewCacherStub(),
		Marshalizer:      testMarshalizer,
		Hasher:           testHasher,
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		WhiteListHandler: &mock.WhiteListHandlerStub{},
	}
}

func createInteceptedMiniblock(miniblock *block.MiniBlock) *interceptedBlocks.InterceptedMiniblock {
	arg := &interceptedBlocks.ArgInterceptedMiniblock{
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Hasher:           testHasher,
		Marshalizer:      testMarshalizer,
	}

	arg.MiniblockBuff, _ = testMarshalizer.Marshal(miniblock)
	inMb, _ := interceptedBlocks.NewInterceptedMiniblock(arg)

	return inMb
}

func TestNewMiniblockInterceptorProcessor_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	mip, err := processor.NewMiniblockInterceptorProcessor(nil)

	assert.Nil(t, mip)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewMiniblockInterceptorProcessor_NilMiniblocksShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockArgument()
	arg.MiniblockCache = nil
	mip, err := processor.NewMiniblockInterceptorProcessor(arg)

	assert.Nil(t, mip)
	assert.Equal(t, process.ErrNilMiniBlockPool, err)
}

func TestNewMiniblockInterceptorProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockArgument()
	arg.Marshalizer = nil
	mip, err := processor.NewMiniblockInterceptorProcessor(arg)

	assert.Nil(t, mip)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewMiniblockInterceptorProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockArgument()
	arg.Hasher = nil
	mip, err := processor.NewMiniblockInterceptorProcessor(arg)

	assert.Nil(t, mip)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewMiniblockInterceptorProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockArgument()
	arg.ShardCoordinator = nil
	mip, err := processor.NewMiniblockInterceptorProcessor(arg)

	assert.Nil(t, mip)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewMiniblockInterceptorProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	mip, err := processor.NewMiniblockInterceptorProcessor(createMockMiniblockArgument())

	assert.False(t, check.IfNil(mip))
	assert.Nil(t, err)
}

//------- Validate

func TestMiniblockInterceptorProcessor_ValidateShouldWork(t *testing.T) {
	t.Parallel()

	mip, _ := processor.NewMiniblockInterceptorProcessor(createMockMiniblockArgument())

	assert.Nil(t, mip.Validate(nil, ""))
}

//------- Save

func TestMiniblockInterceptorProcessor_SaveWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	mip, _ := processor.NewMiniblockInterceptorProcessor(createMockMiniblockArgument())

	err := mip.Save(nil, "", "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestMiniblockInterceptorProcessor_NilMiniblockShouldNotAdd(t *testing.T) {
	t.Parallel()

	arg := createMockMiniblockArgument()
	cacher := arg.MiniblockCache.(*testscommon.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
		assert.Fail(t, "hasOrAdd should have not been called")
		return
	}
	mip, _ := processor.NewMiniblockInterceptorProcessor(arg)

	err := mip.Save(nil, "", "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestMiniblockInterceptorProcessor_SaveMiniblockNotForCurrentShardShouldNotAdd(t *testing.T) {
	t.Parallel()

	miniblock := &block.MiniBlock{
		TxHashes:        make([][]byte, 0),
		ReceiverShardID: 1,
		SenderShardID:   2,
		Type:            0,
	}

	arg := createMockMiniblockArgument()
	cacher := arg.MiniblockCache.(*testscommon.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
		assert.Fail(t, "hasOrAdd should have not been called")
		return
	}
	_, _ = processor.NewMiniblockInterceptorProcessor(arg)
	inMb := createInteceptedMiniblock(miniblock)
	assert.False(t, inMb.IsForCurrentShard())
}

func TestMiniblockInterceptorProcessor_SaveMiniblockWithSenderInSameShardShouldAdd(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	miniblock := &block.MiniBlock{
		TxHashes:        make([][]byte, 0),
		ReceiverShardID: 3,
		SenderShardID:   currentShard,
		Type:            0,
	}

	arg := createMockMiniblockArgument()
	cacher := arg.MiniblockCache.(*testscommon.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
		_, ok := value.(*block.MiniBlock)
		if !ok {
			assert.Fail(t, "hasOrAdd called for an invalid type")
			return false, false
		}

		return false, true
	}
	mip, _ := processor.NewMiniblockInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedMiniblock(miniblock)

	err := mip.Save(inTxBlkBdy, "", "")

	assert.Nil(t, err)
}

func TestMiniblockInterceptorProcessor_SaveMiniblocksWithReceiverInSameShardShouldAdd(t *testing.T) {
	t.Parallel()

	currentShard := uint32(0)
	miniblock := &block.MiniBlock{
		TxHashes:        make([][]byte, 0),
		ReceiverShardID: currentShard,
		SenderShardID:   3,
		Type:            0,
	}

	arg := createMockMiniblockArgument()
	cacher := arg.MiniblockCache.(*testscommon.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
		_, ok := value.(*block.MiniBlock)
		if !ok {
			assert.Fail(t, "hasOrAdd called for an invalid type")
			return false, false
		}

		return false, true
	}
	mip, _ := processor.NewMiniblockInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedMiniblock(miniblock)
	chanCalled := make(chan struct{}, 1)
	mip.RegisterHandler(func(topic string, hash []byte, data interface{}) {
		chanCalled <- struct{}{}
	})

	err := mip.Save(inTxBlkBdy, "", "")
	assert.Nil(t, err)

	timeout := time.Second * 2
	select {
	case <-chanCalled:
	case <-time.After(timeout):
		assert.Fail(t, "save did not notify handler in a timely fashion")
	}
}

func TestMiniblockInterceptorProcessor_SaveMiniblockCrossShardForMeNotWhiteListedShouldNotAdd(t *testing.T) {
	t.Parallel()

	miniblock := &block.MiniBlock{
		TxHashes:        make([][]byte, 0),
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            0,
	}

	arg := createMockMiniblockArgument()
	whiteListHandler := arg.WhiteListHandler.(*mock.WhiteListHandlerStub)
	whiteListHandler.IsWhiteListedCalled = func(interceptedData process.InterceptedData) bool {
		return false
	}

	cacher := arg.MiniblockCache.(*testscommon.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
		assert.Fail(t, "hasOrAdd should have not been called")
		return
	}
	tbip, _ := processor.NewMiniblockInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedMiniblock(miniblock)

	err := tbip.Save(inTxBlkBdy, "", "")
	assert.Nil(t, err)
}

func TestMiniblockInterceptorProcessor_SaveMiniblockCrossShardForMeWhiteListedShouldBeAddInPool(t *testing.T) {
	t.Parallel()

	miniblock := &block.MiniBlock{
		TxHashes:        make([][]byte, 0),
		ReceiverShardID: 0,
		SenderShardID:   1,
		Type:            0,
	}

	arg := createMockMiniblockArgument()
	whiteListHandler := arg.WhiteListHandler.(*mock.WhiteListHandlerStub)
	whiteListHandler.IsWhiteListedCalled = func(interceptedData process.InterceptedData) bool {
		return true
	}

	addedInPool := false
	cacher := arg.MiniblockCache.(*testscommon.CacherStub)
	cacher.HasOrAddCalled = func(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
		addedInPool = true
		return false, true
	}
	tbip, _ := processor.NewMiniblockInterceptorProcessor(arg)
	inTxBlkBdy := createInteceptedMiniblock(miniblock)

	err := tbip.Save(inTxBlkBdy, "", "")
	assert.Nil(t, err)
	assert.True(t, addedInPool)
}
