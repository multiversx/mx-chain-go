package bls_test

import (
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type headerWithProof interface {
	GetProof() *block.Proof
}

var expectedErr = errors.New("expected error")

func defaultSubroundForSRBlock(consensusState *spos.ConsensusState, ch chan bool,
	container *mock.ConsensusCoreMock, appStatusHandler core.AppStatusHandler) (*spos.Subround, error) {
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

func defaultSubroundBlockFromSubround(sr *spos.Subround) (bls.SubroundBlock, error) {
	srBlock, err := bls.NewSubroundBlock(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
	)

	return srBlock, err
}

func defaultSubroundBlockWithoutErrorFromSubround(sr *spos.Subround) bls.SubroundBlock {
	srBlock, _ := bls.NewSubroundBlock(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
	)

	return srBlock
}

func initSubroundBlock(
	blockChain data.ChainHandler,
	container *mock.ConsensusCoreMock,
	appStatusHandler core.AppStatusHandler,
) bls.SubroundBlock {
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

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	container.SetBlockchain(blockChain)

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, appStatusHandler)
	srBlock, _ := defaultSubroundBlockFromSubround(sr)
	return srBlock
}

func createConsensusContainers() []*mock.ConsensusCoreMock {
	consensusContainers := make([]*mock.ConsensusCoreMock, 0)
	container := mock.InitConsensusCore()
	consensusContainers = append(consensusContainers, container)
	container = mock.InitConsensusCoreHeaderV2()
	consensusContainers = append(consensusContainers, container)
	return consensusContainers
}

func initSubroundBlockWithBlockProcessor(
	bp *testscommon.BlockProcessorStub,
	container *mock.ConsensusCoreMock,
) bls.SubroundBlock {
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
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})
	srBlock, _ := defaultSubroundBlockFromSubround(sr)
	return srBlock
}

func TestSubroundBlock_NewSubroundBlockNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srBlock, err := bls.NewSubroundBlock(
		nil,
		extend,
		bls.ProcessingThresholdPercent,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundBlock_NewSubroundBlockNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetBlockchain(nil)

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundBlock_NewSubroundBlockNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetBlockProcessor(nil)

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubroundBlock_NewSubroundBlockNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	sr.ConsensusState = nil

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundBlock_NewSubroundBlockNilHasherShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetHasher(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubroundBlock_NewSubroundBlockNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetMarshalizer(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestSubroundBlock_NewSubroundBlockNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetMultiSignerContainer(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestSubroundBlock_NewSubroundBlockNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetRoundHandler(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubroundBlock_NewSubroundBlockNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetShardCoordinator(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestSubroundBlock_NewSubroundBlockNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})

	container.SetSyncTimer(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundBlock_NewSubroundBlockShouldWork(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
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
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("round index lower than last committed block should return false", func(t *testing.T) {
		t.Parallel()
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		sr.SetSelfPubKey(sr.ConsensusGroup()[0])
		_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrBlock, true)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("leader job done should return false", func(t *testing.T) {
		t.Parallel()
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		sr.SetSelfPubKey(sr.ConsensusGroup()[0])
		_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrBlock, true)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("subround finished should return false", func(t *testing.T) {
		t.Parallel()
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		sr.SetSelfPubKey(sr.ConsensusGroup()[0])
		_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrBlock, false)
		sr.SetStatus(bls.SrBlock, spos.SsFinished)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("create header error should return false", func(t *testing.T) {
		t.Parallel()
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		sr.SetSelfPubKey(sr.ConsensusGroup()[0])
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
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})
		sr.SetSelfPubKey(sr.ConsensusGroup()[0])
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
	t.Run("send block error should return false", func(t *testing.T) {
		t.Parallel()
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})

		sr.SetSelfPubKey(sr.ConsensusGroup()[0])
		bpm := mock.InitBlockProcessorMock(container.Marshalizer())
		container.SetBlockProcessor(bpm)
		bm := &mock.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return expectedErr
			},
		}
		container.SetBroadcastMessenger(bm)
		r := sr.DoBlockJob()
		assert.False(t, r)
	})
	t.Run("should work, consensus propagation changes flag enabled", func(t *testing.T) {
		t.Parallel()
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

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
		container.SetEnableEpochsHandler(enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ConsensusPropagationChangesFlag))

		providedSignature := []byte("provided signature")
		providedBitmap := []byte("provided bitmap")
		providedHeadr := &block.HeaderV2{
			Header: &block.Header{
				Signature:     providedSignature,
				PubKeysBitmap: providedBitmap,
			},
		}
		container.SetBlockchain(&testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return providedHeadr
			},
		})

		sr.SetSelfPubKey(sr.ConsensusGroup()[0])
		bpm := mock.InitBlockProcessorMock(container.Marshalizer())
		container.SetBlockProcessor(bpm)
		bpm.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return &block.HeaderV2{
				Header: &block.Header{
					Round: round,
					Nonce: nonce,
				},
			}, nil
		}
		bm := &mock.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return nil
			},
		}
		container.SetBroadcastMessenger(bm)
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		r := sr.DoBlockJob()
		assert.True(t, r)
		assert.Equal(t, uint64(1), sr.Header.GetNonce())

		expectedProof := &block.Proof{
			PreviousPubkeysBitmap:       providedBitmap,
			PreviousAggregatedSignature: providedSignature,
		}
		hdrWithProof, ok := sr.Header.(headerWithProof)
		assert.True(t, ok)
		assert.Equal(t, expectedProof, hdrWithProof.GetProof())
	})
	t.Run("should work, consensus propagation changes flag not enabled", func(t *testing.T) {
		t.Parallel()
		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

		container.SetRoundHandler(&testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 1
			},
		})

		sr.SetSelfPubKey(sr.ConsensusGroup()[0])
		bpm := mock.InitBlockProcessorMock(container.Marshalizer())
		container.SetBlockProcessor(bpm)
		bm := &mock.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return nil
			},
		}
		container.SetBroadcastMessenger(bm)
		container.SetRoundHandler(&mock.RoundHandlerMock{
			RoundIndex: 1,
		})
		r := sr.DoBlockJob()
		assert.True(t, r)
		assert.Equal(t, uint64(1), sr.Header.GetNonce())
	})

}

