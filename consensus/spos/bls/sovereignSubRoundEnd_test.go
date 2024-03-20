package bls_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	sovCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/errors"
	sovBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

type sovEndRoundHandler interface {
	consensus.SubroundHandler
	DoSovereignEndRoundJob(ctx context.Context) bool
}

func createSovSubRoundEndWithSelfLeader(
	pool sovBlock.OutGoingOperationsPool,
	bridgeHandler bls.BridgeOperationsHandler,
	header data.HeaderHandler,
) sovEndRoundHandler {
	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	srV2, _ := bls.NewSubroundEndRoundV2(&sr)
	sovEndRound, _ := bls.NewSovereignSubRoundEndRound(srV2, pool, bridgeHandler)

	sovEndRound.SetSelfPubKey("A")
	sovEndRound.SetThreshold(bls.SrEndRound, 1)
	_ = sovEndRound.SetJobDone(sovEndRound.ConsensusGroup()[0], bls.SrSignature, true)
	sovEndRound.Header = header

	return sovEndRound
}

func createSovSubRoundEndWithParticipant(
	pool sovBlock.OutGoingOperationsPool,
	bridgeHandler bls.BridgeOperationsHandler,
	header data.HeaderHandler,
) sovEndRoundHandler {
	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	srV2, _ := bls.NewSubroundEndRoundV2(&sr)
	sovEndRound, _ := bls.NewSovereignSubRoundEndRound(srV2, pool, bridgeHandler)

	sovEndRound.SetSelfPubKey("*")
	sovEndRound.SetThreshold(bls.SrEndRound, 1)
	// set previous as finished
	sr.SetStatus(2, spos.SsFinished)
	// set current as not finished
	sr.SetStatus(3, spos.SsNotFinished)
	sovEndRound.Header = header
	sr.AddReceivedHeader(header)

	return sovEndRound
}

func TestNewSovereignSubRoundEndRound(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	srV2, _ := bls.NewSubroundEndRoundV2(&sr)

	t.Run("nil subround end, should return error", func(t *testing.T) {
		sovEndRound, err := bls.NewSovereignSubRoundEndRound(
			nil,
			&sovereign.OutGoingOperationsPoolMock{},
			&sovereign.BridgeOperationsHandlerMock{},
		)
		require.Equal(t, spos.ErrNilSubround, err)
		require.Nil(t, sovEndRound)
	})
	t.Run("nil outgoing op pool, should return error", func(t *testing.T) {
		sovEndRound, err := bls.NewSovereignSubRoundEndRound(
			srV2,
			nil,
			&sovereign.BridgeOperationsHandlerMock{},
		)
		require.Equal(t, errors.ErrNilOutGoingOperationsPool, err)
		require.Nil(t, sovEndRound)
	})
	t.Run("nil bridge op handler, should return error", func(t *testing.T) {
		sovEndRound, err := bls.NewSovereignSubRoundEndRound(
			srV2,
			&sovereign.OutGoingOperationsPoolMock{},
			nil,
		)
		require.Equal(t, errors.ErrNilBridgeOpHandler, err)
		require.Nil(t, sovEndRound)
	})
	t.Run("should work", func(t *testing.T) {
		sovEndRound, err := bls.NewSovereignSubRoundEndRound(
			srV2,
			&sovereign.OutGoingOperationsPoolMock{},
			&sovereign.BridgeOperationsHandlerMock{},
		)
		require.Nil(t, err)
		require.False(t, sovEndRound.IsInterfaceNil())
	})
}

