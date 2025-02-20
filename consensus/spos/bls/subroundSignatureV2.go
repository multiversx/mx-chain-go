package bls

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type subroundSignatureV2 struct {
	*subroundSignature
}

// NewSubroundSignatureV2 creates a subroundSignatureV2 object
func NewSubroundSignatureV2(subroundSignature *subroundSignature) (*subroundSignatureV2, error) {
	if subroundSignature == nil {
		return nil, spos.ErrNilSubround
	}

	sr := &subroundSignatureV2{
		subroundSignature,
	}

	sr.getMessageToSignFunc = sr.getMessageToSign

	return sr, nil
}

func (sr *subroundSignatureV2) getMessageToSign() []byte {
	headerHash, err := core.CalculateHash(sr.Marshalizer(), sr.Hasher(), sr.Header)
	if err != nil {
		log.Error("subroundSignatureV2.getMessageToSign", "error", err.Error())
		return nil
	}

	return headerHash
}
