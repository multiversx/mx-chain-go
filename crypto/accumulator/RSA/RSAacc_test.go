package RSA_test

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/accumulator/RSA"
)

func TestHashToPrime(t *testing.T) {
	var acc RSA.Accumulator

	for i := 0; i < 1000; i++ {
		res := acc.HashToPrime([]byte(strconv.Itoa(i) + "i"))
		assert.True(t, res.ProbablyPrime(50))
	}
}

func TestAccumulate(t *testing.T) {
	var acc RSA.Accumulator

	proofs := acc.Accumulate([]byte("one"), []byte("two"), []byte("three"))

	assert.Equal(t, 0, new(big.Int).Exp(proofs[0], acc.HashToPrime([]byte("one")), acc.GetModulus()).Cmp(acc.GetAccValue()))
	assert.Equal(t, 0, new(big.Int).Exp(proofs[1], acc.HashToPrime([]byte("two")), acc.GetModulus()).Cmp(acc.GetAccValue()))
	assert.Equal(t, 0, new(big.Int).Exp(proofs[2], acc.HashToPrime([]byte("three")), acc.GetModulus()).Cmp(acc.GetAccValue()))

	assert.NotEqual(t, 0, new(big.Int).Exp(proofs[0], acc.HashToPrime([]byte("four")), acc.GetModulus()).Cmp(acc.GetAccValue()))
	assert.NotEqual(t, 0, new(big.Int).Exp(proofs[1], acc.HashToPrime([]byte("three")), acc.GetModulus()).Cmp(acc.GetAccValue()))

}

func TestVerify(t *testing.T) {
	var acc RSA.Accumulator

	proofs := acc.Accumulate([]byte("one"), []byte("two"), []byte("three"))

	assert.True(t, acc.Verify([]byte("one"), proofs[0]))
	assert.True(t, acc.Verify([]byte("two"), proofs[1]))
	assert.True(t, acc.Verify([]byte("three"), proofs[2]))

	assert.False(t, acc.Verify([]byte("four"), proofs[0]))
	assert.False(t, acc.Verify([]byte("three"), proofs[0]))

}

func TestAggregateProofs(t *testing.T) {
	var acc RSA.Accumulator
	data := [][]byte{[]byte("one"), []byte("two"), []byte("three"), []byte("four"), []byte("five")}
	wrongData := [][]byte{[]byte("one"), []byte("two"), []byte("three"), []byte("four"), []byte("six")}

	proofs := acc.Accumulate(data...)
	proof, err := acc.AggregateProofs(proofs, data...)
	_, err1 := acc.AggregateProofs(proofs, []byte("one"))

	assert.Nil(t, err)
	assert.True(t, acc.VerifySetOfData(proof, data...))
	assert.False(t, acc.VerifySetOfData(proof, wrongData...))
	assert.NotNil(t, err1)

}

func BenchmarkAccumulatorHashToPrime(b *testing.B) {
	var acc RSA.Accumulator
	for i := 0; i < b.N; i++ {
		acc.HashToPrime([]byte(strconv.Itoa(i)))
	}
}