func TestSovereignSubRoundEnd_DoEndJobByLeader(t *testing.T) {
	t.Parallel()

	t.Run("invalid header, should not finish with success", func(t *testing.T) {
		t.Parallel()

		sovEndRound := createSovSubRoundEndWithSelfLeader(
			&sovereign.OutGoingOperationsPoolMock{},
			&sovereign.BridgeOperationsHandlerMock{},
			&block.Header{})
		success := sovEndRound.DoSovereignEndRoundJob(context.Background())
		require.False(t, success)
	})

	t.Run("no outgoing operations, should finish with success", func(t *testing.T) {
		t.Parallel()

		wasDataSent := false
		spyChan := make(chan struct{})
		bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
			SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
				wasDataSent = true
				spyChan <- struct{}{}
				return &sovCore.BridgeOperationsResponse{}, nil
			},
		}
		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
			},
			OutGoingMiniBlockHeader: nil,
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(&sovereign.OutGoingOperationsPoolMock{}, bridgeHandler, sovHdr)
		ctx := context.Background()
		success := sovEndRound.DoSovereignEndRoundJob(ctx)

		select {
		case <-spyChan:
			require.Fail(t, "should not have sent with no outgoing operations")
		case <-time.After(time.Second):
		}

		require.True(t, success)
		require.False(t, wasDataSent)
	})

	t.Run("outgoing operations found", func(t *testing.T) {
		t.Parallel()

		outGoingDataHash := []byte("hash")
		outGoingOpHash := []byte("hashOp")
		outGoingOpData := []byte("bridgeOp")
		aggregatedSig := []byte("aggregatedSig")
		leaderSig := []byte("leaderSig")
		getCallCt := 0
		pool := &sovereign.OutGoingOperationsPoolMock{
			GetCalled: func(hash []byte) *sovCore.BridgeOutGoingData {
				require.Equal(t, outGoingDataHash, hash)

				defer func() {
					getCallCt++
				}()

				switch getCallCt {
				case 0:
					return &sovCore.BridgeOutGoingData{
						Hash: outGoingDataHash,
						OutGoingOperations: []*sovCore.OutGoingOperation{
							{
								Hash: outGoingOpHash,
								Data: outGoingOpData,
							},
						},
					}
				default:
					require.Fail(t, "should not call get from pool anymore")
				}

				return nil
			},
			DeleteCalled: func(hash []byte) {
				require.Equal(t, outGoingDataHash, hash)
			},
			AddCalled: func(data *sovCore.BridgeOutGoingData) {
				require.Equal(t, &sovCore.BridgeOutGoingData{
					Hash: outGoingDataHash,
					OutGoingOperations: []*sovCore.OutGoingOperation{
						{
							Hash: outGoingOpHash,
							Data: outGoingOpData,
						},
					},
					AggregatedSignature: aggregatedSig,
					LeaderSignature:     leaderSig,
				}, data)
			},
		}

		wasDataSent := false
		currCtx := context.Background()
		wg := sync.WaitGroup{}
		wg.Add(1)
		bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
			SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
				defer func() {
					wg.Done()
				}()

				require.Equal(t, currCtx, ctx)
				require.Equal(t, &sovCore.BridgeOperations{
					Data: []*sovCore.BridgeOutGoingData{
						{
							Hash: outGoingDataHash,
							OutGoingOperations: []*sovCore.OutGoingOperation{
								{
									Hash: outGoingOpHash,
									Data: outGoingOpData,
								},
							},
							LeaderSignature:     leaderSig,
							AggregatedSignature: aggregatedSig,
						},
					},
				}, data)
				wasDataSent = true
				return &sovCore.BridgeOperationsResponse{}, nil
			},
		}

		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
			},
			OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
				OutGoingOperationsHash:                outGoingDataHash,
				AggregatedSignatureOutGoingOperations: aggregatedSig,
				LeaderSignatureOutGoingOperations:     leaderSig,
			},
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)
		wg.Wait()
		require.True(t, success)
		require.True(t, wasDataSent)
		require.Equal(t, 1, getCallCt)
	})

	t.Run("outgoing operations found with unconfirmed operations", func(t *testing.T) {
		t.Parallel()

		outGoingDataHash := []byte("hash")
		outGoingOpHash := []byte("hashOp")
		outGoingOpData := []byte("bridgeOp")
		aggregatedSig := []byte("aggregatedSig")
		leaderSig := []byte("leaderSig")
		currentBridgeOutGoingData := &sovCore.BridgeOutGoingData{
			Hash: outGoingDataHash,
			OutGoingOperations: []*sovCore.OutGoingOperation{
				{
					Hash: outGoingOpHash,
					Data: outGoingOpData,
				},
			},
			AggregatedSignature: aggregatedSig,
			LeaderSignature:     leaderSig,
		}

		unconfirmedBridgeOutGoingData := &sovCore.BridgeOutGoingData{
			Hash: []byte("hash2"),
			OutGoingOperations: []*sovCore.OutGoingOperation{
				{
					Hash: []byte("hashOp2"),
					Data: []byte("bridgeOp2"),
				},
			},
		}

		pool := &sovereign.OutGoingOperationsPoolMock{
			GetCalled: func(hash []byte) *sovCore.BridgeOutGoingData {
				return currentBridgeOutGoingData
			},
			GetUnconfirmedOperationsCalled: func() []*sovCore.BridgeOutGoingData {
				return []*sovCore.BridgeOutGoingData{unconfirmedBridgeOutGoingData}
			},
		}

		wasDataSent := false
		currCtx := context.Background()
		wg := sync.WaitGroup{}
		wg.Add(1)
		bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
			SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
				defer func() {
					wg.Done()
				}()

				require.Equal(t, currCtx, ctx)
				require.Equal(t, &sovCore.BridgeOperations{
					Data: []*sovCore.BridgeOutGoingData{
						unconfirmedBridgeOutGoingData,
						currentBridgeOutGoingData,
					},
				}, data)

				wasDataSent = true
				return &sovCore.BridgeOperationsResponse{}, nil
			},
		}

		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
			},
			OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
				OutGoingOperationsHash:                outGoingDataHash,
				AggregatedSignatureOutGoingOperations: aggregatedSig,
				LeaderSignatureOutGoingOperations:     leaderSig,
			},
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)

		wg.Wait()
		require.True(t, success)
		require.True(t, wasDataSent)
	})

	t.Run("no outgoing operations in current block, but found unconfirmed operations, leader should send them", func(t *testing.T) {
		t.Parallel()

		unconfirmedBridgeOutGoingData := &sovCore.BridgeOutGoingData{
			Hash: []byte("hashOfHashes"),
			OutGoingOperations: []*sovCore.OutGoingOperation{
				{
					Hash: []byte("hashOp1"),
					Data: []byte("bridgeOp1"),
				},
			},
		}

		pool := &sovereign.OutGoingOperationsPoolMock{
			GetUnconfirmedOperationsCalled: func() []*sovCore.BridgeOutGoingData {
				return []*sovCore.BridgeOutGoingData{unconfirmedBridgeOutGoingData}
			},
		}

		wasDataSent := false
		currCtx := context.Background()
		wg := sync.WaitGroup{}
		wg.Add(1)
		bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
			SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
				defer func() {
					wg.Done()
				}()

				require.Equal(t, currCtx, ctx)
				require.Equal(t, &sovCore.BridgeOperations{
					Data: []*sovCore.BridgeOutGoingData{
						unconfirmedBridgeOutGoingData,
					},
				}, data)

				wasDataSent = true
				return &sovCore.BridgeOperationsResponse{}, nil
			},
		}

		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
			},
			OutGoingMiniBlockHeader: nil,
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)

		wg.Wait()
		require.True(t, success)
		require.True(t, wasDataSent)
	})
}

