package hashing

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-vm-common"
)

type VMCrypto struct {
}

func (vmc *VMCrypto) Sha256(str string) (string, error) {
	panic("implement me")
}

func (vmc *VMCrypto) Keccak256(str string) (string, error) {
	panic("implement me")
}

func (vmc *VMCrypto) Ripemd160(str string) (string, error) {
	panic("implement me")
}

func (vmc *VMCrypto) EcdsaRecover(hash string, v *big.Int, r string, s string) (string, error) {
	panic("implement me")
}

func (vmc *VMCrypto) Bn128valid(p vmcommon.Bn128Point) (bool, error) {
	panic("implement me")
}

func (vmc *VMCrypto) Bn128g2valid(p vmcommon.Bn128G2Point) (bool, error) {
	panic("implement me")
}

func (vmc *VMCrypto) Bn128add(p1 vmcommon.Bn128Point, p2 vmcommon.Bn128Point) (vmcommon.Bn128Point, error) {
	panic("implement me")
}

func (vmc *VMCrypto) Bn128mul(k *big.Int, p vmcommon.Bn128Point) (vmcommon.Bn128Point, error) {
	panic("implement me")
}

func (vmc *VMCrypto) Bn128ate(l1 []vmcommon.Bn128Point, l2 []vmcommon.Bn128G2Point) (bool, error) {
	panic("implement me")
}
