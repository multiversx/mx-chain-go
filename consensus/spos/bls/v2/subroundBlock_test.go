package v2_test

import (
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

var expectedErr = errors.New("expected error")

func defaultSubroundForSRBlock(consensusState *spos.ConsensusState, ch chan bool,
	container *spos.ConsensusCore, appStatusHandler core.AppStatusHandler) (*spos.Subround, error) {
	return spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		appStatusHandler,
	)
}

func createDefaultHeader() *block.Header {
	return &block.Header{
		Nonce:           1,
		PrevHash:        []byte("prev hash"),
		PrevRandSeed:    []byte("prev rand seed"),
		RandSeed:        []byte("rand seed"),
		RootHash:        []byte("roothash"),
		TxCount:         0,
		ChainID:         []byte("chain ID"),
		SoftwareVersion: []byte("software version"),
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
}

func defaultSubroundBlockFromSubround(sr *spos.Subround) (v2.SubroundBlock, error) {
	srBlock, err := v2.NewSubroundBlock(
		sr,
		v2.ProcessingThresholdPercent,
		&consensusMocks.SposWorkerMock{},
	)

	return srBlock, err
}

func defaultSubroundBlockWithoutErrorFromSubround(sr *spos.Subround) v2.SubroundBlock {
	srBlock, _ := v2.NewSubroundBlock(
		sr,
		v2.ProcessingThresholdPercent,
		&consensusMocks.SposWorkerMock{},
	)

	return srBlock
}

func initSubroundBlock(
	blockChain data.ChainHandler,
	container *spos.ConsensusCore,
	appStatusHandler core.AppStatusHandler,
) v2.SubroundBlock {
	if blockChain == nil {
		blockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Nonce:     uint64(0),
					Signature: []byte("genesis signature"),
					RandSeed:  []byte{0},
				}
			},
			GetGenesisHeaderHashCalled: func() []byte {
				return []byte("genesis header hash")
			},
		}
	}

	consensusState := initializers.InitConsensusStateWithNodesCoordinator(container.NodesCoordinator())
	ch := make(chan bool, 1)

	container.SetBlockchain(blockChain)

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, appStatusHandler)
	srBlock, _ := defaultSubroundBlockFromSubround(sr)
	return srBlock
}

func createConsensusContainers() []*spos.ConsensusCore {
	consensusContainers := make([]*spos.ConsensusCore, 0)
	container := consensusMocks.InitConsensusCore()
	consensusContainers = append(consensusContainers, container)
	container = consensusMocks.InitConsensusCoreHeaderV2()
	consensusContainers = append(consensusContainers, container)
	return consensusContainers
}

func initSubroundBlockWithBlockProcessor(
	bp *testscommon.BlockProcessorStub,
	container *spos.ConsensusCore,
) v2.SubroundBlock {
	blockChain := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Nonce:     uint64(0),
				Signature: []byte("genesis signature"),
			}
		},
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("genesis header hash")
		},
	}
	blockProcessorMock := bp

	container.SetBlockchain(blockChain)
	container.SetBlockProcessor(blockProcessorMock)
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})
	srBlock, _ := defaultSubroundBlockFromSubround(sr)
	return srBlock
}

