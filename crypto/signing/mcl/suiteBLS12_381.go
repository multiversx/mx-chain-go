package mcl

import (
	"crypto/cipher"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/herumi/bls-go-binary/bls"
)

var log = logger.GetOrCreate("process/block")

var _ crypto.Group = (*SuiteBLS12)(nil)
var _ crypto.Random = (*SuiteBLS12)(nil)
var _ crypto.Suite = (*SuiteBLS12)(nil)

// SuiteBLS12 provides an implementation of the Suite interface for BLS12-381
type SuiteBLS12 struct {
	G1       *groupG1
	G2       *groupG2
	GT       *groupGT
	strSuite string
}

// Note: BLS_SWAP_G flag, (currently this flag is not set)
// Compiling with the flag will give Public Keys on G1 (48 bytes) and Signatures on G2 (96 bytes)
// Compiling without the flag will give Public Keys on G2 and Signatures on G1
// For Elrond the public keys for the validators are known during an epoch and also are not set on blocks
// BLS signatures are however set on every block header, so in order to optimise the header size flag will be false
// to have smaller signatures, so on G1(48 bytes)

var (
	g2str  = "1 352701069587466618187139116011060144890029952792775240219908644239793785735715026873347600343865175952761926303160 3059144344244213709971259814753781636986470325476647558659373206291635324768958432433509563104347017837885763365758 1985150602287291935568054521177171638300868978215655730859378665066344726373823718423869104263333984641494340347905 927553665492332455747201965776037880757740193453592970025027978793976877002675564980949289727957565575433344219582"
	g1str  = "1 3685416753713387016781088315183077757961620795782546409894578378688607592378376318836054947676345821548104185464507 1339506544944476473020471379941921221584933875938349620426543736416511423956333506472724655353366534992391756441569"
	doInit sync.Once
)

func blsInit() {
	if err := bls.Init(bls.BLS12_381); err != nil {
		panic(fmt.Sprintf("could not initialize BLS12-381 curve %v", err))
	}

	pubKey := &bls.PublicKey{}
	bls.BlsGetGeneratorOfPublicKey(pubKey)
	generatorG2 := bls.CastFromPublicKey(pubKey)
	g2str = generatorG2.GetString(10)
}

func init() {
	doInit.Do(blsInit)
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
func (s *SuiteBLS12) CreatePointForScalar(scalar crypto.Scalar) (crypto.Point, error) {
	if check.IfNil(scalar) {
		return nil, crypto.ErrNilPrivateKeyScalar
	}
	sc, ok := scalar.GetUnderlyingObj().(*bls.Fr)
	if !ok {
		return nil, crypto.ErrInvalidScalar
	}

	if sc.IsZero() || !sc.IsValid() {
		return nil, crypto.ErrInvalidPrivateKey
	}

	point := s.G2.CreatePointForScalar(scalar)

	return point, nil
}

// PointLen returns the max length of point in nb of bytes
func (s *SuiteBLS12) PointLen() int {
	return s.G2.PointLen()
}

// CreateKeyPair returns a pair of private public BLS keys.
// The private key is a scalarInt, while the public key is a Point on G2 curve
func (s *SuiteBLS12) CreateKeyPair() (crypto.Scalar, crypto.Point) {
	var sc crypto.Scalar
	var err error

	sc = s.G2.CreateScalar()
	sc, err = sc.Pick()
	if err != nil {
		log.Error("SuiteBLS12 CreateKeyPair", "error", err.Error())
		return nil, nil
	}

	p := s.G2.CreatePointForScalar(sc)

	return sc, p
}

// GetUnderlyingSuite returns the underlying suite
func (s *SuiteBLS12) GetUnderlyingSuite() interface{} {
	return s
}

// CheckPointValid returns error if the point is not valid (zero is also not valid), otherwise nil
func (s *SuiteBLS12) CheckPointValid(pointBytes []byte) error {
	if len(pointBytes) != s.PointLen() {
		return crypto.ErrInvalidParam
	}

	point := s.G2.CreatePoint()
	err := point.UnmarshalBinary(pointBytes)
	if err != nil {
		return err
	}

	pG2, ok := point.GetUnderlyingObj().(*bls.G2)
	if !ok || !pG2.IsValid() || !pG2.IsValidOrder() || pG2.IsZero() {
		return crypto.ErrInvalidPoint
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SuiteBLS12) IsInterfaceNil() bool {
	return s == nil
}

// baseG1 returns the generator point for G1
func baseG1() string {
	return g1str
}

// baseG2 returns the generator point for G2
func baseG2() string {
	return g2str
}
