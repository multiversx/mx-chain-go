package process

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/vm"
)

type messageSigVerifier struct {
	kg           crypto.KeyGenerator
	singleSigner crypto.SingleSigner
}

func NewMessageSigVerifier(
	kg crypto.KeyGenerator,
	singleSigner crypto.SingleSigner,
) (*messageSigVerifier, error) {

	if check.IfNil(kg) {
		return nil, vm.ErrNilKeyGenerator
	}
	if check.IfNil(singleSigner) {
		return nil, vm.ErrSingleSigner
	}

	return &messageSigVerifier{
		kg:           kg,
		singleSigner: singleSigner,
	}, nil
}

// Verify checks if message was signed by public key given as byte array
func (m *messageSigVerifier) Verify(message []byte, signedMessage []byte, pubKey []byte) error {
	actPubKey, err := m.kg.PublicKeyFromByteArray(pubKey)
	if err != nil {
		return err
	}

	return m.singleSigner.Verify(actPubKey, message, signedMessage)
}

// IsInterfaceNil returns if underlying object is nil
func (m *messageSigVerifier) IsInterfaceNil() bool {
	panic("implement me")
}