func TestSubroundBlock_NewSubroundBlockNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srBlock, err := v2.NewSubroundBlock(
		nil,
		v2.ProcessingThresholdPercent,
		&consensusMocks.SposWorkerMock{},
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundBlock_NewSubroundBlockNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetBlockchain(nil)

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundBlock_NewSubroundBlockNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetBlockProcessor(nil)

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubroundBlock_NewSubroundBlockNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	sr.ConsensusStateHandler = nil

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundBlock_NewSubroundBlockNilHasherShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetHasher(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubroundBlock_NewSubroundBlockNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetMarshalizer(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestSubroundBlock_NewSubroundBlockNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetMultiSignerContainer(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestSubroundBlock_NewSubroundBlockNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetRoundHandler(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubroundBlock_NewSubroundBlockNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetShardCoordinator(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestSubroundBlock_NewSubroundBlockNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetSyncTimer(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundBlock_NewSubroundBlockNilWorkerShouldFail(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	srBlock, err := v2.NewSubroundBlock(
		sr,
		v2.ProcessingThresholdPercent,
		nil,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilWorker, err)
}

func TestSubroundBlock_NewSubroundBlockShouldWork(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.NotNil(t, srBlock)
	assert.Nil(t, err)
}

func TestSubroundBlock_DoBlockJob(t *testing.T) {
	t.Parallel()

	t.Run("not leader should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("round index lower than last committed block should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)
		_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrBlock, true)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("leader job done should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)
		_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrBlock, true)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("subround finished should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)
		_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrBlock, false)
		sr.SetStatus(bls.SrBlock, spos.SsFinished)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("create header error should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)
		sr.SetStatus(bls.SrBlock, spos.SsNotFinished)
		bpm := &testscommon.BlockProcessorStub{}

		bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return nil, expectedErr
		}
		container.SetBlockProcessor(bpm)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("create block error should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)
		sr.SetStatus(bls.SrBlock, spos.SsNotFinished)
		bpm := &testscommon.BlockProcessorStub{}
		bpm.CreateBlockCalled = func(header data.HeaderHandler, remainingTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return header, nil, expectedErr
		}
		bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return &block.Header{}, nil
		}
		container.SetBlockProcessor(bpm)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("sign block header failure should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})

		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)

		bpm := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
		container.SetBlockProcessor(bpm)
		cnt := uint32(0)
		sh := &consensusMocks.SigningHandlerStub{
			CreateSignatureForPublicKeyCalled: func(message []byte, publicKeyBytes []byte) ([]byte, error) {
				cnt++
				if cnt > 1 { // first call is from create header
					return nil, expectedErr
				}

				return []byte("sig"), nil
			},
		}
		container.SetSigningHandler(sh)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("set leader signature failure should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})

		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)

		bpm := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
		bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return &testscommon.HeaderHandlerStub{
				SetLeaderSignatureCalled: func(signature []byte) error {
					if len(signature) > 0 {
						return expectedErr
					}
					return nil
				},
				CloneCalled: func() data.HeaderHandler {
					return &block.HeaderV2{
						Header: &block.Header{},
					}
				},
			}, nil
		}
		container.SetBlockProcessor(bpm)
		sh := &consensusMocks.SigningHandlerStub{
			CreateSignatureForPublicKeyCalled: func(message []byte, publicKeyBytes []byte) ([]byte, error) {
				return []byte("sig"), nil
			},
		}
		container.SetSigningHandler(sh)

		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("send block error should return false", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})

		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)
		bpm := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
		container.SetBlockProcessor(bpm)
		bm := &consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return expectedErr
			},
		}
		container.SetBroadcastMessenger(bm)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedSignature := []byte("provided signature")
		providedBitmap := []byte("provided bitmap")
		providedHash := []byte("provided hash")
		providedHeadr := &block.HeaderV2{
			Header: &block.Header{
				Signature:     []byte("signature"),
				PubKeysBitmap: []byte("bitmap"),
			},
		}

		container := consensusMocks.InitConsensusCore()
		chainHandler := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return providedHeadr
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return providedHash
			},
		}
		container.SetBlockchain(chainHandler)

		consensusState := initializers.InitConsensusStateWithNodesCoordinator(container.NodesCoordinator())
		ch := make(chan bool, 1)

		baseSr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})
		sr, _ := v2.NewSubroundBlock(
			baseSr,
			v2.ProcessingThresholdPercent,
			&consensusMocks.SposWorkerMock{},
		)

		providedLeaderSignature := []byte("leader signature")
		container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
			CreateSignatureForPublicKeyCalled: func(message []byte, publicKeyBytes []byte) ([]byte, error) {
				return providedLeaderSignature, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig []byte, msg []byte, epoch uint32) error {
				assert.Fail(t, "should have not been called for leader")
				return nil
			},
		})
		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		leader, err := sr.GetLeader()
		assert.Nil(t, err)

		sr.SetSelfPubKey(leader)
		bpm := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
		container.SetBlockProcessor(bpm)
		bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return &block.HeaderV2{
				Header: &block.Header{
					Round: round,
					Nonce: nonce,
				},
			}, nil
		}
		bm := &consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return nil
			},
		}
		container.SetBroadcastMessenger(bm)
		container.SetRoundHandler(&consensusMocks.RoundHandlerMock{
			RoundIndex: 1,
		})
		container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				return &block.HeaderProof{
					HeaderHash:          headerHash,
					AggregatedSignature: providedSignature,
					PubKeysBitmap:       providedBitmap,
				}, nil
			},
		})

		r := sr.DoBlockJob()
		assert.True(t, r)
		assert.Equal(t, uint64(1), sr.GetHeader().GetNonce())
	})
}

