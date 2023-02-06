package bootstrap

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func TestNewStorageEpochStartMetaBlockProcessor_InvalidArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	sesmbp, err := NewStorageEpochStartMetaBlockProcessor(
		nil,
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)
	assert.True(t, check.IfNil(sesmbp))
	assert.Equal(t, epochStart.ErrNilMessenger, err)

	sesmbp, err = NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		nil,
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)
	assert.True(t, check.IfNil(sesmbp))
	assert.Equal(t, epochStart.ErrNilRequestHandler, err)

	sesmbp, err = NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		nil,
		&hashingMocks.HasherMock{},
	)
	assert.True(t, check.IfNil(sesmbp))
	assert.Equal(t, epochStart.ErrNilMarshalizer, err)

	sesmbp, err = NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		nil,
	)
	assert.True(t, check.IfNil(sesmbp))
	assert.Equal(t, epochStart.ErrNilHasher, err)
}

func TestNewStorageEpochStartMetaBlockProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	sesmbp, err := NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)
	assert.False(t, check.IfNil(sesmbp))
	assert.Nil(t, err)
}

func TestStorageEpochStartMetaBlockProcessor_Validate(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)
	assert.Nil(t, sesmbp.Validate(nil, ""))
}

func TestStorageEpochStartMetaBlockProcessor_RegisterHandlerShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not paniced: %v", r))
		}
	}()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)
	sesmbp.RegisterHandler(nil)
}

func TestStorageEpochStartMetaBlockProcessor_SaveNilData(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)

	err := sesmbp.Save(nil, "", "")
	assert.Nil(t, err)
	mb, _ := sesmbp.getMetablock()
	assert.True(t, check.IfNil(mb))
}

func TestStorageEpochStartMetaBlockProcessor_SaveNotAHeader(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)

	err := sesmbp.Save(&testscommon.InterceptedDataStub{}, "", "")
	assert.Nil(t, err)
	mb, _ := sesmbp.getMetablock()
	assert.True(t, check.IfNil(mb))
}

func TestStorageEpochStartMetaBlockProcessor_SaveNotAnEpochStartBlock(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)

	metaBlock := &dataBlock.MetaBlock{
		Nonce: 1,
		Round: 1,
	}
	interceptedMetaBlock := mock.NewInterceptedMetaBlockMock(metaBlock, []byte("hash"))

	err := sesmbp.Save(interceptedMetaBlock, "", "")
	assert.Nil(t, err)
	mb, _ := sesmbp.getMetablock()
	assert.True(t, check.IfNil(mb))
}

func TestStorageEpochStartMetaBlockProcessor_SaveShouldWork(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)

	metaBlock := &dataBlock.MetaBlock{
		Nonce: 1,
		Round: 1,
		EpochStart: dataBlock.EpochStart{
			LastFinalizedHeaders: make([]dataBlock.EpochStartShardData, 1),
		},
	}
	interceptedMetaBlock := mock.NewInterceptedMetaBlockMock(metaBlock, []byte("hash"))

	err := sesmbp.Save(interceptedMetaBlock, "", "")
	assert.Nil(t, err)
	mb, _ := sesmbp.getMetablock()
	assert.True(t, mb == metaBlock) //pointer testing
}

func TestStorageEpochStartMetaBlockProcessor_GetEpochStartMetaBlockShouldRequestRepeatedly(t *testing.T) {
	t.Parallel()

	numUnsuccessfulRequests := 2
	numRequests := 0
	metaBlock := &dataBlock.MetaBlock{
		Nonce: 1,
		Round: 1,
		EpochStart: dataBlock.EpochStart{
			LastFinalizedHeaders: make([]dataBlock.EpochStartShardData, 1),
		},
	}
	interceptedMetaBlock := mock.NewInterceptedMetaBlockMock(metaBlock, []byte("hash"))

	var sesmbp EpochStartMetaBlockInterceptorProcessor
	sesmbp, _ = NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{
			RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
				numRequests++
				if numRequests > numUnsuccessfulRequests {
					_ = sesmbp.Save(interceptedMetaBlock, "", "")
				}
			},
		},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)

	recoveredMetaBlock, err := sesmbp.GetEpochStartMetaBlock(context.Background())
	assert.Nil(t, err)
	assert.True(t, metaBlock == recoveredMetaBlock)
}

func TestStorageEpochStartMetaBlockProcessor_GetEpochStartMetaBlockContextDoneShouldReturnImmediately(t *testing.T) {
	t.Parallel()

	var sesmbp EpochStartMetaBlockInterceptorProcessor
	sesmbp, _ = NewStorageEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
	)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second * 2)
		cancelFunc()
	}()

	recoveredMetaBlock, err := sesmbp.GetEpochStartMetaBlock(ctx)
	assert.Equal(t, process.ErrNilMetaBlockHeader, err)
	assert.True(t, check.IfNil(recoveredMetaBlock))
}