func TestSovereignSubRoundEnd_DoEndJobByParticipant(t *testing.T) {
	t.Parallel()

	t.Run("outgoing operations found with unconfirmed, should not send them", func(t *testing.T) {
		t.Parallel()

		outGoingDataHash := []byte("hash")
		outGoingOpHash := []byte("hashOp")
		outGoingOpData := []byte("bridgeOp")
		aggregatedSig := []byte("aggregatedSig")
		leaderSig := []byte("leaderSig")
		bridgeOutGoingData := &sovCore.BridgeOutGoingData{
			Hash: outGoingDataHash,
			OutGoingOperations: []*sovCore.OutGoingOperation{
				{
					Hash: outGoingOpHash,
					Data: outGoingOpData,
				},
			},
		}

		getCallCt := 0
		getUnconfirmedCalled := 0
		pool := &sovereign.OutGoingOperationsPoolMock{
			GetCalled: func(hash []byte) *sovCore.BridgeOutGoingData {
				require.Equal(t, outGoingDataHash, hash)

				defer func() {
					getCallCt++
				}()

				switch getCallCt {
				case 0:
					return &sovCore.BridgeOutGoingData{
						Hash: outGoingDataHash,
						OutGoingOperations: []*sovCore.OutGoingOperation{
							{
								Hash: outGoingOpHash,
								Data: outGoingOpData,
							},
						},
					}
				default:
					require.Fail(t, "should not call get from pool anymore")
				}

				return nil
			},
			DeleteCalled: func(hash []byte) {
				require.Equal(t, outGoingDataHash, hash)
			},
			AddCalled: func(data *sovCore.BridgeOutGoingData) {
				require.Equal(t, &sovCore.BridgeOutGoingData{
					Hash: outGoingDataHash,
					OutGoingOperations: []*sovCore.OutGoingOperation{
						{
							Hash: outGoingOpHash,
							Data: outGoingOpData,
						},
					},
					AggregatedSignature: aggregatedSig,
					LeaderSignature:     leaderSig,
				}, data)
			},
			GetUnconfirmedOperationsCalled: func() []*sovCore.BridgeOutGoingData {
				getUnconfirmedCalled++
				return make([]*sovCore.BridgeOutGoingData, 1)
			},
		}

		wasDataSent := false
		currCtx := context.Background()
		spyChan := make(chan struct{})
		bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
			SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
				require.Equal(t, currCtx, ctx)
				require.Equal(t, &sovCore.BridgeOperations{
					Data: []*sovCore.BridgeOutGoingData{
						bridgeOutGoingData,
					},
				}, data)

				spyChan <- struct{}{}
				wasDataSent = true
				return &sovCore.BridgeOperationsResponse{}, nil
			},
		}

		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
			},
			OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
				OutGoingOperationsHash:                outGoingDataHash,
				AggregatedSignatureOutGoingOperations: aggregatedSig,
				LeaderSignatureOutGoingOperations:     leaderSig,
			},
		}
		sovEndRound := createSovSubRoundEndWithParticipant(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)

		select {
		case <-spyChan:
			require.Fail(t, "should not have sent data as participant")
		case <-time.After(time.Second):
		}

		require.True(t, success)
		require.False(t, wasDataSent)
		require.Equal(t, 1, getCallCt)
		require.Equal(t, 0, getUnconfirmedCalled)
	})

	t.Run("no outgoing operations in current block, but found unconfirmed operations, participant should NOT send them", func(t *testing.T) {
		t.Parallel()

		getUnconfirmedCalled := 0
		pool := &sovereign.OutGoingOperationsPoolMock{
			GetUnconfirmedOperationsCalled: func() []*sovCore.BridgeOutGoingData {
				getUnconfirmedCalled++
				return make([]*sovCore.BridgeOutGoingData, 1)
			},
		}

		wasDataSent := false
		currCtx := context.Background()
		spyChan := make(chan struct{})
		bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
			SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
				spyChan <- struct{}{}
				wasDataSent = true
				return &sovCore.BridgeOperationsResponse{}, nil
			},
		}

		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
			},
			OutGoingMiniBlockHeader: nil,
		}
		sovEndRound := createSovSubRoundEndWithParticipant(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)

		select {
		case <-spyChan:
			require.Fail(t, "should not have sent data as participant")
		case <-time.After(time.Second):
		}

		require.True(t, success)
		require.False(t, wasDataSent)
		require.Zero(t, getUnconfirmedCalled)
	})
}