func TestSubroundBlock_ReceivedBlock(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	blkBody := &block.Body{}
	blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)
	leader, err := sr.GetLeader()
	assert.Nil(t, err)
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		blkBodyStr,
		nil,
		[]byte(leader),
		[]byte("sig"),
		int(bls.MtBlockBody),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)
	sr.SetBody(&block.Body{})
	r := sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.SetBody(nil)
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0])
	sr.SetStatus(bls.SrBlock, spos.SsFinished)
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(bls.SrBlock, spos.SsNotFinished)
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	leader, _ := sr.GetLeader()
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		nil,
		nil,
		[]byte(leader),
		[]byte("sig"),
		int(bls.MtBlockBodyAndHeader),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockFails(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	blProcMock := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
	err := errors.New("error process block")
	blProcMock.ProcessBlockCalled = func(data.HeaderHandler, data.BodyHandler, func() time.Duration) error {
		return err
	}
	container.SetBlockProcessor(blProcMock)
	hdr := &block.Header{}
	blkBody := &block.Body{}
	blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)
	leader, _ := sr.GetLeader()
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		blkBodyStr,
		nil,
		[]byte(leader),
		[]byte("sig"),
		int(bls.MtBlockBody),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)
	sr.SetHeader(hdr)
	sr.SetBody(blkBody)
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	hdr := &block.Header{}
	blkBody := &block.Body{}
	blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)
	leader, _ := sr.GetLeader()
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		blkBodyStr,
		nil,
		[]byte(leader),
		[]byte("sig"),
		int(bls.MtBlockBody),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)
	sr.SetHeader(hdr)
	sr.SetBody(blkBody)
	blockProcessorMock := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
	blockProcessorMock.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return expectedErr
	}
	container.SetBlockProcessor(blockProcessorMock)
	container.SetRoundHandler(&consensusMocks.RoundHandlerMock{RoundIndex: 1})
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnTrue(t *testing.T) {
	t.Parallel()

	consensusContainers := createConsensusContainers()
	for _, container := range consensusContainers {
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		hdr, _ := container.BlockProcessor().CreateNewHeader(1, 1)
		hdr, blkBody, _ := container.BlockProcessor().CreateBlock(hdr, func() bool { return true })

		blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)
		leader, _ := sr.GetLeader()
		cnsMsg := consensus.NewConsensusMessage(
			nil,
			nil,
			blkBodyStr,
			nil,
			[]byte(leader),
			[]byte("sig"),
			int(bls.MtBlockBody),
			0,
			chainID,
			nil,
			nil,
			nil,
			currentPid,
			nil,
		)
		sr.SetHeader(hdr)
		sr.SetBody(blkBody)
		assert.True(t, sr.ProcessReceivedBlock(cnsMsg))
	}
}

