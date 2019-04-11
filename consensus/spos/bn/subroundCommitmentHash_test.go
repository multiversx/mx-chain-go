package bn_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

func initSubroundCommitmentHash() bn.SubroundCommitmentHash {
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
	)

	srCommitmentHash, _ := bn.NewSubroundCommitmentHash(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	return srCommitmentHash
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		nil,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, err, spos.ErrNilSubround)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
	)

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		nil,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
	)

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		consensusState,
		nil,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, err, spos.ErrNilHasher)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilMultisignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
	)

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		consensusState,
		hasherMock,
		nil,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
	)

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		nil,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
	)

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		nil,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, err, spos.ErrNilSyncTimer)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashNilSendConsensusMessageFunctionShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
	)

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		nil,
		extend,
	)

	assert.Nil(t, srCommitmentHash)
	assert.Equal(t, err, spos.ErrNilSendConsensusMessageFunction)
}

func TestSubroundCommitmentHash_NewSubroundCommitmentHashShouldWork(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(COMMITMENT_HASH)",
		ch,
	)

	srCommitmentHash, err := bn.NewSubroundCommitmentHash(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.NotNil(t, srCommitmentHash)
	assert.Nil(t, err)
}

func TestSubroundCommitmentHash_DoCommitmentHashJob(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	sr.ConsensusState().SetSelfPubKey(sr.ConsensusState().ConsensusGroup()[0])

	r := sr.DoCommitmentHashJob()
	assert.True(t, r)

	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	r = sr.DoCommitmentHashJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrCommitmentHash, true)

	r = sr.DoCommitmentHashJob()
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrCommitmentHash, false)
	sr.ConsensusState().Data = nil

	r = sr.DoCommitmentHashJob()
	assert.False(t, r)

	dta := []byte("X")
	sr.ConsensusState().Data = dta

	r = sr.DoCommitmentHashJob()
	assert.True(t, r)
}

func TestSubroundCommitmentHash_ReceivedCommitmentHash(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	commitment := []byte("commitment")

	cnsMsg := spos.NewConsensusMessage(
		sr.ConsensusState().Data,
		commitment,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		nil,
		int(bn.MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().Data = nil
	r := sr.ReceivedCommitmentHash(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().Data = []byte("X")
	cnsMsg.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[0] + "X")
	r = sr.ReceivedCommitmentHash(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[0])
	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	r = sr.ReceivedCommitmentHash(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	sr.ConsensusState().SetSelfPubKey(sr.ConsensusState().ConsensusGroup()[0])
	cnsMsg.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[1])

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		if sr.ConsensusState().ConsensusGroup()[i] != string(cnsMsg.PubKey) {
			sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, true)
		}
	}

	r = sr.ReceivedCommitmentHash(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().SetSelfPubKey(sr.ConsensusState().ConsensusGroup()[1])
	cnsMsg.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[0])
	sr.ConsensusState().ResetRoundState()

	r = sr.ReceivedCommitmentHash(cnsMsg)
	assert.True(t, r)
	isCommHashJobDone, _ := sr.ConsensusState().JobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitmentHash)
	assert.True(t, isCommHashJobDone)
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()
	sr.ConsensusState().RoundCanceled = true
	assert.False(t, sr.DoCommitmentHashConsensusCheck())
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()
	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	assert.True(t, sr.DoCommitmentHashConsensusCheck())
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnTrueWhenIsCommitmentHashReceivedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	assert.True(t, sr.DoCommitmentHashConsensusCheck())
}

func TestSubroundCommitmentHash_DoCommitmentHashConsensusCheckShouldReturnTrueWhenCommitmentHashesCollectedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBitmap); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, true)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, true)
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

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBlock, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitment, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.IsCommitmentHashReceived(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrCommitmentHash, true)
	isJobDone, _ := sr.ConsensusState().JobDone("A", bn.SrCommitmentHash)
	assert.True(t, isJobDone)

	ok = sr.IsCommitmentHashReceived(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("B", bn.SrCommitmentHash, true)
	ok = sr.IsCommitmentHashReceived(2)
	assert.True(t, ok)

	sr.ConsensusState().SetJobDone("C", bn.SrCommitmentHash, true)
	ok = sr.IsCommitmentHashReceived(2)
	assert.True(t, ok)
}

func TestSubroundCommitmentHash_CommitmentHashesCollected(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBlock, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitment, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrBitmap, true)
	sr.ConsensusState().SetJobDone("C", bn.SrBitmap, true)
	isJobDone, _ := sr.ConsensusState().JobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("B", bn.SrCommitmentHash, true)
	isJobDone, _ = sr.ConsensusState().JobDone("B", bn.SrCommitmentHash)
	assert.True(t, isJobDone)

	ok = sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("C", bn.SrCommitmentHash, true)
	ok = sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrCommitmentHash, true)
	ok = sr.CommitmentHashesCollected(2)
	assert.True(t, ok)
}

func TestSubroundCommitmentHash_GenCommitmentHashShouldRetunErrOnIndexSelfConsensusGroup(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	sr.ConsensusState().SetSelfPubKey("X")

	multiSignerMock := initMultiSignerMock()

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

	sr := *initSubroundCommitmentHash()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte) {
		return []byte("commSecret"), []byte("comm")
	}

	err := errors.New("error add commitment")

	multiSignerMock.StoreCommitmentHashMock = func(uint16, []byte) error {
		return err
	}

	sr.SetMultiSigner(multiSignerMock)

	_, err2 := sr.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestSubroundCommitmentHash_GenCommitmentHashShouldRetunNil(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte) {
		return []byte("commSecret"), []byte("comm")
	}

	multiSignerMock.StoreCommitmentHashMock = func(uint16, []byte) error {
		return nil
	}

	sr.SetMultiSigner(multiSignerMock)

	_, err := sr.GenCommitmentHash()
	assert.Equal(t, nil, err)
}

func TestSubroundCommitmentHash_ReceivedCommitmentHashReturnFalseWhenConsensusDataIsNotEqual(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitmentHash()

	cnsMsg := spos.NewConsensusMessage(
		append(sr.ConsensusState().Data, []byte("X")...),
		[]byte("commitment"),
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	assert.False(t, sr.ReceivedCommitmentHash(cnsMsg))
}
