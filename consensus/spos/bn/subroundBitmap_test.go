package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
)

func initSubroundBitmap() bn.SubroundBitmap {
	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bn.SrCommitmentHash,
		bn.SrBitmap,
		bn.SrCommitment,
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
	)

	srBitmap, _ := bn.NewSubroundBitmap(
		sr,
		extend,
	)

	return srBitmap
}

func TestSubroundBitmap_NewSubroundBitmapNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srBitmap, err := bn.NewSubroundBitmap(
		nil,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundBitmap_NewSubroundBitmapNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bn.SrCommitmentHash,
		bn.SrBitmap,
		bn.SrCommitment,
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
	)

	container.SetBlockProcessor(nil)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubroundBitmap_NewSubroundBitmapNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()
	consensusState := initConsensusState()

	sr, _ := spos.NewSubround(
		bn.SrCommitmentHash,
		bn.SrBitmap,
		bn.SrCommitment,
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
	)

	sr.ConsensusState = nil
	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundBitmap_NewSubroundBitmapNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bn.SrCommitmentHash,
		bn.SrBitmap,
		bn.SrCommitment,
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
	)

	container.SetRounder(nil)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubroundBitmap_NewSubroundBitmapNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bn.SrCommitmentHash,
		bn.SrBitmap,
		bn.SrCommitment,
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
	)

	container.SetSyncTimer(nil)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundBitmap_NewSubroundBitmapShouldWork(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	container := mock.InitConsensusCore()

	sr, _ := spos.NewSubround(
		bn.SrCommitmentHash,
		bn.SrBitmap,
		bn.SrCommitment,
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
	)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		extend,
	)

	assert.NotNil(t, srBitmap)
	assert.Nil(t, err)
}

func TestSubroundBitmap_DoBitmapJob(t *testing.T) {
	t.Parallel()

	sr := *initSubroundBitmap()

	sr.Header = &block.Header{}

	r := sr.DoBitmapJob()
	assert.False(t, r)

	sr.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	sr.SetStatus(bn.SrBitmap, spos.SsFinished)

	r = sr.DoBitmapJob()
	assert.False(t, r)

	sr.SetStatus(bn.SrBitmap, spos.SsNotFinished)
	_ = sr.SetJobDone(sr.SelfPubKey(), bn.SrBitmap, true)

	r = sr.DoBitmapJob()
	assert.False(t, r)

	_ = sr.SetJobDone(sr.SelfPubKey(), bn.SrBitmap, false)
	sr.SetSelfPubKey(sr.ConsensusGroup()[1])

	r = sr.DoBitmapJob()
	assert.False(t, r)

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])
	sr.Data = nil

	r = sr.DoBitmapJob()
	assert.False(t, r)

	dta := []byte("X")
	sr.Data = dta
	_ = sr.SetJobDone(sr.SelfPubKey(), bn.SrCommitmentHash, true)

	r = sr.DoBitmapJob()
	assert.True(t, r)
	isBitmapJobDone, _ := sr.JobDone(sr.SelfPubKey(), bn.SrBitmap)
	assert.True(t, isBitmapJobDone)
}

func TestSubroundBitmap_ReceivedBitmap(t *testing.T) {
	t.Parallel()

	sr := *initSubroundBitmap()

	sr.Header = &block.Header{}

	commitment := []byte("commitment")
	cnsMsg := consensus.NewConsensusMessage(
		sr.Data,
		commitment,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		0,
		chainID,
	)

	sr.Data = nil
	r := sr.ReceivedBitmap(cnsMsg)
	assert.False(t, r)

	sr.Data = []byte("X")
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0] + "X")
	r = sr.ReceivedBitmap(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0])
	sr.SetStatus(bn.SrBitmap, spos.SsFinished)
	r = sr.ReceivedBitmap(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	bitmap := make([]byte, 3)

	cnGroup := sr.ConsensusGroup()

	selfIndexInConsensusGroup, _ := sr.SelfConsensusGroupIndex()

	// fill ony few of the signers in bitmap
	for i := 0; i < 5; i++ {
		if i != selfIndexInConsensusGroup {
			bitmap[i/8] |= 1 << uint16(i%8)
		}
	}

	cnsMsg.SubRoundData = bitmap

	r = sr.ReceivedBitmap(cnsMsg)
	assert.False(t, r)

	//fill the rest except self
	for i := 5; i < len(cnGroup); i++ {
		if i != selfIndexInConsensusGroup {
			bitmap[i/8] |= 1 << uint16(i%8)
		}
	}

	cnsMsg.SubRoundData = bitmap
	r = sr.ReceivedBitmap(cnsMsg)
	assert.False(t, r)

	//fill self
	sr.ResetRoundState()
	bitmap[selfIndexInConsensusGroup/8] |= 1 << uint16(selfIndexInConsensusGroup%8)

	cnsMsg.SubRoundData = bitmap
	r = sr.ReceivedBitmap(cnsMsg)
	assert.True(t, r)
}

func TestSubroundBitmap_DoBitmapConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundBitmap()
	sr.RoundCanceled = true
	assert.False(t, sr.DoBitmapConsensusCheck())
}

func TestSubroundBitmap_DoBitmapConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundBitmap()
	sr.SetStatus(bn.SrBitmap, spos.SsFinished)
	assert.True(t, sr.DoBitmapConsensusCheck())
}

func TestSubroundBitmap_DoBitmapConsensusCheckShouldReturnTrueWhenBitmapIsReceivedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundBitmap()

	for i := 0; i < sr.Threshold(bn.SrBitmap); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	assert.True(t, sr.DoBitmapConsensusCheck())
}

func TestSubroundBitmap_DoBitmapConsensusCheckShouldReturnFalseWhenBitmapIsReceivedReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundBitmap()
	assert.False(t, sr.DoBitmapConsensusCheck())
}

func TestSubroundBitmap_IsBitmapReceived(t *testing.T) {
	t.Parallel()

	sr := *initSubroundBitmap()

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBlock, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBitmap, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitment, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.IsBitmapReceived(2)
	assert.False(t, ok)

	_ = sr.SetJobDone("A", bn.SrBitmap, true)
	isJobDone, _ := sr.JobDone("A", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.IsBitmapReceived(2)
	assert.False(t, ok)

	_ = sr.SetJobDone("B", bn.SrBitmap, true)
	ok = sr.IsBitmapReceived(2)
	assert.True(t, ok)

	_ = sr.SetJobDone("C", bn.SrBitmap, true)
	ok = sr.IsBitmapReceived(2)
	assert.True(t, ok)
}

func TestSubroundBitmap_ReceivedBitmapReturnFalseWhenConsensusDataIsNotEqual(t *testing.T) {
	t.Parallel()

	sr := *initSubroundBitmap()

	sr.Header = &block.Header{}

	cnsMsg := consensus.NewConsensusMessage(
		append(sr.Data, []byte("X")...),
		[]byte("commitment"),
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		0,
		chainID,
	)

	assert.False(t, sr.ReceivedBitmap(cnsMsg))
}
