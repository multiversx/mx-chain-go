package rsa

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

// g is the initial accumulator value
var g = big.NewInt(3)
var bigZero = big.NewInt(0)
var bigOne = big.NewInt(1)
var log = logger.NewDefaultLogger()

// Modulus taken from https://en.wikipedia.org/wiki/RSA_numbers#RSA-2048
const modulus = "25195908475657893494027183240048398571429282126204032027777137836043662020707595556264018525880784406918290641249515082189298559149176184502808489120072844992687392807287776735971418347270261896375014971824691165077613379859095700097330459748808428401797429100642458691817195118746121515172654632282216869987549182422433637259085141865462043576798423387184774447920739934236584823824281198163815010674810451660377306056201619676256133844143603833904414952634432190114657544454178424020924616515723350778707749817125772467962926386356373289912154831438167899885040445364023527381951378636564391212010397122822120720357"

type accumulator struct {
	value *big.Int
}

// NewAccumulator returns an initialised rsa accumulator
func NewAccumulator() *accumulator {
	return &accumulator{new(big.Int).Set(g)}
}

// GetValue returns the value of the accumulator
func (acc *accumulator) GetValue() *big.Int {
	return acc.value
}

// HashToPrime takes some data and maps that data to a prime number
func HashToPrime(data []byte) *big.Int {
	var b2b blake2b.Blake2b

	hash := b2b.Compute(string(data))
	p := new(big.Int).SetBytes(hash)

	for !p.ProbablyPrime(20) {
		hash = b2b.Compute(p.String())
		p.SetBytes(hash)
	}
	return p
}

// Accumulate takes in some data, adds each item to the accumulator, and returns proofs with which
// you can check that that data was added to the accumulator. This function does not update old proofs.
func (acc *accumulator) Accumulate(data ...[]byte) (proofs []*big.Int) {
	var primes []*big.Int

	proofs = make([]*big.Int, len(data))

	for i := range data {
		primes = append(primes, HashToPrime(data[i]))
	}

	for i := range primes {
		_ = acc.value.Exp(acc.value, primes[i], GetModulus())

	}

	for i := range primes {
		proof := new(big.Int).Set(g)
		for j := range primes {
			if i != j {
				_ = proof.Exp(proof, primes[j], GetModulus())
			}
		}

		proofs[i] = proof
	}

	return
}

// Verify checks, using the proof, that the data was added to the accumulator
func (acc *accumulator) Verify(proof *big.Int, data []byte) bool {
	prime := HashToPrime(data)
	v := new(big.Int).Exp(proof, prime, GetModulus())

	return v.Cmp(acc.value) == 0
}

// VerifySetOfData verifies if a set of data has been added to the accumulator
func (acc *accumulator) VerifySetOfData(proof *big.Int, data ...[]byte) bool {
	var primes []*big.Int
	mul := new(big.Int).Set(bigOne)

	for i := range data {
		primes = append(primes, HashToPrime(data[i]))
		_ = mul.Mul(mul, primes[i])
	}

	v := new(big.Int).Exp(proof, mul, GetModulus())

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
func shamirTrick(p1, p2, x, y *big.Int) *big.Int {

	a, b := bezoutCoefficients(x, y)
	mul := new(big.Int).Mul(new(big.Int).Exp(p1, b, GetModulus()), new(big.Int).Exp(p2, a, GetModulus()))
	return new(big.Int).Mod(mul, GetModulus())
}

// AggregateProofs takes the proofs for a set of items and the set of items, and returns a single
// proof for the whole set
func (acc *accumulator) AggregateProofs(proofs []*big.Int, data ...[]byte) (proof *big.Int, err error) {

	if len(proofs) != len(data) {
		return nil, errors.New("length of data must me equal to the length of proofs")
	}
	var primes []*big.Int
	mul := new(big.Int).Set(bigOne)

	for i := range data {
		primes = append(primes, HashToPrime(data[i]))
	}

	proof = proofs[0]
	for i := 1; i < len(primes); i++ {
		_ = mul.Mul(mul, primes[i-1])
		proof = shamirTrick(proof, proofs[i], mul, primes[i])
	}

	return proof, nil
}

func GetModulus() *big.Int {
	m, ok := new(big.Int).SetString(modulus, 10)

	if !ok {
		log.Error("error converting modulus to big int")
	}
	return m
}
