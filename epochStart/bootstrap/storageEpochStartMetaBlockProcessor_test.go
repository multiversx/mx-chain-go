package bootstrap

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewStorageEpochStartMetaBlockProcessor_InvalidArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	sesmbp, err := NewStorageEpochStartMetaBlockProcessor(
		nil,
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
	)
	assert.True(t, check.IfNil(sesmbp))
	assert.Equal(t, epochStart.ErrNilMessenger, err)

	sesmbp, err = NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
	)
	assert.True(t, check.IfNil(sesmbp))
	assert.Equal(t, epochStart.ErrNilRequestHandler, err)

	sesmbp, err = NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		nil,
		&mock.HasherMock{},
	)
	assert.True(t, check.IfNil(sesmbp))
	assert.Equal(t, epochStart.ErrNilMarshalizer, err)

	sesmbp, err = NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		nil,
	)
	assert.True(t, check.IfNil(sesmbp))
	assert.Equal(t, epochStart.ErrNilHasher, err)
}

func TestNewStorageEpochStartMetaBlockProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	sesmbp, err := NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
	)
	assert.False(t, check.IfNil(sesmbp))
	assert.Nil(t, err)
}

func TestStorageEpochStartMetaBlockProcessor_Validate(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
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
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
	)
	sesmbp.RegisterHandler(nil)
}

func TestStorageEpochStartMetaBlockProcessor_SaveNilData(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
	)

	err := sesmbp.Save(nil, "", "")
	assert.Nil(t, err)
	mb, _ := sesmbp.getMetablock()
	assert.True(t, check.IfNil(mb))
}

func TestStorageEpochStartMetaBlockProcessor_SaveNotAHeader(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
	)

	err := sesmbp.Save(&mock.InterceptedDataStub{}, "", "")
	assert.Nil(t, err)
	mb, _ := sesmbp.getMetablock()
	assert.True(t, check.IfNil(mb))
}

func TestStorageEpochStartMetaBlockProcessor_SaveNotAnEpochStartBlock(t *testing.T) {
	t.Parallel()

	sesmbp, _ := NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
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
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
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
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{
			RequestStartOfEpochMetaBlockCalled: func(epoch uint32) {
				numRequests++
				if numRequests > numUnsuccessfulRequests {
					_ = sesmbp.Save(interceptedMetaBlock, "", "")
				}
			},
		},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
	)

	recoveredMetaBlock, err := sesmbp.GetEpochStartMetaBlock(context.Background())
	assert.Nil(t, err)
	assert.True(t, metaBlock == recoveredMetaBlock)
}

func TestStorageEpochStartMetaBlockProcessor_GetEpochStartMetaBlockContextDoneShouldReturnImmediately(t *testing.T) {
	t.Parallel()

	var sesmbp EpochStartMetaBlockInterceptorProcessor
	sesmbp, _ = NewStorageEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
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
