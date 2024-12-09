package bls_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignSubRoundEndV2Creator(t *testing.T) {
	t.Parallel()

	t.Run("nil outgoing operations pool, should return error", func(t *testing.T) {
		creator, err := bls.NewSovereignSubRoundEndCreator(nil, &sovereign.BridgeOperationsHandlerMock{})
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilOutGoingOperationsPool, err)
	})
	t.Run("nil bridge op handler, should return error", func(t *testing.T) {
		creator, err := bls.NewSovereignSubRoundEndCreator(&sovereign.OutGoingOperationsPoolMock{}, nil)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilBridgeOpHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		creator, err := bls.NewSovereignSubRoundEndCreator(&sovereign.OutGoingOperationsPoolMock{}, &sovereign.BridgeOperationsHandlerMock{})
		require.Nil(t, err)
		require.NotNil(t, creator)
		require.False(t, creator.IsInterfaceNil())
		require.Implements(t, new(bls.SubRoundEndV2Creator), creator)
		require.Equal(t, "*bls.sovereignSubRoundEndCreator", fmt.Sprintf("%T", creator))
	})

}

func TestSovereignSubRoundEndV2Creator_CreateAndAddSubRoundEnd(t *testing.T) {
	t.Parallel()

	addReceivedMessageCallCt := 0
	addReceivedHeaderHandlerCallCt := 0
	workerHandler := &mock.SposWorkerMock{
		AddReceivedMessageCallCalled: func(messageType consensus.MessageType, receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool) {
			addReceivedMessageCallCt++
			require.True(t, messageType == bls.MtBlockHeaderFinalInfo || messageType == bls.MtInvalidSigners)
		},
		AddReceivedHeaderHandlerCalled: func(handler func(data.HeaderHandler)) {
			addReceivedHeaderHandlerCallCt++
		},
	}

	addSubRoundCalledCt := 0
	consensusCore := &mock.ConsensusCoreMock{}
	consensusCore.SetChronology(&mock.ChronologyHandlerMock{
		AddSubroundCalled: func(handler consensus.SubroundHandler) {
			addSubRoundCalledCt++
			require.Equal(t, "*bls.sovereignSubRoundEnd", fmt.Sprintf("%T", handler))
		},
	})

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	creator, _ := bls.NewSovereignSubRoundEndCreator(&sovereign.OutGoingOperationsPoolMock{}, &sovereign.BridgeOperationsHandlerMock{})
	err := creator.CreateAndAddSubRoundEnd(sr, workerHandler, consensusCore)
	require.Nil(t, err)
	require.Equal(t, 2, addReceivedMessageCallCt)
	require.Equal(t, 1, addReceivedHeaderHandlerCallCt)
	require.Equal(t, 1, addSubRoundCalledCt)
}
