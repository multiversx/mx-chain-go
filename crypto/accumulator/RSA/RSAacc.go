package RSA

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// g is the initial accumulator value
var g = big.NewInt(3)
var bigZero = big.NewInt(0)
var bigOne = big.NewInt(1)

// Accumulator represents the RSA accumulator
type Accumulator struct {
	value   *big.Int
	modulus *big.Int
	isInit  bool
}

// GetAccValue returns the value of the accumulator
func (acc *Accumulator) GetAccValue() *big.Int {
	return acc.value
}

// GetModulus returns the modulus of the accumulator
func (acc *Accumulator) GetModulus() *big.Int {
	return acc.modulus
}

func (acc *Accumulator) initAccumulator() {
	acc.value = new(big.Int).Set(g)
	acc.modulus = generateModulus()
	acc.isInit = true

}

func generateModulus() *big.Int {
	p, err := rand.Prime(rand.Reader, 1024)
	if err != nil {
		fmt.Println(err)
	}

	q, err := rand.Prime(rand.Reader, 1024)
	if err != nil {
		fmt.Println(err)
	}

	return new(big.Int).Mul(p, q)
}

// HashToPrime takes some data and maps that data to a prime number
func (*Accumulator) HashToPrime(data []byte) *big.Int {
	reader := sha3.NewShake256()
	reader.Write(data)

	p, err := rand.Prime(reader, 256)
	if err != nil {
		fmt.Println(err)
	}
	return p
}

// Accumulate takes in some data, adds each item to the accumulator, and returns proofs with which
// you can check that that data was added to the accumulator. This function does not update old proofs.
func (acc *Accumulator) Accumulate(data ...[]byte) (proofs []*big.Int) {
	var primes []*big.Int

	if !acc.isInit {
		acc.initAccumulator()
	}

	proofs = make([]*big.Int, len(data))

	for i := range data {
		primes = append(primes, acc.HashToPrime(data[i]))
	}

	for i := range primes {
		acc.value.Exp(acc.value, primes[i], acc.modulus)

	}

	for i := range primes {
		proof := new(big.Int).Set(g)
		for j := range primes {
			if i != j {
				proof.Exp(proof, primes[j], acc.modulus)
			}
		}

		proofs[i] = proof
	}

	return
}

// Verify checks, using the proof, that the data was added to the accumulator
func (acc *Accumulator) Verify(data []byte, proof *big.Int) bool {
	prime := acc.HashToPrime(data)
	v := new(big.Int).Exp(proof, prime, acc.modulus)

	return v.Cmp(acc.value) == 0
}

// VerifySetOfData verifies, given the proof for the whole set and the set, that the set
// of data has been added to the accumulator
func (acc *Accumulator) VerifySetOfData(proof *big.Int, data ...[]byte) bool {
	var primes []*big.Int
	mul := new(big.Int).Set(bigOne)

	for i := range data {
		primes = append(primes, acc.HashToPrime(data[i]))
		mul.Mul(mul, primes[i])
	}

	v := new(big.Int).Exp(proof, mul, acc.modulus)

	return v.Cmp(acc.value) == 0
}

// bezoutCoefficients computes the Bezout Coefficients of two given numbers
func bezoutCoefficients(a, b *big.Int) (oldS, oldT *big.Int) {
	var quotient *big.Int
	s := new(big.Int).Set(bigZero)
	oldS = new(big.Int).Set(bigOne)
	t := new(big.Int).Set(bigOne)
	oldT = new(big.Int).Set(bigZero)
	r := b
	oldR := a

	for r.Cmp(bigZero) != 0 {
		quotient = new(big.Int).Div(oldR, r)
		oldR, r = r, new(big.Int).Sub(oldR, new(big.Int).Mul(quotient, r))
		oldS, s = s, new(big.Int).Sub(oldS, new(big.Int).Mul(quotient, s))
		oldT, t = t, new(big.Int).Sub(oldT, new(big.Int).Mul(quotient, t))
	}

	return
}

// shamirTrick computes the (xy)-th root of an element
func (acc *Accumulator) shamirTrick(p1, p2, x, y *big.Int) *big.Int {
	a, b := bezoutCoefficients(x, y)
	mul := new(big.Int).Mul(new(big.Int).Exp(p1, b, acc.GetModulus()), new(big.Int).Exp(p2, a, acc.GetModulus()))
	return new(big.Int).Mod(mul, acc.GetModulus())
}

// AggregateProofs takes the proofs for a set of items and the set of items, and returns a single
// proof for the whole set
func (acc *Accumulator) AggregateProofs(proofs []*big.Int, data ...[]byte) (proof *big.Int, err error) {

	if len(proofs) != len(data) {
		return nil, errors.New("length of data must me equal to the length of proofs")
	}
	var primes []*big.Int
	mul := new(big.Int).Set(bigOne)

	for i := range data {
		primes = append(primes, acc.HashToPrime(data[i]))
	}

	proof = proofs[0]
	for i := 1; i < len(primes); i++ {
		mul.Mul(mul, primes[i-1])
		proof = acc.shamirTrick(proof, proofs[i], mul, primes[i])
	}

	return proof, nil
}
