package mock

import "github.com/ElrondNetwork/elrond-go-sandbox/crypto"

type PubKeyStub struct {
	ToByteArrayCalled func() ([]byte, error)
	SuiteCalled       func() crypto.Suite
	PointCalled       func() crypto.Point
}

func (pks *PubKeyStub) ToByteArray() ([]byte, error) {
	return pks.ToByteArrayCalled()
}

func (pks *PubKeyStub) Suite() crypto.Suite {
	return pks.SuiteCalled()
}

func (pks *PubKeyStub) Point() crypto.Point {
	return pks.PointCalled()
}
