package multisig

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
)

func (bms *BlsMultiSigner) ScalarMulSig(suite crypto.Suite, scalarBytes []byte, sigPoint *mcl.PointG1) (*mcl.PointG1, error) {
	return bms.scalarMulSig(suite, scalarBytes, sigPoint)
}
