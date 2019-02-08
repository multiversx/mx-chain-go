package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWorker_SendCommitmentHash(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetSelfPubKey(wrk.SPoS.ConsensusGroup()[0])

	r := wrk.DoCommitmentHashJob()
	assert.True(t, r)

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	r = wrk.DoCommitmentHashJob()
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrCommitmentHash, true)

	r = wrk.DoCommitmentHashJob()
	assert.False(t, r)

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrCommitmentHash, false)
	wrk.SPoS.Data = nil

	r = wrk.DoCommitmentHashJob()
	assert.False(t, r)

	dta := []byte("X")
	wrk.SPoS.Data = dta

	r = wrk.DoCommitmentHashJob()
	assert.True(t, r)
}

func TestWorker_DoCommitmentHashJobErrCreateCommitmentShouldFail(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()
	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, crypto.ErrNilHasher
	}

	wrk.SetMultiSigner(multiSignerMock)

	wrk.SendMessage = sendMessage
	wrk.BroadcastBlockBody = broadcastMessage
	wrk.BroadcastHeader = broadcastMessage

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	done := wrk.DoCommitmentHashJob()

	assert.False(t, done)
}

func TestWorker_ReceivedCommitmentHash(t *testing.T) {
	wrk := initWorker()

	dta := []byte("X")

	cnsDta := spos.NewConsensusData(
		dta,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		nil,
		int(bn.MtCommitmentHash),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	for i := 0; i < wrk.SPoS.Threshold(bn.SrCommitmentHash); i++ {
		wrk.SPoS.RoundConsensus.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	r := wrk.ReceivedCommitmentHash(cnsDta)
	assert.False(t, r)

	for i := 0; i < wrk.SPoS.Threshold(bn.SrCommitmentHash); i++ {
		wrk.SPoS.RoundConsensus.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
	}

	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	r = wrk.ReceivedCommitmentHash(cnsDta)
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	r = wrk.ReceivedCommitmentHash(cnsDta)
	assert.True(t, r)
	isCommHashJobDone, _ := wrk.SPoS.GetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrCommitmentHash)
	assert.True(t, isCommHashJobDone)
}

func TestWorker_CheckCommitmentHashConsensus(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetSelfPubKey(wrk.SPoS.ConsensusGroup()[0])

	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok := wrk.CheckCommitmentHashConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrCommitmentHash))

	for i := 0; i < wrk.SPoS.Threshold(bn.SrCommitmentHash); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	ok = wrk.CheckCommitmentHashConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrCommitmentHash))

	wrk.SPoS.RoundConsensus.SetSelfPubKey("2")

	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok = wrk.CheckCommitmentHashConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrCommitmentHash))

	for i := 0; i < wrk.SPoS.Threshold(bn.SrBitmap); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	ok = wrk.CheckCommitmentHashConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrCommitmentHash))

	for i := 0; i < wrk.SPoS.Threshold(bn.SrBitmap); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, false)
	}

	for i := 0; i < len(wrk.SPoS.RoundConsensus.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)

	ok = wrk.CheckCommitmentHashConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrCommitmentHash))
}

func TestWorker_IsCommitmentHashReceived(t *testing.T) {
	wrk := initWorker()

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := wrk.IsCommitmentHashReceived(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrCommitmentHash, true)
	isJobDone, _ := wrk.SPoS.GetJobDone("A", bn.SrCommitmentHash)
	assert.True(t, isJobDone)

	ok = wrk.IsCommitmentHashReceived(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("B", bn.SrCommitmentHash, true)
	ok = wrk.IsCommitmentHashReceived(2)
	assert.True(t, ok)

	wrk.SPoS.SetJobDone("C", bn.SrCommitmentHash, true)
	ok = wrk.IsCommitmentHashReceived(2)
	assert.True(t, ok)
}

func TestWorker_CommitmentHashesCollected(t *testing.T) {
	wrk := initWorker()

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := wrk.CommitmentHashesCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrBitmap, true)
	wrk.SPoS.SetJobDone("C", bn.SrBitmap, true)
	isJobDone, _ := wrk.SPoS.GetJobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = wrk.CommitmentHashesCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("B", bn.SrCommitmentHash, true)
	isJobDone, _ = wrk.SPoS.GetJobDone("B", bn.SrCommitmentHash)
	assert.True(t, isJobDone)

	ok = wrk.CommitmentHashesCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("C", bn.SrCommitmentHash, true)
	ok = wrk.CommitmentHashesCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrCommitmentHash, true)
	ok = wrk.CommitmentHashesCollected(2)
	assert.True(t, ok)
}

func TestWorker_ExtendCommitmentHash(t *testing.T) {
	wrk := initWorker()

	wrk.ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, wrk.SPoS.Status(bn.SrCommitmentHash))

	for i := 0; i < len(wrk.SPoS.RoundConsensus.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrCommitmentHash, true)
	}

	wrk.ExtendCommitmentHash()
	assert.Equal(t, spos.SsExtended, wrk.SPoS.Status(bn.SrCommitmentHash))
}

func TestWorker_GenCommitmentHashShouldRetunErrOnCreateCommitment(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()
	err := errors.New("error create commitment")
	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, err
	}

	wrk.SetMultiSigner(multiSignerMock)

	_, err2 := wrk.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnIndexSelfConsensusGroup(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetSelfPubKey("X")

	multiSignerMock := initMultisigner()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	multiSignerMock.AddCommitmentMock = func(uint16, []byte) error {
		return spos.ErrSelfNotFoundInConsensus
	}

	_, err := wrk.GenCommitmentHash()
	assert.Equal(t, spos.ErrSelfNotFoundInConsensus, err)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnAddCommitment(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()

	multiSignerMock.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return nil, nil, nil
	}

	err := errors.New("error add commitment")

	multiSignerMock.AddCommitmentMock = func(uint16, []byte) error {
		return err
	}

	wrk.SetMultiSigner(multiSignerMock)

	_, err2 := wrk.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnSetCommitmentSecret(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()

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

	wrk.SetMultiSigner(multiSignerMock)

	_, err2 := wrk.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunErrOnAddCommitmentHash(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()

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

	wrk.SetMultiSigner(multiSignerMock)

	_, err2 := wrk.GenCommitmentHash()
	assert.Equal(t, err, err2)
}

func TestWorker_GenCommitmentHashShouldRetunNil(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()

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

	wrk.SetMultiSigner(multiSignerMock)

	_, err := wrk.GenCommitmentHash()
	assert.Equal(t, nil, err)
}

func TestWorker_IsBlockSubroundUnfinishedShouldRetunFalseWhenSubroundIsFinished(t *testing.T) {
	wrk := initWorker()
	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, wrk.IsBlockSubroundUnfinished())
}

func TestWorker_IsBlockSubroundUnfinishedShouldRetunFalseWhenJobIsDone(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetSelfPubKey(wrk.SPoS.ConsensusGroup()[0])

	assert.False(t, wrk.IsBlockSubroundUnfinished())
}

func TestWorker_IsBlockSubroundUnfinishedShouldRetunTrueWhenDoBlockJobReturnsFalse(t *testing.T) {
	wrk := initWorker()

	assert.True(t, wrk.IsBlockSubroundUnfinished())
}