func TestSubroundBlock_RemainingTimeShouldReturnNegativeValue(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	roundHandlerMock := initRoundHandlerMock()
	container.SetRoundHandler(roundHandlerMock)

	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	remainingTimeInThisRound := func() time.Duration {
		roundStartTime := sr.RoundHandler().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.RoundHandler().TimeDuration()*85/100 - elapsedTime

		return remainingTime
	}
	container.SetSyncTimer(&consensusMocks.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 84 / 100)
	}})
	ret := remainingTimeInThisRound()
	assert.True(t, ret > 0)

	container.SetSyncTimer(&consensusMocks.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 85 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret == 0)

	container.SetSyncTimer(&consensusMocks.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 86 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret < 0)
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	sr.SetRoundCanceled(true)
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	sr.SetStatus(bls.SrBlock, spos.SsFinished)
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenBlockIsReceivedReturnTrue(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	for i := 0; i < sr.Threshold(bls.SrBlock); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrBlock, true)
	}
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenBlockIsReceivedReturnFalse(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_IsBlockReceived(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrBlock, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, false)
	}
	ok := sr.IsBlockReceived(1)
	assert.False(t, ok)

	_ = sr.SetJobDone("A", bls.SrBlock, true)
	isJobDone, _ := sr.JobDone("A", bls.SrBlock)
	assert.True(t, isJobDone)

	ok = sr.IsBlockReceived(1)
	assert.True(t, ok)

	ok = sr.IsBlockReceived(2)
	assert.False(t, ok)
}

func TestSubroundBlock_HaveTimeInCurrentSubroundShouldReturnTrue(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	haveTimeInCurrentSubound := func() bool {
		roundStartTime := sr.RoundHandler().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.EndTime() - int64(elapsedTime)

		return time.Duration(remainingTime) > 0
	}
	roundHandlerMock := &consensusMocks.RoundHandlerMock{}
	roundHandlerMock.TimeDurationCalled = func() time.Duration {
		return 4000 * time.Millisecond
	}
	roundHandlerMock.TimeStampCalled = func() time.Time {
		return time.Unix(0, 0)
	}
	syncTimerMock := &consensusMocks.SyncTimerMock{}
	timeElapsed := sr.EndTime() - 1
	syncTimerMock.CurrentTimeCalled = func() time.Time {
		return time.Unix(0, timeElapsed)
	}
	container.SetRoundHandler(roundHandlerMock)
	container.SetSyncTimer(syncTimerMock)

	assert.True(t, haveTimeInCurrentSubound())
}

func TestSubroundBlock_HaveTimeInCurrentSuboundShouldReturnFalse(t *testing.T) {
	t.Parallel()
	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	haveTimeInCurrentSubound := func() bool {
		roundStartTime := sr.RoundHandler().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.EndTime() - int64(elapsedTime)

		return time.Duration(remainingTime) > 0
	}
	roundHandlerMock := &consensusMocks.RoundHandlerMock{}
	roundHandlerMock.TimeDurationCalled = func() time.Duration {
		return 4000 * time.Millisecond
	}
	roundHandlerMock.TimeStampCalled = func() time.Time {
		return time.Unix(0, 0)
	}
	syncTimerMock := &consensusMocks.SyncTimerMock{}
	timeElapsed := sr.EndTime() + 1
	syncTimerMock.CurrentTimeCalled = func() time.Time {
		return time.Unix(0, timeElapsed)
	}
	container.SetRoundHandler(roundHandlerMock)
	container.SetSyncTimer(syncTimerMock)

	assert.False(t, haveTimeInCurrentSubound())
}

func TestSubroundBlock_CreateHeaderNilCurrentHeader(t *testing.T) {
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return nil
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Nonce:     uint64(0),
				Signature: []byte("genesis signature"),
				RandSeed:  []byte{0},
			}
		},
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("genesis header hash")
		},
	}

	consensusContainers := createConsensusContainers()
	for _, container := range consensusContainers {
		sr := initSubroundBlock(blockChain, container, &statusHandler.AppStatusHandlerStub{})
		_ = sr.BlockChain().SetCurrentBlockHeaderAndRootHash(nil, nil)
		header, _ := sr.CreateHeader()
		header, body, _ := sr.CreateBlock(header)
		marshalizedBody, _ := sr.Marshalizer().Marshal(body)
		marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
		_ = sr.SendBlockBody(body, marshalizedBody)
		_ = sr.SendBlockHeader(header, marshalizedHeader)

		expectedHeader, _ := container.BlockProcessor().CreateNewHeader(uint64(sr.RoundHandler().Index()), uint64(1))
		err := expectedHeader.SetTimeStamp(uint64(sr.RoundHandler().TimeStamp().Unix()))
		require.Nil(t, err)
		err = expectedHeader.SetRootHash([]byte{})
		require.Nil(t, err)
		err = expectedHeader.SetPrevHash(sr.BlockChain().GetGenesisHeaderHash())
		require.Nil(t, err)
		err = expectedHeader.SetPrevRandSeed(sr.BlockChain().GetGenesisHeader().GetRandSeed())
		require.Nil(t, err)
		err = expectedHeader.SetRandSeed(make([]byte, 0))
		require.Nil(t, err)
		err = expectedHeader.SetMiniBlockHeaderHandlers(header.GetMiniBlockHeaderHandlers())
		require.Nil(t, err)
		err = expectedHeader.SetChainID(chainID)
		require.Nil(t, err)
		require.Equal(t, expectedHeader, header)
	}
}

