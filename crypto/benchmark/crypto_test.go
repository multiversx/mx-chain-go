package crypto_test

import (
	"crypto/cipher"
	"crypto/sha512"
	"fmt"
	"hash"
	"math/rand"
	"testing"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/pairing"
	"github.com/dedis/kyber/pairing/bn256"
	"github.com/dedis/kyber/share"
	"github.com/dedis/kyber/sign/bls"
	"github.com/dedis/kyber/sign/cosi"
	"github.com/dedis/kyber/sign/schnorr"
	"github.com/dedis/kyber/sign/tbls"
	"github.com/dedis/kyber/util/key"
	"github.com/dedis/kyber/util/random"
	"github.com/dedis/kyber/xof/blake2xb"
	"github.com/pkg/errors"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type cosiSuite struct {
	cosi.Suite
	r kyber.XOF
}

func (m *cosiSuite) Hash() hash.Hash {
	return sha512.New()
}

func (m *cosiSuite) RandomStream() cipher.Stream { return m.r }

var testSuite = &cosiSuite{edwards25519.NewBlakeSHA256Ed25519(), blake2xb.New(nil)}

/* CoSi */
func cosiSign(message []byte, privates []kyber.Scalar, n int, f int, masks []*cosi.Mask, byteMasks [][]byte) ([][]byte, error) {
	// Compute commitments
	var v []kyber.Scalar // random
	var V []kyber.Point  // commitment
	var err error
	sigs := make([][]byte, n-f)

	for i := 0; i < n-f; i++ {
		x, X := cosi.Commit(testSuite)
		v = append(v, x)
		V = append(V, X)
	}

	// Aggregate commitments
	aggV, aggMask, err := cosi.AggregateCommitments(testSuite, V, byteMasks)
	if err != nil {
		fmt.Println("Cosi commitment aggregation failed", err)
		return nil, errors.New("Cosi commitment aggregation failed")
	}

	// Set aggregate mask in nodes
	for i := 0; i < n-f; i++ {
		masks[i].SetMask(aggMask)
	}

	// Compute challenge
	var c []kyber.Scalar
	for i := 0; i < n-f; i++ {
		ci, err := cosi.Challenge(testSuite, aggV, masks[i].AggregatePublic, message)
		if err != nil {
			fmt.Println("CoSi challenge creation failed", err)
			return nil, errors.New("CoSi challenge creation failed")
		}
		c = append(c, ci)
	}

	// Compute responses
	var r []kyber.Scalar
	for i := 0; i < n-f; i++ {
		ri, _ := cosi.Response(testSuite, privates[i], v[i], c[i])
		r = append(r, ri)
	}

	// Aggregate responses
	aggr, err := cosi.AggregateResponses(testSuite, r)
	if err != nil {
		fmt.Println("CoSi responses aggregation failed", err)
		return nil, errors.New("CoSi responses aggregation failed")
	}

	for i := 0; i < n-f; i++ {
		// Sign
		sigs[i], err = cosi.Sign(testSuite, aggV, aggr, masks[i])
		if err != nil {
			fmt.Println("CoSi signing failed", err)
			return nil, errors.New("CoSi signing failed")
		}
	}

	return sigs, nil
}

func cosiVerify(message []byte, sig [][]byte, publics []kyber.Point, n int, f int) error {
	for i := 0; i < n-f; i++ {
		// Set policy depending on threshold f and then Verify
		var p cosi.Policy
		if f == 0 {
			p = nil
		} else {
			p = cosi.NewThresholdPolicy(n - f)
		}
		if err := cosi.Verify(testSuite, publics, message, sig[i], p); err != nil {
			fmt.Println("CoSi verification failed")
			return errors.New("CoSi verification failed")
		}
	}
	return nil
}

/* TBLS */
func tblsSign(msg []byte, suite pairing.Suite, pubPoly *share.PubPoly, priPoly *share.PriPoly, n int, f int) ([]byte, error) {
	var err error
	sigShares := make([][]byte, 0)
	for _, x := range priPoly.Shares(n) {
		sig, err := tbls.Sign(suite, x, msg)
		if err != nil {
			fmt.Println("TBLS share sign failed", err)
			return nil, errors.New("TBLS share sign failed")
		}
		sigShares = append(sigShares, sig)
	}
	var sig []byte
	sig, err = tbls.Recover(suite, pubPoly, msg, sigShares, f, n)
	if err != nil {
		fmt.Println("TBLS threshold sign failed", err)
		return nil, errors.New("TBLS threshold sign failed")
	}
	return sig, nil
}

func tblsVerify(sig []byte, msg []byte, suite pairing.Suite, pubKey kyber.Point) error {
	err := bls.Verify(suite, pubKey, msg, sig)

	if err != nil {
		fmt.Println("TBLS threshold sign failed", err)
		return errors.New("TBLS threshold sign failed")
	}
	return nil
}

func benchmarkMultiSig(b *testing.B, n int, f int) {
	// Generate key pairs
	var kps []*key.Pair
	var privates []kyber.Scalar
	var publics []kyber.Point

	msg := []byte(RandStringRunes(32))

	for i := 0; i < n; i++ {
		kp := key.NewKeyPair(testSuite)
		kps = append(kps, kp)
		privates = append(privates, kp.Private)
		publics = append(publics, kp.Public)
	}

	// Init masks
	var masks []*cosi.Mask
	var byteMasks [][]byte
	for i := 0; i < n-f; i++ {
		m, err := cosi.NewMask(testSuite, publics, publics[i])
		if err != nil {
			fmt.Println("Error CoSi masking")
		}
		masks = append(masks, m)
		byteMasks = append(byteMasks, masks[i].Mask())
	}

	nameCosi := fmt.Sprintf("Cosi group size:%d offline: %d", n, f)
	nameTbls := fmt.Sprintf("Tbls group size: %d, offline: %d", n, f)

	var sigs [][][]byte
	var sig [][]byte
	var err error

	b.Run(nameCosi+"sign", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sig, err = cosiSign(msg, privates, n, f, masks, byteMasks)
			if err != nil {
				return
			}
			sigs = append(sigs, sig)
		}
	})

	nbValCosi := len(sigs)

	b.Run(nameCosi+"verify", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cosiVerify(msg, sigs[i%nbValCosi], publics, n, f)
		}
	})

	var sigsTbls [][]byte
	var sigTbls []byte

	// set some threshold. e.g 20%
	t := n - (2*n)/10

	suite := bn256.NewSuite()
	secret := suite.G1().Scalar().Pick(suite.RandomStream())
	priPoly := share.NewPriPoly(suite.G2(), t, secret, suite.RandomStream())
	pubPoly := priPoly.Commit(suite.G2().Point().Base())

	b.Run(nameTbls+"sign", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sigTbls, err = tblsSign(msg, suite, pubPoly, priPoly, n, t)
			if err != nil {
				return
			}
			sigsTbls = append(sigsTbls, sigTbls)
		}
	})

	nbValTbls := len(sigsTbls)

	b.Run(nameTbls+"verify", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tblsVerify(sigsTbls[i%nbValTbls], msg, suite, pubPoly.Commit())
		}
	})
}

