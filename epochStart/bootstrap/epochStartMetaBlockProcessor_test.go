package bootstrap

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func TestNewEpochStartMetaBlockProcessor_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		nil,
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	assert.Equal(t, epochStart.ErrNilMessenger, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessor_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		nil,
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	assert.Equal(t, epochStart.ErrNilRequestHandler, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		nil,
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	assert.Equal(t, epochStart.ErrNilMarshalizer, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		nil,
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	assert.Equal(t, epochStart.ErrNilHasher, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessor_InvalidConsensusPercentageShouldErr(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		101,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	assert.Equal(t, epochStart.ErrInvalidConsensusThreshold, err)
	assert.True(t, check.IfNil(esmbp))
}

func TestNewEpochStartMetaBlockProcessorOkValsShouldWork(t *testing.T) {
	t.Parallel()

	esmbp, err := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			}},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
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
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				peers := getConnectedPeers(counter)
				counter++
				return peers
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	assert.NoError(t, err)
	assert.False(t, check.IfNil(esmbp))
}

func TestEpochStartMetaBlockProcessor_Validate(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			}},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	assert.Nil(t, esmbp.Validate(nil, ""))
}

func TestEpochStartMetaBlockProcessor_SaveNilInterceptedDataShouldNotReturnError(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			}},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	err := esmbp.Save(nil, "peer0", "")
	assert.NoError(t, err)
}

func TestEpochStartMetaBlockProcessor_SaveOkInterceptedDataShouldWork(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
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
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
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
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		99,
		3,
		5,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	expectedMetaBlock := &block.MetaBlock{
		Nonce:      10,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{Round: 1}}},
	}
	confirmationMetaBlock := &block.MetaBlock{
		Nonce: 11,
	}
	intData := mock.NewInterceptedMetaBlockMock(expectedMetaBlock, []byte("hash"))
	intData2 := mock.NewInterceptedMetaBlockMock(confirmationMetaBlock, []byte("hash2"))

	for i := 0; i < esmbp.minNumOfPeersToConsiderBlockValid; i++ {
		_ = esmbp.Save(intData, core.PeerID(fmt.Sprintf("peer_%d", i)), "")
	}

	for i := 0; i < esmbp.minNumOfPeersToConsiderBlockValid; i++ {
		_ = esmbp.Save(intData2, core.PeerID(fmt.Sprintf("peer_%d", i)), "")
	}

	// we need a slightly more time than 1 second in order to also properly test the select branches
	timeout := 2*time.Second + time.Millisecond*500
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	mb, err := esmbp.GetEpochStartMetaBlock(ctx)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, expectedMetaBlock, mb)
}

func TestEpochStartMetaBlockProcessor_GetEpochStartMetaBlockShouldWorkFromFirstTry(t *testing.T) {
	t.Parallel()

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		50,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
	)

	expectedMetaBlock := &block.MetaBlock{
		Nonce:      10,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{Round: 1}}},
	}
	confirmationMetaBlock := &block.MetaBlock{
		Nonce: 11,
	}
	intData := mock.NewInterceptedMetaBlockMock(expectedMetaBlock, []byte("hash"))
	intData2 := mock.NewInterceptedMetaBlockMock(confirmationMetaBlock, []byte("hash2"))

	for i := 0; i < 6; i++ {
		_ = esmbp.Save(intData, core.PeerID(fmt.Sprintf("peer_%d", i)), "")
	}

	for i := 0; i < 6; i++ {
		_ = esmbp.Save(intData2, core.PeerID(fmt.Sprintf("peer_%d", i)), "")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	mb, err := esmbp.GetEpochStartMetaBlock(ctx)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, expectedMetaBlock, mb)
}

func TestEpochStartMetaBlockProcessor_GetEpochStartMetaBlock_BeforeAndromeda(t *testing.T) {
	t.Parallel()

	tts := durationBetweenChecks - 10*time.Millisecond

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		64,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&dataRetriever.ProofsPoolMock{},
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

func TestEpochStartMetaBlockProcessor_GetEpochStartMetaBlock_AfterAndromeda(t *testing.T) {
	t.Parallel()

	tts := durationBetweenChecks - 10*time.Millisecond

	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1", "peer_2", "peer_3", "peer_4", "peer_5"}
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		64,
		3,
		3,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		},
		&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		},
	)
	expectedMetaBlock := &block.MetaBlock{
		Nonce:      10,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{Round: 1}}},
	}
	intData := mock.NewInterceptedMetaBlockMock(expectedMetaBlock, []byte("hash"))

	confirmationMetaBlock := &block.MetaBlock{
		Nonce: 11,
	}
	intData2 := mock.NewInterceptedMetaBlockMock(confirmationMetaBlock, []byte("hash2"))

	go func() {
		index := 0
		for {
			time.Sleep(tts)
			_ = esmbp.Save(intData, core.PeerID(fmt.Sprintf("peer_%d", index)), "")
			_ = esmbp.Save(intData, core.PeerID(fmt.Sprintf("peer_%d", index+1)), "")
			index += 2
		}
	}()

	go func() {
		index := 0
		for {
			time.Sleep(tts)
			_ = esmbp.Save(intData2, core.PeerID(fmt.Sprintf("peer_%d", index)), "")
			_ = esmbp.Save(intData2, core.PeerID(fmt.Sprintf("peer_%d", index+1)), "")
			index += 2
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	mb, err := esmbp.GetEpochStartMetaBlock(ctx)
	cancel()
	assert.NoError(t, err)
	assert.Equal(t, expectedMetaBlock, mb)
}