func TestSubroundBlock_CreateHeaderNotNilCurrentHeader(t *testing.T) {
	consensusContainers := createConsensusContainers()
	for _, container := range consensusContainers {
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		_ = sr.BlockChain().SetCurrentBlockHeaderAndRootHash(&block.Header{
			Nonce: 1,
		}, []byte("root hash"))

		header, _ := sr.CreateHeader()
		header, body, _ := sr.CreateBlock(header)
		marshalizedBody, _ := sr.Marshalizer().Marshal(body)
		marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
		_ = sr.SendBlockBody(body, marshalizedBody)
		_ = sr.SendBlockHeader(header, marshalizedHeader)

		expectedHeader, _ := container.BlockProcessor().CreateNewHeader(
			uint64(sr.RoundHandler().Index()),
			sr.BlockChain().GetCurrentBlockHeader().GetNonce()+1)
		err := expectedHeader.SetTimeStamp(uint64(sr.RoundHandler().TimeStamp().Unix()))
		require.Nil(t, err)
		err = expectedHeader.SetRootHash([]byte{})
		require.Nil(t, err)
		err = expectedHeader.SetPrevHash(sr.BlockChain().GetCurrentBlockHeaderHash())
		require.Nil(t, err)
		err = expectedHeader.SetRandSeed(make([]byte, 0))
		require.Nil(t, err)
		err = expectedHeader.SetMiniBlockHeaderHandlers(header.GetMiniBlockHeaderHandlers())
		require.Nil(t, err)
		err = expectedHeader.SetChainID(chainID)
		require.Nil(t, err)
		require.Equal(t, expectedHeader, header)
	}
}

func TestSubroundBlock_CreateHeaderMultipleMiniBlocks(t *testing.T) {
	mbHeaders := []block.MiniBlockHeader{
		{Hash: []byte("mb1"), SenderShardID: 1, ReceiverShardID: 1},
		{Hash: []byte("mb2"), SenderShardID: 1, ReceiverShardID: 2},
		{Hash: []byte("mb3"), SenderShardID: 2, ReceiverShardID: 3},
	}
	blockChainMock := testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Nonce: 1,
			}
		},
	}
	container := consensusMocks.InitConsensusCore()
	bp := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
	bp.CreateBlockCalled = func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		shardHeader, _ := header.(*block.Header)
		shardHeader.MiniBlockHeaders = mbHeaders
		shardHeader.RootHash = []byte{}

		return shardHeader, &block.Body{}, nil
	}
	sr := initSubroundBlockWithBlockProcessor(bp, container)
	container.SetBlockchain(&blockChainMock)

	header, _ := sr.CreateHeader()
	header, body, _ := sr.CreateBlock(header)
	marshalizedBody, _ := sr.Marshalizer().Marshal(body)
	marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
	_ = sr.SendBlockBody(body, marshalizedBody)
	_ = sr.SendBlockHeader(header, marshalizedHeader)

	expectedHeader := &block.Header{
		Round:            uint64(sr.RoundHandler().Index()),
		TimeStamp:        uint64(sr.RoundHandler().TimeStamp().Unix()),
		RootHash:         []byte{},
		Nonce:            sr.BlockChain().GetCurrentBlockHeader().GetNonce() + 1,
		PrevHash:         sr.BlockChain().GetCurrentBlockHeaderHash(),
		RandSeed:         make([]byte, 0),
		MiniBlockHeaders: mbHeaders,
		ChainID:          chainID,
	}

	assert.Equal(t, expectedHeader, header)
}

