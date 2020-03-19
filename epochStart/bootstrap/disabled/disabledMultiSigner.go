package disabled

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

type multiSigner struct {
}

// NewMultiSigner returns a new instance of multiSigner
func NewMultiSigner() *multiSigner {
	return &multiSigner{}
}

func (m *multiSigner) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	return nil, nil
}

func (m *multiSigner) SetAggregatedSig([]byte) error {
	return nil
}

func (m *multiSigner) Verify(msg []byte, bitmap []byte) error {
	return nil
}

func (m *multiSigner) Reset(pubKeys []string, index uint16) error {
	return nil
}

func (m *multiSigner) CreateSignatureShare(msg []byte, bitmap []byte) ([]byte, error) {
	return nil, nil
}

func (m *multiSigner) StoreSignatureShare(index uint16, sig []byte) error {
	return nil
}

func (m *multiSigner) SignatureShare(index uint16) ([]byte, error) {
	return nil, nil
}

func (m *multiSigner) VerifySignatureShare(index uint16, sig []byte, msg []byte, bitmap []byte) error {
	return nil
}

func (m *multiSigner) AggregateSigs(bitmap []byte) ([]byte, error) {
	return nil, nil
}

func (m *multiSigner) IsInterfaceNil() bool {
	return m == nil
}