func TestSubroundBlock_ReceivedBlockBodyAndHeaderDataAlreadySet(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	hdr := &block.Header{Nonce: 1}
	blkBody := &block.Body{}

	cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)

	sr.Data = []byte("some data")
	r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
	assert.False(t, r)
}

func TestSubroundBlock_ReceivedBlockBodyAndHeaderNodeNotLeaderInCurrentRound(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	hdr := &block.Header{Nonce: 1}
	blkBody := &block.Body{}

	cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[1]), bls.MtBlockBodyAndHeader)

	sr.Data = nil
	r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
	assert.False(t, r)
}

func TestSubroundBlock_ReceivedBlockBodyAndHeaderCannotProcessJobDone(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	hdr := &block.Header{Nonce: 1}
	blkBody := &block.Body{}

	cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)

	sr.Data = nil
	_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrBlock, true)
	r := sr.ReceivedBlockBodyAndHeader(cnsMsg)

	assert.False(t, r)
}

func TestSubroundBlock_ReceivedBlockBodyAndHeaderErrorDecoding(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	blProc := mock.InitBlockProcessorMock(container.Marshalizer())
	blProc.DecodeBlockHeaderCalled = func(dta []byte) data.HeaderHandler {
		// error decoding so return nil
		return nil
	}
	container.SetBlockProcessor(blProc)

	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	hdr := &block.Header{Nonce: 1}
	blkBody := &block.Body{}

	cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)

	sr.Data = nil
	r := sr.ReceivedBlockBodyAndHeader(cnsMsg)

	assert.False(t, r)
}

func TestSubroundBlock_ReceivedBlockBodyAndHeaderBodyAlreadyReceived(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	hdr := &block.Header{Nonce: 1}
	blkBody := &block.Body{}

	cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)

	sr.Data = nil
	sr.Body = &block.Body{}
	r := sr.ReceivedBlockBodyAndHeader(cnsMsg)

	assert.False(t, r)
}

func TestSubroundBlock_ReceivedBlockBodyAndHeaderHeaderAlreadyReceived(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})

	hdr := &block.Header{Nonce: 1}
	blkBody := &block.Body{}

	cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)

	sr.Data = nil
	sr.Header = &block.Header{Nonce: 1}
	r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
	assert.False(t, r)
}