func TestSubroundBlock_CreateHeaderNilMiniBlocks(t *testing.T) {
	expectedErr := errors.New("nil mini blocks")
	container := consensusMocks.InitConsensusCore()
	bp := consensusMocks.InitBlockProcessorMock(container.Marshalizer())
	bp.CreateBlockCalled = func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		return nil, nil, expectedErr
	}
	sr := initSubroundBlockWithBlockProcessor(bp, container)
	_ = sr.BlockChain().SetCurrentBlockHeaderAndRootHash(&block.Header{
		Nonce: 1,
	}, []byte("root hash"))
	header, _ := sr.CreateHeader()
	_, _, err := sr.CreateBlock(header)
	assert.Equal(t, expectedErr, err)
}

func TestSubroundBlock_CallFuncRemainingTimeWithStructShouldWork(t *testing.T) {
	roundStartTime := time.Now()
	maxTime := 100 * time.Millisecond
	newRoundStartTime := roundStartTime
	remainingTimeInCurrentRound := func() time.Duration {
		return RemainingTimeWithStruct(newRoundStartTime, maxTime)
	}
	assert.True(t, remainingTimeInCurrentRound() > 0)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, remainingTimeInCurrentRound() < 0)
}

func TestSubroundBlock_CallFuncRemainingTimeWithStructShouldNotWork(t *testing.T) {
	roundStartTime := time.Now()
	maxTime := 100 * time.Millisecond
	remainingTimeInCurrentRound := func() time.Duration {
		return RemainingTimeWithStruct(roundStartTime, maxTime)
	}
	assert.True(t, remainingTimeInCurrentRound() > 0)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, remainingTimeInCurrentRound() < 0)

	roundStartTime = roundStartTime.Add(500 * time.Millisecond)
	assert.False(t, remainingTimeInCurrentRound() < 0)
}

func RemainingTimeWithStruct(startTime time.Time, maxTime time.Duration) time.Duration {
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(startTime)
	remainingTime := maxTime - elapsedTime
	return remainingTime
}

func TestSubroundBlock_ReceivedBlockComputeProcessDuration(t *testing.T) {
	t.Parallel()

	srStartTime := int64(5 * roundTimeDuration / 100)
	srEndTime := int64(25 * roundTimeDuration / 100)
	srDuration := srEndTime - srStartTime
	delay := srDuration * 430 / 1000

	container := consensusMocks.InitConsensusCore()
	receivedValue := uint64(0)
	container.SetBlockProcessor(&testscommon.BlockProcessorStub{
		ProcessBlockCalled: func(_ data.HeaderHandler, _ data.BodyHandler, _ func() time.Duration) error {
			time.Sleep(time.Duration(delay))
			return nil
		},
	})
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			receivedValue = value
		}})
	hdr := &block.Header{}
	blkBody := &block.Body{}
	blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)

	leader, err := sr.GetLeader()
	assert.Nil(t, err)
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		blkBodyStr,
		nil,
		[]byte(leader),
		[]byte("sig"),
		int(bls.MtBlockBody),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)
	sr.SetHeader(hdr)
	sr.SetBody(blkBody)

	minimumExpectedValue := uint64(delay * 100 / srDuration)
	_ = sr.ProcessReceivedBlock(cnsMsg)

	assert.True(t,
		receivedValue >= minimumExpectedValue,
		fmt.Sprintf("minimum expected was %d, got %d", minimumExpectedValue, receivedValue),
	)
}

func TestSubroundBlock_ReceivedBlockComputeProcessDurationWithZeroDurationShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced", r)
		}
	}()

	container := consensusMocks.InitConsensusCore()

	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})
	srBlock := defaultSubroundBlockWithoutErrorFromSubround(sr)

	srBlock.ComputeSubroundProcessingMetric(time.Now(), "dummy")
}

