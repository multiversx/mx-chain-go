package multisig

import "github.com/ElrondNetwork/elrond-go/crypto"

func (kms *KyberMultiSignerBLS) AggregatePreparedPublicKeys(suite crypto.Suite, pubKeys ...crypto.Point) ([]byte, error) {
	return kms.aggregatePreparedPublicKeys(suite, pubKeys...)
}

func (kms *KyberMultiSignerBLS) ScalarMulSig(suite crypto.Suite, scalarBytes []byte, sig []byte) ([]byte, error) {
	return kms.scalarMulSig(suite, scalarBytes, sig)
}
