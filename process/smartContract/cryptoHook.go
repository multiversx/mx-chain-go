package smartContract

import (
	"github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type cryptoHook struct {
}

func NewCryptoHook() vmcommon.CryptoHook {
	return &cryptoHook{}
}

func (ch *cryptoHook) Sha256(str string) (string, error) {
	return "", nil
}

func (ch *cryptoHook) Keccak256(str string) (string, error) {
	return "", nil
}

func (ch *cryptoHook) Ripemd160(str string) (string, error) {
	return "", nil
}

func (ch *cryptoHook) EcdsaRecover(hash string, v *big.Int, r string, s string) (string, error) {
	return "", nil
}

func (ch *cryptoHook) Bn128valid(p vmcommon.Bn128Point) (bool, error) {
	return true, nil
}

func (ch *cryptoHook) Bn128g2valid(p vmcommon.Bn128G2Point) (bool, error) {
	return true, nil
}

func (ch *cryptoHook) Bn128add(p1 vmcommon.Bn128Point, p2 vmcommon.Bn128Point) (vmcommon.Bn128Point, error) {
	return vmcommon.Bn128Point{}, nil
}

func (ch *cryptoHook) Bn128mul(k *big.Int, p vmcommon.Bn128Point) (vmcommon.Bn128Point, error) {
	return vmcommon.Bn128Point{}, nil
}

func (ch *cryptoHook) Bn128ate(l1 []vmcommon.Bn128Point, l2 []vmcommon.Bn128G2Point) (bool, error) {
	return true, nil
}