func TestSubroundBlock_ReceivedBlockBodyAndHeaderOK(t *testing.T) {
	t.Parallel()

	t.Run("block is valid", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		hdr := createDefaultHeader()
		blkBody := &block.Body{}
		cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)
		sr.Data = nil
		r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
		assert.True(t, r)
	})
	t.Run("block is not valid", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		hdr := &block.Header{
			Nonce: 1,
		}
		blkBody := &block.Body{}
		cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)
		sr.Data = nil
		r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
		assert.False(t, r)
	})
	t.Run("header with proof before flag activation should error", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		container.SetBlockProcessor(&testscommon.BlockProcessorStub{
			DecodeBlockHeaderCalled: func(dta []byte) data.HeaderHandler {
				hdr := &block.HeaderV2{}
				_ = container.Marshalizer().Unmarshal(hdr, dta)
				return hdr
			},
		})
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		blkBody := &block.Body{}
		hdr := &block.HeaderV2{
			Header: &block.Header{},
			Proof:  &block.Proof{},
		}
		cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)
		sr.Data = nil
		r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
		assert.False(t, r)
	})
	t.Run("header without proof after flag activation should error", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		container.SetBlockProcessor(&testscommon.BlockProcessorStub{
			DecodeBlockHeaderCalled: func(dta []byte) data.HeaderHandler {
				hdr := &block.HeaderV2{}
				_ = container.Marshalizer().Unmarshal(hdr, dta)
				return hdr
			},
		})
		container.SetEnableEpochsHandler(enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ConsensusPropagationChangesFlag))
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		blkBody := &block.Body{}
		hdr := &block.HeaderV2{
			Header: &block.Header{},
			Proof:  nil,
		}
		cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)
		sr.Data = nil
		r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
		assert.False(t, r)
	})
	t.Run("header with leader sig after flag activation should error", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		container.SetBlockProcessor(&testscommon.BlockProcessorStub{
			DecodeBlockHeaderCalled: func(dta []byte) data.HeaderHandler {
				hdr := &block.HeaderV2{}
				_ = container.Marshalizer().Unmarshal(hdr, dta)
				return hdr
			},
		})
		container.SetEnableEpochsHandler(enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ConsensusPropagationChangesFlag))
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		blkBody := &block.Body{}
		hdr := &block.HeaderV2{
			Header: &block.Header{
				LeaderSignature: []byte("leader signature"),
			},
			Proof: &block.Proof{},
		}
		cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)
		sr.Data = nil
		r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
		assert.False(t, r)
	})
	t.Run("header with proof after flag activation should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		blockProcessor := mock.InitBlockProcessorHeaderV2Mock()
		blockProcessor.DecodeBlockHeaderCalled = func(dta []byte) data.HeaderHandler {
			hdr := &block.HeaderV2{}
			_ = container.Marshalizer().Unmarshal(hdr, dta)
			return hdr
		}
		container.SetBlockProcessor(blockProcessor)
		container.SetEnableEpochsHandler(enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ConsensusPropagationChangesFlag))
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		blkBody := &block.Body{}
		hdr := &block.HeaderV2{
			Header:                   createDefaultHeader(),
			ScheduledDeveloperFees:   big.NewInt(1),
			ScheduledAccumulatedFees: big.NewInt(1),
			ScheduledRootHash:        []byte("scheduled root hash"),
			Proof:                    &block.Proof{},
		}
		cnsMsg := createConsensusMessage(hdr, blkBody, []byte(sr.ConsensusGroup()[0]), bls.MtBlockBodyAndHeader)
		sr.Data = nil
		r := sr.ReceivedBlockBodyAndHeader(cnsMsg)
		assert.True(t, r)
	})
}

func createConsensusMessage(header data.HeaderHandler, body *block.Body, leader []byte, topic consensus.MessageType) *consensus.Message {
	marshaller := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	hdrStr, _ := marshaller.Marshal(header)
	hdrHash := hasher.Compute(string(hdrStr))
	blkBodyStr, _ := marshaller.Marshal(body)

	return consensus.NewConsensusMessage(
		hdrHash,
		nil,
		blkBodyStr,
		hdrStr,
		leader,
		[]byte("sig"),
		int(topic),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)
}

func TestSubroundBlock_ReceivedBlock(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	blockProcessorMock := mock.InitBlockProcessorMock(container.Marshalizer())
	blkBody := &block.Body{}
	blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		blkBodyStr,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
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
	sr.Body = &block.Body{}
	r := sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.Body = nil
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

	hdr := createDefaultHeader()
	hdr.Nonce = 2
	hdrStr, _ := container.Marshalizer().Marshal(hdr)
	hdrHash := (&hashingMocks.HasherMock{}).Compute(string(hdrStr))
	cnsMsg = consensus.NewConsensusMessage(
		hdrHash,
		nil,
		nil,
		hdrStr,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bls.MtBlockHeader),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.Data = nil
	sr.Header = hdr
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.Header = nil
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0])
	sr.SetStatus(bls.SrBlock, spos.SsFinished)
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(bls.SrBlock, spos.SsNotFinished)
	container.SetBlockProcessor(blockProcessorMock)
	sr.Data = nil
	sr.Header = nil
	hdr = createDefaultHeader()
	hdr.Nonce = 1
	hdrStr, _ = mock.MarshalizerMock{}.Marshal(hdr)
	hdrHash = (&hashingMocks.HasherMock{}).Compute(string(hdrStr))
	cnsMsg.BlockHeaderHash = hdrHash
	cnsMsg.Header = hdrStr
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.True(t, r)
}

