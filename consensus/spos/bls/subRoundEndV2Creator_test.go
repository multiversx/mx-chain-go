package bls_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

func TestNewSubRoundEndV2Creator(t *testing.T) {
	t.Parallel()

	creator := bls.NewSubRoundEndV2Creator()
	require.False(t, creator.IsInterfaceNil())
	require.Implements(t, new(bls.SubRoundEndV2Creator), creator)
	require.Equal(t, "*bls.subRoundEndV2Creator", fmt.Sprintf("%T", creator))
}

func TestSubRoundEndV2Creator_CreateAndAddSubRoundEnd(t *testing.T) {
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
			require.Equal(t, "*bls.subroundEndRoundV2", fmt.Sprintf("%T", handler))
		},
	})

	sr := initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	creator := bls.NewSubRoundEndV2Creator()
	err := creator.CreateAndAddSubRoundEnd(sr, workerHandler, consensusCore)
	require.Nil(t, err)
	require.Equal(t, 2, addReceivedMessageCallCt)
	require.Equal(t, 1, addReceivedHeaderHandlerCallCt)
	require.Equal(t, 1, addSubRoundCalledCt)
}