/* Benchmarking */
func BenchmarkSingleSig(b *testing.B) {
	suiteSch := edwards25519.NewBlakeSHA256Ed25519()
	suiteBls := bn256.NewSuite()
	var msg string
	var sigSch, sigBls []byte
	var err error
	msg = RandStringRunes(32)

	kpSchnorr := key.NewKeyPair(suiteSch)
	privateBls, publicBls := bls.NewKeyPair(suiteBls, random.New())
	var sigsSch [][]byte
	var sigsBls [][]byte

	b.Run("Schnorr sign", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sigSch, err = schnorr.Sign(suiteSch, kpSchnorr.Private, []byte(msg))
			sigsSch = append(sigsSch, sigSch)
			if err != nil {
				fmt.Println("error signing")
			}
		}
	})

	b.Run("BLS sign", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sigBls, err = bls.Sign(suiteBls, privateBls, []byte(msg))
			sigsBls = append(sigsBls, sigBls)
			if err != nil {
				fmt.Println("error signing")
			}
		}
	})

	nbValSch := len(sigsSch)
	nbValBls := len(sigsBls)

	b.Run("Verifying Schnorr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = schnorr.Verify(suiteSch, kpSchnorr.Public, []byte(msg), sigsSch[i%nbValSch])
			if err != nil {
				fmt.Println("error verifying")
			}
		}
	})

	b.Run("Verifying BLS", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = bls.Verify(suiteBls, publicBls, []byte(msg), sigsBls[i%nbValBls])
			if err != nil {
				fmt.Println("error verifying")
			}
		}
	})
}

func BenchmarkMultiSig(b *testing.B) {
	benchmarkMultiSig(b, 21, 0)
	benchmarkMultiSig(b, 67, 0)
	benchmarkMultiSig(b, 111, 0)
}