func TestSubroundBlock_ReceivedBlockHeader(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	prevHash := []byte("header hash")
	prevHeader := createDefaultHeader()
	blockchain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderHashCalled: func() []byte {
			return prevHash
		},
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.HeaderV2{
				Header: prevHeader,
			}
		},
	}
	container.SetBlockchain(blockchain)

	// nil header
	sr.ReceivedBlockHeader(nil)

	// header not for current consensus
	sr.ReceivedBlockHeader(&testscommon.HeaderHandlerStub{})

	// nil fields on header
	sr.ReceivedBlockHeader(&testscommon.HeaderHandlerStub{
		CheckFieldsForNilCalled: func() error {
			return expectedErr
		},
	})

	// header not for current consensus
	sr.ReceivedBlockHeader(&testscommon.HeaderHandlerStub{})

	headerForCurrentConsensus := &testscommon.HeaderHandlerStub{
		GetShardIDCalled: func() uint32 {
			return container.ShardCoordinator().SelfId()
		},
		RoundField: uint64(container.RoundHandler().Index()),
		GetPrevHashCalled: func() []byte {
			return prevHash
		},
		GetNonceCalled: func() uint64 {
			return prevHeader.GetNonce() + 1
		},
		GetPrevRandSeedCalled: func() []byte {
			return prevHeader.RandSeed
		},
	}

	// leader
	defaultLeader := sr.Leader()
	sr.SetLeader(sr.SelfPubKey())
	sr.ReceivedBlockHeader(headerForCurrentConsensus)
	sr.SetLeader(defaultLeader)

	// consensus data already set
	sr.SetData([]byte("some data"))
	sr.ReceivedBlockHeader(headerForCurrentConsensus)
	sr.SetData(nil)

	// header leader is not the current one
	sr.SetLeader("X")
	sr.ReceivedBlockHeader(headerForCurrentConsensus)
	sr.SetLeader(defaultLeader)

	// header already received
	sr.SetHeader(&testscommon.HeaderHandlerStub{})
	sr.ReceivedBlockHeader(headerForCurrentConsensus)
	sr.SetHeader(nil)

	// self job already done
	_ = sr.SetJobDone(sr.SelfPubKey(), sr.Current(), true)
	sr.ReceivedBlockHeader(headerForCurrentConsensus)
	_ = sr.SetJobDone(sr.SelfPubKey(), sr.Current(), false)

	// subround already finished
	sr.SetStatus(sr.Current(), spos.SsFinished)
	sr.ReceivedBlockHeader(headerForCurrentConsensus)
	sr.SetStatus(sr.Current(), spos.SsNotFinished)

	// marshal error
	container.SetMarshalizer(&testscommon.MarshallerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, expectedErr
		},
	})
	sr.ReceivedBlockHeader(headerForCurrentConsensus)
	container.SetMarshalizer(&testscommon.MarshallerStub{})

	// should work
	sr.ReceivedBlockHeader(headerForCurrentConsensus)
}

func TestSubroundBlock_GetLeaderForHeader(t *testing.T) {
	t.Parallel()

	t.Run("should fail if not able to compute consensus group", func(t *testing.T) {
		t.Parallel()

		expErr := errors.New("expected error")

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetNodesCoordinator(&shardingMocks.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(randomness []byte, round uint64, shardId, epoch uint32) (leader nodesCoordinator.Validator, validatorsGroup []nodesCoordinator.Validator, err error) {
				return nil, nil, expErr
			},
		})

		leader, err := sr.GetLeaderForHeader(&block.Header{
			Epoch: 10,
		})

		require.Nil(t, leader)
		require.Equal(t, expErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		expLeader := shardingMocks.NewValidatorMock([]byte("pubKey"), 1, 1)

		container.SetNodesCoordinator(&shardingMocks.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(randomness []byte, round uint64, shardId, epoch uint32) (leader nodesCoordinator.Validator, validatorsGroup []nodesCoordinator.Validator, err error) {
				return expLeader, make([]nodesCoordinator.Validator, 0), nil
			},
		})

		leader, err := sr.GetLeaderForHeader(&block.Header{
			Epoch: 10,
		})

		require.Nil(t, err)
		require.Equal(t, expLeader.PubKey(), leader)
	})
}

func TestSubroundBlock_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	sr := initSubroundBlock(nil, container, nil)
	require.True(t, sr.IsInterfaceNil())

	sr = initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	require.False(t, sr.IsInterfaceNil())
}
