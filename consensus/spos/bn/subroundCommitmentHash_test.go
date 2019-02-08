package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func initSubroundCommitmentHash() bn.SubroundCommitmentHash {
	//blockChain := blockchain.BlockChain{}
	//blockProcessorMock := initBlockProcessorMock()
	//bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
	//	return false
	//}}

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	//marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	//shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	//validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

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

func TestWorker_SendCommitmentHash(t *testing.T) {
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

func TestWorker_DoCommitmentHashJobErrCreateCommitmentShouldFail(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	multiSignerMock := initMultiSignerMock()
	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, crypto.ErrNilHasher
	}

	sr.SetMultiSigner(multiSignerMock)

	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	done := sr.DoCommitmentHashJob()

	assert.False(t, done)
}

func TestWorker_ReceivedCommitmentHash(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	dta := []byte("X")

	cnsDta := spos.NewConsensusData(
		dta,
		nil,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		nil,
		int(bn.MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrCommitmentHash); i++ {
		sr.ConsensusState().RoundConsensus.SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	r := sr.ReceivedCommitmentHash(cnsDta)
	assert.False(t, r)

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrCommitmentHash); i++ {
		sr.ConsensusState().RoundConsensus.SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, false)
	}

	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	r = sr.ReceivedCommitmentHash(cnsDta)
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	r = sr.ReceivedCommitmentHash(cnsDta)
	assert.True(t, r)
	isCommHashJobDone, _ := sr.ConsensusState().GetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitmentHash)
	assert.True(t, isCommHashJobDone)
}

func TestWorker_CheckCommitmentHashConsensus(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	sr.ConsensusState().SetSelfPubKey(sr.ConsensusState().ConsensusGroup()[0])

	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok := sr.DoCommitmentHashConsensusCheck()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sr.ConsensusState().Status(bn.SrCommitmentHash))

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrCommitmentHash); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	ok = sr.DoCommitmentHashConsensusCheck()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sr.ConsensusState().Status(bn.SrCommitmentHash))

	sr.ConsensusState().RoundConsensus.SetSelfPubKey("2")

	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok = sr.DoCommitmentHashConsensusCheck()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sr.ConsensusState().Status(bn.SrCommitmentHash))

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBitmap); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	ok = sr.DoCommitmentHashConsensusCheck()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sr.ConsensusState().Status(bn.SrCommitmentHash))

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBitmap); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, false)
	}

	for i := 0; i < len(sr.ConsensusState().RoundConsensus.ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok = sr.DoCommitmentHashConsensusCheck()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sr.ConsensusState().Status(bn.SrCommitmentHash))
}

func TestWorker_IsCommitmentHashReceived(t *testing.T) {
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
	isJobDone, _ := sr.ConsensusState().GetJobDone("A", bn.SrCommitmentHash)
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

func TestWorker_CommitmentHashesCollected(t *testing.T) {
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
	isJobDone, _ := sr.ConsensusState().GetJobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.CommitmentHashesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("B", bn.SrCommitmentHash, true)
	isJobDone, _ = sr.ConsensusState().GetJobDone("B", bn.SrCommitmentHash)
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

func TestWorker_GenCommitmentHashShouldRetunErrOnCreateCommitment(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	multiSignerMock := initMultiSignerMock()
	err := errors.New("error create commitment")
	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, err
	}

	sr.SetMultiSigner(multiSignerMock)

	_, err2 := sr.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnIndexSelfConsensusGroup(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	sr.ConsensusState().SetSelfPubKey("X")

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multiSignerMock.AddCommitmentMock = func(uint16, []byte) error {
		return spos.ErrSelfNotFoundInConsensus
	}

	_, err := sr.GenCommitmentHash()
	assert.Equal(t, spos.ErrSelfNotFoundInConsensus, err)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnAddCommitment(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	err := errors.New("error add commitment")

	multiSignerMock.AddCommitmentMock = func(uint16, []byte) error {
		return err
	}

	sr.SetMultiSigner(multiSignerMock)

	_, err2 := sr.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnSetCommitmentSecret(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multiSignerMock.AddCommitmentMock = func(uint16, []byte) error {
		return nil
	}

	err := errors.New("error set commitment secret")

	multiSignerMock.SetCommitmentSecretMock = func([]byte) error {
		return err
	}

	sr.SetMultiSigner(multiSignerMock)

	_, err2 := sr.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnAddCommitmentHash(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multiSignerMock.AddCommitmentMock = func(uint16, []byte) error {
		return nil
	}

	multiSignerMock.SetCommitmentSecretMock = func([]byte) error {
		return nil
	}

	err := errors.New("error add commitment hash")

	multiSignerMock.AddCommitmentHashMock = func(uint16, []byte) error {
		return err
	}

	sr.SetMultiSigner(multiSignerMock)

	_, err2 := sr.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunNil(t *testing.T) {
	sr := *initSubroundCommitmentHash()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multiSignerMock.AddCommitmentMock = func(uint16, []byte) error {
		return nil
	}

	multiSignerMock.SetCommitmentSecretMock = func([]byte) error {
		return nil
	}

	multiSignerMock.AddCommitmentHashMock = func(uint16, []byte) error {
		return nil
	}

	sr.SetMultiSigner(multiSignerMock)

	_, err := sr.GenCommitmentHash()
	assert.Equal(t, nil, err)
}
