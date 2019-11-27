package rsa_test

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto/accumulator/rsa"
	"github.com/stretchr/testify/assert"
)

func TestHashToPrime(t *testing.T) {
	t.Skip("package not in use")
	for i := 0; i < 1000; i++ {
		res := rsa.HashToPrime([]byte(strconv.Itoa(i) + "i"))
		assert.True(t, res.ProbablyPrime(50))
	}
}

func TestAccumulator_ProofForValueOk(t *testing.T) {
	acc := rsa.NewAccumulator()
	proofs := acc.Accumulate([]byte("one"), []byte("two"), []byte("three"))

	acc1 := new(big.Int).Exp(proofs[0], rsa.HashToPrime([]byte("one")), rsa.Modulus)
	acc2 := new(big.Int).Exp(proofs[1], rsa.HashToPrime([]byte("two")), rsa.Modulus)
	acc3 := new(big.Int).Exp(proofs[2], rsa.HashToPrime([]byte("three")), rsa.Modulus)

	assert.Equal(t, acc.GetValue(), acc1)
	assert.Equal(t, acc.GetValue(), acc2)
	assert.Equal(t, acc.GetValue(), acc3)

}

func TestAccumulator_WrongProofForValue(t *testing.T) {
	acc := rsa.NewAccumulator()
	proofs := acc.Accumulate([]byte("one"), []byte("two"), []byte("three"))

	acc1 := new(big.Int).Exp(proofs[0], rsa.HashToPrime([]byte("four")), rsa.Modulus)
	acc2 := new(big.Int).Exp(proofs[1], rsa.HashToPrime([]byte("three")), rsa.Modulus)

	assert.NotEqual(t, acc.GetValue(), acc1)
	assert.NotEqual(t, acc.GetValue(), acc2)
}

func TestAccumulator_VerifyCorrectData(t *testing.T) {
	acc := rsa.NewAccumulator()

	proofs := acc.Accumulate([]byte("one"), []byte("two"), []byte("three"))

	assert.True(t, acc.Verify(proofs[0], []byte("one")))
	assert.True(t, acc.Verify(proofs[1], []byte("two")))
	assert.True(t, acc.Verify(proofs[2], []byte("three")))
}

func TestAccumulator_VerifyWrongData(t *testing.T) {
	acc := rsa.NewAccumulator()

	proofs := acc.Accumulate([]byte("one"), []byte("two"), []byte("three"))

	assert.False(t, acc.Verify(proofs[0], []byte("four")))
	assert.False(t, acc.Verify(proofs[0], []byte("three")))
}

func TestAccumulator_AggregateProofsOk(t *testing.T) {
	acc := rsa.NewAccumulator()
	data := [][]byte{[]byte("one"), []byte("two"), []byte("three"), []byte("four"), []byte("five")}

	proofs := acc.Accumulate(data...)
	proof, err := acc.AggregateProofs(proofs, data...)

	assert.Nil(t, err)
	assert.True(t, acc.VerifySetOfData(proof, data...))
}

func TestAccumulator_AggregateProofsMismatchDataAndProofs(t *testing.T) {
	acc := rsa.NewAccumulator()

	data := [][]byte{[]byte("one"), []byte("two"), []byte("three"), []byte("four"), []byte("five")}
	wrongData := [][]byte{[]byte("one"), []byte("two"), []byte("three"), []byte("four"), []byte("six")}

	proofs := acc.Accumulate(data...)
	proof, _ := acc.AggregateProofs(proofs, data...)
	_, err1 := acc.AggregateProofs(proofs, []byte("one"))

	assert.False(t, acc.VerifySetOfData(proof, wrongData...))
	assert.NotNil(t, err1)
}

func BenchmarkAccumulatorHashToPrime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rsa.HashToPrime([]byte(strconv.Itoa(i)))
	}
}
