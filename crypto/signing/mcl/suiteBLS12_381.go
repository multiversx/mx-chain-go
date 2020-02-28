package mcl

import (
	"crypto/cipher"
	"fmt"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/bls-go-binary/bls"
)

type SuiteBLS12 struct {
	G1       *groupG1
	G2       *groupG2
	GT       *groupGT
	strSuite string
}

// Enable if compiling mcl/bls/bls-go-binary with BLS_SWAP_G flag
// Compiling with the flag will give Public Keys on G1 (48 bytes) and Signatures on G2 (96 bytes)
// Compiling without the flag will give Public Keys on G2 and Signatures on G1
// For Elrond the public keys for the validators are known during an epoch and also are not set on blocks
// BLS signatures are however set on every block header, so in order to optimise the header size flag will be false
// to have smaller signatures, so on G1(48 bytes)
const bls_swap_g = false

const g2str = "1 352701069587466618187139116011060144890029952792775240219908644239793785735715026873347600343865175952761926303160 3059144344244213709971259814753781636986470325476647558659373206291635324768958432433509563104347017837885763365758 1985150602287291935568054521177171638300868978215655730859378665066344726373823718423869104263333984641494340347905 927553665492332455747201965776037880757740193453592970025027978793976877002675564980949289727957565575433344219582"
const g1str = "1 3685416753713387016781088315183077757961620795782546409894578378688607592378376318836054947676345821548104185464507 1339506544944476473020471379941921221584933875938349620426543736416511423956333506472724655353366534992391756441569"

var basePointG1Str atomic.Value
var basePointG2Str atomic.Value

func init() {
	if err := bls.Init(bls.BLS12_381); err != nil {
		panic(fmt.Sprintf("could not initialize BLS12-381 curve %v", err))
	}

	pubKey := &bls.PublicKey{}
	bls.BlsGetGeneratorForPublicKey(pubKey)
	if bls_swap_g {
		generatorG1 := &bls.G1{}
		bls.BlsPublicKeyToG1(pubKey, generatorG1)
		basePointG1Str.Store(generatorG1.GetString(10))
		basePointG2Str.Store(g2str)
	} else {
		generatorG2 := &bls.G2{}
		bls.BlsPublicKeyToG2(pubKey, generatorG2)
		basePointG1Str.Store(g1str)
		basePointG2Str.Store(generatorG2.GetString(10))
	}
}

// NewSuiteBLS12 returns a wrapper over a BLS12 curve.
func NewSuiteBLS12() *SuiteBLS12 {
	return &SuiteBLS12{
		G1:       &groupG1{},
		G2:       &groupG2{},
		GT:       &groupGT{},
		strSuite: "BLS12-381 suite",
	}
}

// RandomStream returns a cipher.Stream that returns a key stream
// from crypto/rand.
func (s *SuiteBLS12) RandomStream() cipher.Stream {
	// random stream is internal in mcl library so not needed
	return nil
}

// CreatePoint creates a new point
func (s *SuiteBLS12) CreatePoint() crypto.Point {
	return s.G2.CreatePoint()
}

// String returns the string for the group
func (s *SuiteBLS12) String() string {
	return s.strSuite
}

// ScalarLen returns the maximum length of scalars in bytes
func (s *SuiteBLS12) ScalarLen() int {
	return s.G2.ScalarLen()
}

// CreateScalar creates a new Scalar
func (s *SuiteBLS12) CreateScalar() crypto.Scalar {
	return s.G2.CreateScalar()
}

// CreatePointForScalar creates a new point corresponding to the given scalar
func (s *SuiteBLS12) CreatePointForScalar(scalar crypto.Scalar) crypto.Point {
	return s.G2.CreatePointForScalar(scalar)
}

// PointLen returns the max length of point in nb of bytes
func (s *SuiteBLS12) PointLen() int {
	return s.G2.PointLen()
}

// CreateKeyPair returns a pair of private public BLS keys.
// The private key is a scalarInt, while the public key is a Point on G2 curve
func (s *SuiteBLS12) CreateKeyPair(stream cipher.Stream) (crypto.Scalar, crypto.Point) {
	var sc crypto.Scalar
	var p crypto.Point

	sc = s.G2.CreateScalar()
	sc, _ = sc.Pick(stream)
	p = s.G2.CreatePointForScalar(sc)

	return sc, p
}

// GetUnderlyingSuite returns the underlying suite
func (s *SuiteBLS12) GetUnderlyingSuite() interface{} {
	return s
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SuiteBLS12) IsInterfaceNil() bool {
	return s == nil
}

// BaseG1 returns the generator point for G1
func BaseG1() string {
	v := basePointG1Str.Load()
	vStr := v.(string)
	return vStr
}

// BaseG2 returns the generator point for G2
func BaseG2() string {
	v := basePointG2Str.Load()
	vStr := v.(string)
	return vStr
}
