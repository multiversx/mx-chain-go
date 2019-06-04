package bn_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func initSubroundCommitmentHashWithContainer(container *mock.ConsensusCoreMock) bn.SubroundCommitmentHash {
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)

	srCommitmentHash, _ := bn.NewSubroundCommitmentHash(
		sr,
		extend,
	)

	return srCommitmentHash
}

func initSubroundCommitmentHash() bn.SubroundCommitmentHash {
	container := mock.InitConsensusCore()

	return initSubroundCommitmentHashWithContainer(container)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		nil,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	sr.ConsensusState = nil
	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	container.SetHasher(nil)
	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilMultisignerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	container.SetMultiSigner(nil)
	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	container.SetRounder(nil)
	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	container.SetSyncTimer(nil)
	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		extend,
	)

	assert.NotNil(t, srCommitmentHash)
	assert.Nil(t, err)
}

func TestSubroundCommitmentHash_DoCommitmentHashJob(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	r := sr.DoCommitmentHashJob()
	assert.True(t, r)

	sr.SetStatus(bn.SrBlock, spos.SsFinished)
	sr.SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	r = sr.DoCommitmentHashJob()
	assert.False(t, r)

	sr.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	sr.SetJobDone(sr.SelfPubKey(), bn.SrCommitmentHash, true)

	r = sr.DoCommitmentHashJob()
	assert.False(t, r)

	sr.SetJobDone(sr.SelfPubKey(), bn.SrCommitmentHash, false)
	sr.Data = nil

	r = sr.DoCommitmentHashJob()
	assert.False(t, r)

	dta := []byte("X")
	sr.Data = dta

	r = sr.DoCommitmentHashJob()
	assert.True(t, r)
}

func TestSubroundCommitmentHash_ReceivedCommitmentHash(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	commitment := []byte("commitment")
	cnsMsg := consensus.NewConsensusMessage(
		sr.Data,
		commitment,
		[]byte(sr.ConsensusGroup()[0]),
		nil,
		int(bn.MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.Data = nil
	r := sr.ReceivedCommitmentHash(cnsMsg)
	assert.False(t, r)

	sr.Data = []byte("X")
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0] + "X")
	r = sr.ReceivedCommitmentHash(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0])
	sr.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	r = sr.ReceivedCommitmentHash(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	sr.SetSelfPubKey(sr.ConsensusGroup()[0])
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		if sr.ConsensusGroup()[i] != string(cnsMsg.PubKey) {
			sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitmentHash, true)
		}
	}

	r = sr.ReceivedCommitmentHash(cnsMsg)
	assert.False(t, r)

	sr.SetSelfPubKey(sr.ConsensusGroup()[1])
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0])
	sr.ResetRoundState()

	r = sr.ReceivedCommitmentHash(cnsMsg)
	assert.True(t, r)
	isCommHashJobDone, _ := sr.JobDone(sr.ConsensusGroup()[0], bn.SrCommitmentHash)
	assert.True(t, isCommHashJobDone)
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()
	sr.RoundCanceled = true
	assert.False(t, sr.DoCommitmentHashConsensusCheck())
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()
	sr.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	assert.True(t, sr.DoCommitmentHashConsensusCheck())
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnTrueWhenIsCommitmentHashReceivedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	assert.True(t, sr.DoCommitmentHashConsensusCheck())
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnTrueWhenCommitmentHashesCollectedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	for i := 0; i < sr.Threshold(bn.SrBitmap); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitmentHash, true)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	assert.True(t, sr.DoCommitmentHashConsensusCheck())
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()
	assert.False(t, sr.DoCommitmentHashConsensusCheck())
}

func TestSubroundCommitmentHash_IsCommitmentHashReceived(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBlock, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBitmap, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitment, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.IsCommitmentHashReceived(2)
	assert.False(t, ok)

	sr.SetJobDone("A", bn.SrCommitmentHash, true)
	isJobDone, _ := sr.JobDone("A", bn.SrCommitmentHash)
	assert.True(t, isJobDone)

	ok = sr.IsCommitmentHashReceived(2)
	assert.False(t, ok)

	sr.SetJobDone("B", bn.SrCommitmentHash, true)
	ok = sr.IsCommitmentHashReceived(2)
	assert.True(t, ok)

	sr.SetJobDone("C", bn.SrCommitmentHash, true)
	ok = sr.IsCommitmentHashReceived(2)
	assert.True(t, ok)
}

func TestSubroundCommitmentHash_CommitmentHashesCollected(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBlock, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBitmap, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitment, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("A", bn.SrBitmap, true)
	sr.SetJobDone("C", bn.SrBitmap, true)
	isJobDone, _ := sr.JobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("B", bn.SrCommitmentHash, true)
	isJobDone, _ = sr.JobDone("B", bn.SrCommitmentHash)
	assert.True(t, isJobDone)

	ok = sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("C", bn.SrCommitmentHash, true)
	ok = sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("A", bn.SrCommitmentHash, true)
	ok = sr.CommitmentHashesCollected(2)
	assert.True(t, ok)
}

func TestSubroundCommitmentHash_GenCommitmentHashShouldRetunErrOnIndexSelfConsensusGroup(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	sr.SetSelfPubKey("X")

	multiSignerMock := mock.InitMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte) {
		return []byte("commSecret"), []byte("comm")
	}

	multiSignerMock.StoreCommitmentMock = func(uint16, []byte) error {
		return spos.ErrNotFoundInConsensus
	}

	_, err := sr.GenCommitmentHash()
	assert.Equal(t, spos.ErrNotFoundInConsensus, err)
}

func TestSubroundCommitmentHash_GenCommitmentHashShouldRetunErrOnAddCommitment(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundCommitmentHashWithContainer(container)

	multiSignerMock := mock.InitMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte) {
		return []byte("commSecret"), []byte("comm")
	}

	err := errors.New("error add commitment")

	multiSignerMock.StoreCommitmentHashMock = func(uint16, []byte) error {
		return err
	}

	container.SetMultiSigner(multiSignerMock)

	_, err2 := sr.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestSubroundCommitmentHash_GenCommitmentHashShouldRetunNil(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundCommitmentHashWithContainer(container)

	multiSignerMock := mock.InitMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte) {
		return []byte("commSecret"), []byte("comm")
	}

	multiSignerMock.StoreCommitmentHashMock = func(uint16, []byte) error {
		return nil
	}

	container.SetMultiSigner(multiSignerMock)

	_, err := sr.GenCommitmentHash()
	assert.Equal(t, nil, err)
}

func TestSubroundCommitmentHash_ReceivedCommitmentHashReturnFalseWhenConsensusDataIsNotEqual(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	cnsMsg := consensus.NewConsensusMessage(
		append(sr.Data, []byte("X")...),
		[]byte("commitment"),
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	assert.False(t, sr.ReceivedCommitmentHash(cnsMsg))
}
