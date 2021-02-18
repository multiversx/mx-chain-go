package bootstrap

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewEpochStartMetaBlockProcessor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		nil,
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	assert.Equal(t, epochStart.ErrNilMessenger, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessor_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	assert.Equal(t, epochStart.ErrNilRequestHandler, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		nil,
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	assert.Equal(t, epochStart.ErrNilMarshalizer, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		nil,
		50,
		3,
		3,
	)

	assert.Equal(t, epochStart.ErrNilHasher, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessor_InvalidConsensusPercentageShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		101,
		3,
		3,
	)

	assert.Equal(t, epochStart.ErrInvalidConsensusThreshold, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessorOkValsShouldWork(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	assert.NoError(t, err)
	assert.False(t, check.IfNil(esmbp))
}

func getConnectedPeers(counter int) []core.PeerID {
	switch counter {
	case 0:
		return []core.PeerID{"peer0", "peer1"}
	case 1:
		return []core.PeerID{"peer0", "peer1", "peer2", "peer3"}
	case 2:
		return []core.PeerID{"peer0", "peer1", "peer2", "peer3", "peer4", "peer5"}
	}
	return nil
}

func TestNewEpochStartMetaBlockProcessorOkValsShouldWorkAfterMoreTriesWaitingForConnectedPeers(t *testing.T) {
	t.Parallel()

	counter := 0
	esmbp, err := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				peers := getConnectedPeers(counter)
				counter++
				return peers
			},
		},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	assert.NoError(t, err)
	assert.False(t, check.IfNil(esmbp))
}

func TestEpochStartMetaBlockProcessor_Validate(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	assert.Nil(t, esmbp.Validate(nil, ""))
}

func TestEpochStartMetaBlockProcessor_SaveNilInterceptedDataShouldNotReturnError(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	err := esmbp.Save(nil, "peer0", "")
	assert.NoError(t, err)
}

func TestEpochStartMetaBlockProcessor_SaveOkInterceptedDataShouldWork(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	assert.Zero(t, len(esmbp.GetMapMetaBlock()))
	mb := &block.MetaBlock{
		Nonce:      10,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{Round: 1}}},
	}
	intData := mock.NewInterceptedMetaBlockMock(mb, []byte("hash"))
	err := esmbp.Save(intData, "peer0", "")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(esmbp.GetMapMetaBlock()))
}

func TestEpochStartMetaBlockProcessor_GetEpochStartMetaBlockShouldTimeOut(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	mb, err := esmbp.GetEpochStartMetaBlock(ctx)
	cancel()
	assert.Nil(t, mb)
	assert.Equal(t, epochStart.ErrTimeoutWaitingForMetaBlock, err)
}

func TestEpochStartMetaBlockProcessor_GetEpochStartMetaBlockShouldReturnMostReceivedAfterTimeOut(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		99,
		3,
		3,
	)

	expectedMetaBlock := &block.MetaBlock{
		Nonce:      10,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{Round: 1}}},
	}
	intData := mock.NewInterceptedMetaBlockMock(expectedMetaBlock, []byte("hash"))

	for i := 0; i < esmbp.minNumOfPeersToConsiderBlockValid; i++ {
		_ = esmbp.Save(intData, core.PeerID(fmt.Sprintf("peer_%d", i)), "")
	}

	// we need a slightly more time than 1 second in order to also properly test the select branches
	timeout := time.Second + time.Millisecond*500
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	mb, err := esmbp.GetEpochStartMetaBlock(ctx)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, expectedMetaBlock, mb)
}

func TestEpochStartMetaBlockProcessor_GetEpochStartMetaBlockShouldWorkFromFirstTry(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		50,
		3,
		3,
	)

	expectedMetaBlock := &block.MetaBlock{
		Nonce:      10,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{Round: 1}}},
	}
	intData := mock.NewInterceptedMetaBlockMock(expectedMetaBlock, []byte("hash"))

	for i := 0; i < 6; i++ {
		_ = esmbp.Save(intData, core.PeerID(fmt.Sprintf("peer_%d", i)), "")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	mb, err := esmbp.GetEpochStartMetaBlock(ctx)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, expectedMetaBlock, mb)
}

func TestEpochStartMetaBlockProcessor_GetEpochStartMetaBlockShouldWorkAfterMultipleTries(t *testing.T) {
	t.Parallel()

	testEpochStartMbIsReceivedWithSleepBetweenReceivedMessages(t, durationBetweenChecks-10*time.Millisecond)
}

func TestEpochStartMetaBlockProcessor_GetEpochStartMetaBlockShouldWorkAfterMultipleRequests(t *testing.T) {
	t.Parallel()

	testEpochStartMbIsReceivedWithSleepBetweenReceivedMessages(t, durationBetweenChecks-10*time.Millisecond)
}

func testEpochStartMbIsReceivedWithSleepBetweenReceivedMessages(t *testing.T, tts time.Duration) {
	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&mock.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&mock.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		64,
		3,
		3,
	)
	expectedMetaBlock := &block.MetaBlock{
		Nonce:      10,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{Round: 1}}},
	}
	intData := mock.NewInterceptedMetaBlockMock(expectedMetaBlock, []byte("hash"))
	go func() {
		index := 0
		for {
			time.Sleep(tts)
			_ = esmbp.Save(intData, core.PeerID(fmt.Sprintf("peer_%d", index)), "")
			_ = esmbp.Save(intData, core.PeerID(fmt.Sprintf("peer_%d", index+1)), "")
			index += 2
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	mb, err := esmbp.GetEpochStartMetaBlock(ctx)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, expectedMetaBlock, mb)
}
