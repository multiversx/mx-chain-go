package bls_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	sovCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

type sovEndRoundHandler interface {
	consensus.SubroundHandler
	DoSovereignEndRoundJob(ctx context.Context) bool
	ReceivedBlockHeaderFinalInfo(cnsDta *consensus.Message) bool
	GetInternalHeader() data.HeaderHandler
}

func createSovSubRoundEndWithSelfLeader(
	pool sovereignBlock.OutGoingOperationsPool,
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
	pool sovereignBlock.OutGoingOperationsPool,
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
			OutGoingMiniBlockHeaders: nil,
		}

		wasResetTimerCalled := false
		pool := &sovereign.OutGoingOperationsPoolMock{
			ResetTimerCalled: func(hashes [][]byte) {
				wasResetTimerCalled = true
			},
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		ctx := context.Background()
		success := sovEndRound.DoSovereignEndRoundJob(ctx)

		select {
		case <-spyChan:
			require.Fail(t, "should not have sent with no outgoing operations")
		case <-time.After(time.Second):
		}

		require.True(t, success)
		require.False(t, wasDataSent)
		require.False(t, wasResetTimerCalled)
	})

	t.Run("outgoing operations found", func(t *testing.T) {
		t.Parallel()

		outGoingDataHash := []byte("hash")
		outGoingOpHash := []byte("hashOp")
		outGoingOpData := []byte("bridgeOp")
		aggregatedSig := []byte("aggregatedSig")
		leaderSig := []byte("leaderSig")
		pubKeysBitmap := []byte{0x1, 0x0}
		getCallCt := 0
		wasResetTimerCalled := false
		wg := sync.WaitGroup{}
		wg.Add(2)
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
					PubKeysBitmap:       pubKeysBitmap,
				}, data)
			},
			ResetTimerCalled: func(hashes [][]byte) {
				defer func() {
					wg.Done()
				}()

				require.Equal(t, [][]byte{outGoingDataHash}, hashes)
				wasResetTimerCalled = true
			},
		}

		wasDataSent := false
		currCtx := context.Background()
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
							PubKeysBitmap:       pubKeysBitmap,
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
			OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
				{
					OutGoingOperationsHash:                outGoingDataHash,
					AggregatedSignatureOutGoingOperations: aggregatedSig,
					LeaderSignatureOutGoingOperations:     leaderSig,
				},
			},
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)
		wg.Wait()
		require.True(t, success)
		require.True(t, wasDataSent)
		require.True(t, wasResetTimerCalled)
		require.Equal(t, 1, getCallCt)
	})

	t.Run("outgoing operations in multiple outgoing mbs found", func(t *testing.T) {
		t.Parallel()

		outGoingDataHash1 := []byte("hash1")
		outGoingOpHash1 := []byte("hashOp1")
		outGoingOpData1 := []byte("bridgeOp1")
		aggregatedSig1 := []byte("aggregatedSig1")
		leaderSig1 := []byte("leaderSig1")

		outGoingDataHash2 := []byte("hash2")
		outGoingOpHash2 := []byte("hashOp2")
		outGoingOpData2 := []byte("bridgeOp2")
		aggregatedSig2 := []byte("aggregatedSig2")
		leaderSig2 := []byte("leaderSig2")

		wg := sync.WaitGroup{}
		wg.Add(1)
		currentBridgeOutGoingData1 := &sovCore.BridgeOutGoingData{
			Type: int32(block.OutGoingMbTx),
			Hash: outGoingDataHash1,
			OutGoingOperations: []*sovCore.OutGoingOperation{
				{
					Hash: outGoingOpHash1,
					Data: outGoingOpData1,
				},
			},
			AggregatedSignature: aggregatedSig1,
			LeaderSignature:     leaderSig1,
			PubKeysBitmap:       []byte{0x1, 0x0},
			Epoch:               4,
		}

		currentBridgeOutGoingData2 := &sovCore.BridgeOutGoingData{
			Type: int32(block.OutGoingMbChangeValidatorSet),
			Hash: outGoingDataHash2,
			OutGoingOperations: []*sovCore.OutGoingOperation{
				{
					Hash: outGoingOpHash2,
					Data: outGoingOpData2,
				},
			},
			AggregatedSignature: aggregatedSig2,
			LeaderSignature:     leaderSig2,
			PubKeysBitmap:       []byte{0x1, 0x0},
			Epoch:               4,
		}

		pool := &sovereign.OutGoingOperationsPoolMock{
			GetCalled: func(hash []byte) *sovCore.BridgeOutGoingData {
				switch string(hash) {
				case string(outGoingDataHash1):
					return &sovCore.BridgeOutGoingData{
						Type: int32(block.OutGoingMbTx),
						Hash: outGoingDataHash1,
						OutGoingOperations: []*sovCore.OutGoingOperation{
							{
								Hash: outGoingOpHash1,
								Data: outGoingOpData1,
							},
						},
						AggregatedSignature: nil, // no signature
						LeaderSignature:     nil, // no signature
						Epoch:               4,
					}
				case string(outGoingDataHash2):
					return &sovCore.BridgeOutGoingData{
						Type: int32(block.OutGoingMbChangeValidatorSet),
						Hash: outGoingDataHash2,
						OutGoingOperations: []*sovCore.OutGoingOperation{
							{
								Hash: outGoingOpHash2,
								Data: outGoingOpData2,
							},
						},
						AggregatedSignature: nil, // no signature
						LeaderSignature:     nil, // no signature
						Epoch:               4,
					}
				}

				require.Fail(t, "should not request any other bridge data from pool")
				return nil
			},
		}

		wasDataSent := false
		currCtx := context.Background()
		bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
			SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
				defer func() {
					wg.Done()
				}()

				require.Equal(t, currCtx, ctx)
				require.Equal(t, &sovCore.BridgeOperations{
					Data: []*sovCore.BridgeOutGoingData{
						currentBridgeOutGoingData1,
						currentBridgeOutGoingData2,
					},
				}, data)

				wasDataSent = true
				return &sovCore.BridgeOperationsResponse{}, nil
			},
		}

		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
				Epoch: 4,
			},
			OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
				{
					Type:                                  block.OutGoingMbTx,
					OutGoingOperationsHash:                outGoingDataHash1,
					AggregatedSignatureOutGoingOperations: aggregatedSig1,
					LeaderSignatureOutGoingOperations:     leaderSig1,
				},
				{
					Type:                                  block.OutGoingMbChangeValidatorSet,
					OutGoingOperationsHash:                outGoingDataHash2,
					AggregatedSignatureOutGoingOperations: aggregatedSig2,
					LeaderSignatureOutGoingOperations:     leaderSig2,
				},
			},
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)

		wg.Wait()
		require.True(t, success)
		require.True(t, wasDataSent)
	})

	t.Run("outgoing operations found with unconfirmed operations", func(t *testing.T) {
		t.Parallel()

		outGoingDataHash := []byte("hash")
		outGoingOpHash := []byte("hashOp")
		outGoingOpData := []byte("bridgeOp")
		aggregatedSig := []byte("aggregatedSig")
		leaderSig := []byte("leaderSig")
		wg := sync.WaitGroup{}
		wg.Add(2)
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
			Epoch:               4,
		}

		unconfirmedBridgeOutGoingData := &sovCore.BridgeOutGoingData{
			Hash: []byte("hash2"),
			OutGoingOperations: []*sovCore.OutGoingOperation{
				{
					Hash: []byte("hashOp2"),
					Data: []byte("bridgeOp2"),
				},
			},
			Epoch: 3,
		}

		wasResetTimerCalled := false
		pool := &sovereign.OutGoingOperationsPoolMock{
			GetCalled: func(hash []byte) *sovCore.BridgeOutGoingData {
				return currentBridgeOutGoingData
			},
			GetUnconfirmedOperationsCalled: func() []*sovCore.BridgeOutGoingData {
				return []*sovCore.BridgeOutGoingData{unconfirmedBridgeOutGoingData}
			},
			ResetTimerCalled: func(hashes [][]byte) {
				defer func() {
					wg.Done()
				}()

				require.Equal(t, [][]byte{unconfirmedBridgeOutGoingData.Hash, currentBridgeOutGoingData.Hash}, hashes)
				wasResetTimerCalled = true
			},
		}

		wasDataSent := false
		currCtx := context.Background()
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
				Epoch: 4,
			},
			OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
				{
					OutGoingOperationsHash:                outGoingDataHash,
					AggregatedSignatureOutGoingOperations: aggregatedSig,
					LeaderSignatureOutGoingOperations:     leaderSig,
				},
			},
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)

		wg.Wait()
		require.True(t, success)
		require.True(t, wasDataSent)
		require.True(t, wasResetTimerCalled)
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

		wasResetTimerCalled := false
		wg := sync.WaitGroup{}
		wg.Add(2)
		pool := &sovereign.OutGoingOperationsPoolMock{
			GetUnconfirmedOperationsCalled: func() []*sovCore.BridgeOutGoingData {
				return []*sovCore.BridgeOutGoingData{unconfirmedBridgeOutGoingData}
			},
			ResetTimerCalled: func(hashes [][]byte) {
				defer func() {
					wg.Done()
				}()

				wasResetTimerCalled = true
				require.Equal(t, [][]byte{unconfirmedBridgeOutGoingData.Hash}, hashes)
			},
		}

		wasDataSent := false
		currCtx := context.Background()
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
			OutGoingMiniBlockHeaders: nil,
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		success := sovEndRound.DoSovereignEndRoundJob(currCtx)

		wg.Wait()
		require.True(t, success)
		require.True(t, wasDataSent)
		require.True(t, wasResetTimerCalled)
	})

	t.Run("leader keeps resending previous unconfirmed operations, should keep signatures in pool", func(t *testing.T) {
		t.Parallel()

		epoch := uint32(4)
		unconfirmedBridgeOutGoingData := &sovCore.BridgeOutGoingData{
			Hash: []byte("hashOfHashes"),
			OutGoingOperations: []*sovCore.OutGoingOperation{
				{
					Hash: []byte("hashOp1"),
					Data: []byte("bridgeOp1"),
				},
			},
			Epoch: epoch,
		}

		expiryTime := time.Millisecond * 100
		pool := sovereignBlock.NewOutGoingOperationPool(expiryTime)
		pool.Add(unconfirmedBridgeOutGoingData)

		pubKeysBitmap := []byte{0x1, 0x0}
		sendDataCalledCt := 0
		wasDataSent := false
		currCtx := context.Background()
		wg := sync.WaitGroup{}
		wg.Add(1)
		bridgeData := &sovCore.BridgeOperations{
			Data: []*sovCore.BridgeOutGoingData{
				{
					Hash: []byte("hashOfHashes"),
					OutGoingOperations: []*sovCore.OutGoingOperation{
						{
							Hash: []byte("hashOp1"),
							Data: []byte("bridgeOp1"),
						},
					},
					AggregatedSignature: []byte("aggregatedSig"),
					LeaderSignature:     []byte("leaderSig"),
					PubKeysBitmap:       pubKeysBitmap,
					Epoch:               epoch,
				},
			},
		}
		bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
			SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
				defer func() {
					wg.Done()
				}()

				require.Equal(t, currCtx, ctx)
				require.Equal(t, bridgeData, data)

				wasDataSent = true
				sendDataCalledCt++
				return &sovCore.BridgeOperationsResponse{}, nil
			},
		}

		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Nonce: 4,
				Epoch: epoch,
			},
			OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
				{
					Hash:                                  []byte("hashOfHashes"),
					OutGoingOperationsHash:                []byte("hashOfHashes"),
					AggregatedSignatureOutGoingOperations: []byte("aggregatedSig"),
					LeaderSignatureOutGoingOperations:     []byte("leaderSig"),
				},
			},
		}
		sovEndRound := createSovSubRoundEndWithSelfLeader(pool, bridgeHandler, sovHdr)
		time.Sleep(expiryTime)

		// Before having confirmed signed header, we have in pool data with no signatures
		require.Equal(t, []*sovCore.BridgeOutGoingData{unconfirmedBridgeOutGoingData}, pool.GetUnconfirmedOperations())

		for i := 1; i <= 3; i++ {
			success := false
			wasDataSent = false

			success = sovEndRound.DoSovereignEndRoundJob(currCtx)
			wg.Wait()
			require.True(t, success)
			require.True(t, wasDataSent)
			require.Equal(t, i, sendDataCalledCt)

			time.Sleep(expiryTime)

			// Simulate next rounds with no further outgoing operations
			sovHdr.OutGoingMiniBlockHeaders = nil
			wg.Add(1)

			// After first trying to send it, we should always keep in pool data with signatures
			require.Equal(t, []*sovCore.BridgeOutGoingData{bridgeData.Data[0]}, pool.GetUnconfirmedOperations())
		}
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
		wasResetTimerCalled := false
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
			ResetTimerCalled: func(hashes [][]byte) {
				wasResetTimerCalled = true
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
			OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
				{
					OutGoingOperationsHash:                outGoingDataHash,
					AggregatedSignatureOutGoingOperations: aggregatedSig,
					LeaderSignatureOutGoingOperations:     leaderSig,
				},
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
		require.False(t, wasResetTimerCalled)
		require.Equal(t, 1, getCallCt)
		require.Equal(t, 0, getUnconfirmedCalled)
	})

	t.Run("no outgoing operations in current block, but found unconfirmed operations, participant should NOT send them", func(t *testing.T) {
		t.Parallel()

		getUnconfirmedCalled := 0
		wasResetTimerCalled := false
		pool := &sovereign.OutGoingOperationsPoolMock{
			GetUnconfirmedOperationsCalled: func() []*sovCore.BridgeOutGoingData {
				getUnconfirmedCalled++
				return make([]*sovCore.BridgeOutGoingData, 1)
			},
			ResetTimerCalled: func(hashes [][]byte) {
				wasResetTimerCalled = true
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
			OutGoingMiniBlockHeaders: nil,
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
		require.False(t, wasResetTimerCalled)
		require.Zero(t, getUnconfirmedCalled)
	})
}

func TestSovereignSubRoundEnd_ReceivedBlockHeaderFinalInfo(t *testing.T) {
	t.Parallel()

	outGoingData := &sovCore.BridgeOutGoingData{
		Hash: []byte("hashOfHashes"),
		OutGoingOperations: []*sovCore.OutGoingOperation{
			{
				Hash: []byte("hashOp1"),
				Data: []byte("bridgeOp1"),
			},
		},
	}

	pool := sovereignBlock.NewOutGoingOperationPool(time.Second)
	pool.Add(outGoingData)

	wasDataSent := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	bridgeHandler := &sovereign.BridgeOperationsHandlerMock{
		SendCalled: func(ctx context.Context, data *sovCore.BridgeOperations) (*sovCore.BridgeOperationsResponse, error) {
			defer func() {
				wg.Done()
			}()

			wasDataSent = true
			return &sovCore.BridgeOperationsResponse{}, nil
		},
	}
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce: 4,
		},
		OutGoingMiniBlockHeaders: []*block.OutGoingMiniBlockHeader{
			{
				Hash:                   []byte("hashOfHashes"),
				OutGoingOperationsHash: []byte("hashOfHashes"),
			},
		},
	}

	sovEndRound := createSovSubRoundEndWithParticipant(pool, bridgeHandler, sovHdr)

	aggregatedSig := []byte("aggregatedSigOutGoing")
	leaderSig := []byte("leaderSigOutGoing")
	cnsData := consensus.Message{
		HeaderHash:     []byte("X"),
		PubKey:         []byte("A"),
		InvalidSigners: []byte("invalidSignersData"),
		ExtraSignatures: map[string]*consensus.ExtraSignatureData{
			block.OutGoingMbTx.String(): {
				AggregatedSignatureOutGoingTxData: aggregatedSig,
				LeaderSignatureOutGoingTxData:     leaderSig,
			},
		},
	}

	// Participant should not send any data
	res := sovEndRound.ReceivedBlockHeaderFinalInfo(&cnsData)
	require.True(t, res)
	require.False(t, wasDataSent)

	// Header's outgoing mb is updated with signatures from consensus message
	outGoingMb := sovEndRound.GetInternalHeader().(data.SovereignChainHeaderHandler).GetOutGoingMiniBlockHeaderHandler(int32(block.OutGoingMbTx))
	require.Equal(t, leaderSig, outGoingMb.GetLeaderSignatureOutGoingOperations())
	require.Equal(t, aggregatedSig, outGoingMb.GetAggregatedSignatureOutGoingOperations())

	// Internal outgoing pool is updated with signatures as well
	updatedPoolData := pool.Get([]byte("hashOfHashes"))
	require.Equal(t, &sovCore.BridgeOutGoingData{
		Hash: []byte("hashOfHashes"),
		OutGoingOperations: []*sovCore.OutGoingOperation{
			{
				Hash: []byte("hashOp1"),
				Data: []byte("bridgeOp1"),
			},
		},
		AggregatedSignature: aggregatedSig,
		LeaderSignature:     leaderSig,
	}, updatedPoolData)
}