func TestSubroundBlock_ReceivedBlockShouldWorkWithPropagationChangesFlagEnabled(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	blockProcessorMock := mock.InitBlockProcessorMock(container.Marshalizer())

	container.SetEnableEpochsHandler(enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ConsensusPropagationChangesFlag))

	providedLeaderSignature := []byte("leader signature")
	wasVerifySingleSignatureCalled := false
	wasStoreSignatureShareCalled := false
	container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
		VerifySingleSignatureCalled: func(publicKeyBytes []byte, message []byte, signature []byte) error {
			assert.Equal(t, providedLeaderSignature, signature)
			wasVerifySingleSignatureCalled = true
			return nil
		},
		StoreSignatureShareCalled: func(index uint16, sig []byte) error {
			assert.Equal(t, providedLeaderSignature, sig)
			wasStoreSignatureShareCalled = true
			return nil

		},
	})

	hdr := createDefaultHeader()
	hdr.Nonce = 2
	hdrStr, _ := container.Marshalizer().Marshal(hdr)
	hdrHash := (&hashingMocks.HasherMock{}).Compute(string(hdrStr))
	cnsMsg := consensus.NewConsensusMessage(
		hdrHash,
		providedLeaderSignature,
		nil,
		hdrStr,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bls.MtBlockHeader),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
	)

	sr.SetStatus(bls.SrBlock, spos.SsNotFinished)
	container.SetBlockProcessor(blockProcessorMock)
	sr.Data = nil
	sr.Body = &block.Body{}
	r := sr.ReceivedBlockHeader(cnsMsg)
	assert.True(t, r)
	assert.True(t, wasStoreSignatureShareCalled)
	assert.True(t, wasVerifySingleSignatureCalled)
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		nil,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
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
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	blProcMock := mock.InitBlockProcessorMock(container.Marshalizer())
	err := errors.New("error process block")
	blProcMock.ProcessBlockCalled = func(data.HeaderHandler, data.BodyHandler, func() time.Duration) error {
		return err
	}
	container.SetBlockProcessor(blProcMock)
	hdr := &block.Header{}
	blkBody := &block.Body{}
	blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		blkBodyStr,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
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
	sr.Header = hdr
	sr.Body = blkBody
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	hdr := &block.Header{}
	blkBody := &block.Body{}
	blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		blkBodyStr,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
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
	sr.Header = hdr
	sr.Body = blkBody
	blockProcessorMock := mock.InitBlockProcessorMock(container.Marshalizer())
	blockProcessorMock.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return expectedErr
	}
	container.SetBlockProcessor(blockProcessorMock)
	container.SetRoundHandler(&mock.RoundHandlerMock{RoundIndex: 1})
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnTrue(t *testing.T) {
	t.Parallel()

	consensusContainers := createConsensusContainers()
	for _, container := range consensusContainers {
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		hdr, _ := container.BlockProcessor().CreateNewHeader(1, 1)
		hdr, blkBody, _ := container.BlockProcessor().CreateBlock(hdr, func() bool { return true })

		blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)
		cnsMsg := consensus.NewConsensusMessage(
			nil,
			nil,
			blkBodyStr,
			nil,
			[]byte(sr.ConsensusGroup()[0]),
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
		sr.Header = hdr
		sr.Body = blkBody
		assert.True(t, sr.ProcessReceivedBlock(cnsMsg))
	}
}

