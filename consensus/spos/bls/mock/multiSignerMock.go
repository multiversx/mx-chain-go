package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type MultiSignerMock struct{}

func (msm *MultiSignerMock) Reset(pubKeys []string, index uint16) error {
	return nil
}

func (msm *MultiSignerMock) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	return &MultiSignerMock{}, nil
}

func (msm *MultiSignerMock) SetMessage(msg []byte) error {
	return nil
}

func (msm *MultiSignerMock) SetAggregatedSig([]byte) error {
	return nil
}

func (msm *MultiSignerMock) Verify(msg []byte, bitmap []byte) error {
	return nil
}

func (msm *MultiSignerMock) CreateCommitment() (commSecret []byte, commitment []byte) {
	return []byte("bls"), []byte("bls")
}

func (msm *MultiSignerMock) StoreCommitmentHash(index uint16, commHash []byte) error {
	return nil
}

func (msm *MultiSignerMock) CommitmentHash(index uint16) ([]byte, error) {
	return []byte("bls"), nil
}

func (msm *MultiSignerMock) StoreCommitment(index uint16, value []byte) error {
	return nil
}

func (msm *MultiSignerMock) Commitment(index uint16) ([]byte, error) {
	return []byte("bls"), nil
}

func (msm *MultiSignerMock) AggregateCommitments(bitmap []byte) error {
	return nil
}

func (msm *MultiSignerMock) CreateSignatureShare(msg []byte, bitmap []byte) ([]byte, error) {
	return []byte("bls"), nil
}

func (msm *MultiSignerMock) StoreSignatureShare(index uint16, sig []byte) error {
	return nil
}

func (msm *MultiSignerMock) SignatureShare(index uint16) ([]byte, error) {
	return []byte("bls"), nil
}

func (msm *MultiSignerMock) VerifySignatureShare(index uint16, sig []byte, msg []byte, bitmap []byte) error {
	return nil
}

func (msm *MultiSignerMock) AggregateSigs(bitmap []byte) ([]byte, error) {
	return []byte("bls"), nil
}
