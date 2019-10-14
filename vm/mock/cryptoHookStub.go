package mock

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type CryptoHookStub struct {
	Sha256Called    func(str string) (string, error)
	Keccak256Called func(str string) (string, error)
}

func (c *CryptoHookStub) Sha256(str string) (string, error) {
	if c.Sha256Called != nil {
		return c.Sha256Called(str)
	}
	return "", nil
}

func (c *CryptoHookStub) Keccak256(str string) (string, error) {
	if c.Keccak256Called != nil {
		return c.Keccak256Called(str)
	}
	return "", nil
}

func (c *CryptoHookStub) Ripemd160(str string) (string, error) {
	panic("implement me")
}

func (c *CryptoHookStub) EcdsaRecover(hash string, v *big.Int, r string, s string) (string, error) {
	panic("implement me")
}

func (c *CryptoHookStub) Bn128valid(p vmcommon.Bn128Point) (bool, error) {
	panic("implement me")
}

func (c *CryptoHookStub) Bn128g2valid(p vmcommon.Bn128G2Point) (bool, error) {
	panic("implement me")
}

func (c *CryptoHookStub) Bn128add(p1 vmcommon.Bn128Point, p2 vmcommon.Bn128Point) (vmcommon.Bn128Point, error) {
	panic("implement me")
}

func (c *CryptoHookStub) Bn128mul(k *big.Int, p vmcommon.Bn128Point) (vmcommon.Bn128Point, error) {
	panic("implement me")
}

func (c *CryptoHookStub) Bn128ate(l1 []vmcommon.Bn128Point, l2 []vmcommon.Bn128G2Point) (bool, error) {
	panic("implement me")
}