func TestSubroundBlock_RemainingTimeShouldReturnNegativeValue(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	roundHandlerMock := initRoundHandlerMock()
	container.SetRoundHandler(roundHandlerMock)

	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	remainingTimeInThisRound := func() time.Duration {
		roundStartTime := sr.RoundHandler().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.RoundHandler().TimeDuration()*85/100 - elapsedTime

		return remainingTime
	}
	container.SetSyncTimer(&mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 84 / 100)
	}})
	ret := remainingTimeInThisRound()
	assert.True(t, ret > 0)

	container.SetSyncTimer(&mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 85 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret == 0)

	container.SetSyncTimer(&mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 86 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret < 0)
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	sr.RoundCanceled = true
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	sr.SetStatus(bls.SrBlock, spos.SsFinished)
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenBlockIsReceivedReturnTrue(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	for i := 0; i < sr.Threshold(bls.SrBlock); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrBlock, true)
	}
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenBlockIsReceivedReturnFalse(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_IsBlockReceived(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
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
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	haveTimeInCurrentSubound := func() bool {
		roundStartTime := sr.RoundHandler().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.EndTime() - int64(elapsedTime)

		return time.Duration(remainingTime) > 0
	}
	roundHandlerMock := &mock.RoundHandlerMock{}
	roundHandlerMock.TimeDurationCalled = func() time.Duration {
		return 4000 * time.Millisecond
	}
	roundHandlerMock.TimeStampCalled = func() time.Time {
		return time.Unix(0, 0)
	}
	syncTimerMock := &mock.SyncTimerMock{}
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
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
	haveTimeInCurrentSubound := func() bool {
		roundStartTime := sr.RoundHandler().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.EndTime() - int64(elapsedTime)

		return time.Duration(remainingTime) > 0
	}
	roundHandlerMock := &mock.RoundHandlerMock{}
	roundHandlerMock.TimeDurationCalled = func() time.Duration {
		return 4000 * time.Millisecond
	}
	roundHandlerMock.TimeStampCalled = func() time.Time {
		return time.Unix(0, 0)
	}
	syncTimerMock := &mock.SyncTimerMock{}
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
		sr := *initSubroundBlock(blockChain, container, &statusHandler.AppStatusHandlerStub{})
		_ = sr.BlockChain().SetCurrentBlockHeaderAndRootHash(nil, nil)
		header, _ := sr.CreateHeader()
		header, body, _ := sr.CreateBlock(header)
		marshalizedBody, _ := sr.Marshalizer().Marshal(body)
		marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
		_ = sr.SendBlockBody(body, marshalizedBody)
		_ = sr.SendBlockHeader(header, marshalizedHeader, nil)

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
		sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{})
		_ = sr.BlockChain().SetCurrentBlockHeaderAndRootHash(&block.Header{
			Nonce: 1,
		}, []byte("root hash"))

		header, _ := sr.CreateHeader()
		header, body, _ := sr.CreateBlock(header)
		marshalizedBody, _ := sr.Marshalizer().Marshal(body)
		marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
		_ = sr.SendBlockBody(body, marshalizedBody)
		_ = sr.SendBlockHeader(header, marshalizedHeader, nil)

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
	container := mock.InitConsensusCore()
	bp := mock.InitBlockProcessorMock(container.Marshalizer())
	bp.CreateBlockCalled = func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		shardHeader, _ := header.(*block.Header)
		shardHeader.MiniBlockHeaders = mbHeaders
		shardHeader.RootHash = []byte{}

		return shardHeader, &block.Body{}, nil
	}
	sr := *initSubroundBlockWithBlockProcessor(bp, container)
	container.SetBlockchain(&blockChainMock)

	header, _ := sr.CreateHeader()
	header, body, _ := sr.CreateBlock(header)
	marshalizedBody, _ := sr.Marshalizer().Marshal(body)
	marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
	_ = sr.SendBlockBody(body, marshalizedBody)
	_ = sr.SendBlockHeader(header, marshalizedHeader, nil)

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
	container := mock.InitConsensusCore()
	bp := mock.InitBlockProcessorMock(container.Marshalizer())
	bp.CreateBlockCalled = func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		return nil, nil, expectedErr
	}
	sr := *initSubroundBlockWithBlockProcessor(bp, container)
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

	container := mock.InitConsensusCore()
	receivedValue := uint64(0)
	container.SetBlockProcessor(&testscommon.BlockProcessorStub{
		ProcessBlockCalled: func(_ data.HeaderHandler, _ data.BodyHandler, _ func() time.Duration) error {
			time.Sleep(time.Duration(delay))
			return nil
		},
	})
	sr := *initSubroundBlock(nil, container, &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			receivedValue = value
		}})
	hdr := &block.Header{}
	blkBody := &block.Body{}
	blkBodyStr, _ := mock.MarshalizerMock{}.Marshal(blkBody)

	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		blkBodyStr,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
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
	sr.Header = hdr
	sr.Body = blkBody

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

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container, &statusHandler.AppStatusHandlerStub{})
	srBlock := *defaultSubroundBlockWithoutErrorFromSubround(sr)

	srBlock.ComputeSubroundProcessingMetric(time.Now(), "dummy")
}
