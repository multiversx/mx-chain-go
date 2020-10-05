package factory

import (
	"bytes"
	"encoding/hex"
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
)

// TODO: merge this with Crypto Components
var log = logger.GetOrCreate("main/factory")

type cryptoSigningParamsLoader struct {
	pubkeyConverter     core.PubkeyConverter
	skIndex             int
	skPemFileName       string
	suite               crypto.Suite
	skPkProviderHandler func() ([]byte, []byte, error)
	isInImportMode      bool
}

// NewCryptoSigningParamsLoader returns a new instance of cryptoSigningParamsLoader
func NewCryptoSigningParamsLoader(
	pubkeyConverter core.PubkeyConverter,
	skIndex int,
	skPemFileName string,
	suite crypto.Suite,
	isInImportMode bool,
) (*cryptoSigningParamsLoader, error) {
	if check.IfNil(pubkeyConverter) {
		return nil, ErrNilPubKeyConverter
	}
	if check.IfNil(suite) {
		return nil, ErrNilSuite
	}

	cspf := &cryptoSigningParamsLoader{
		pubkeyConverter: pubkeyConverter,
		skIndex:         skIndex,
		skPemFileName:   skPemFileName,
		suite:           suite,
		isInImportMode:  isInImportMode,
	}
	cspf.skPkProviderHandler = cspf.getSkPk

	return cspf, nil
}

// Get returns a key generator, a private key, and a public key
func (cspf *cryptoSigningParamsLoader) Get() (*CryptoParams, error) {
	cryptoParams := &CryptoParams{}
	cryptoParams.KeyGenerator = signing.NewKeyGenerator(cspf.suite)

	if cspf.isInImportMode {
		return cspf.generateCryptoParams(cryptoParams)
	}

	return cspf.readCryptoParams(cryptoParams)
}

func (cspf *cryptoSigningParamsLoader) readCryptoParams(cryptoParams *CryptoParams) (*CryptoParams, error) {
	sk, readPk, err := cspf.skPkProviderHandler()
	if err != nil {
		return nil, err
	}

	cryptoParams.PrivateKey, err = cryptoParams.KeyGenerator.PrivateKeyFromByteArray(sk)
	if err != nil {
		return nil, err
	}

	cryptoParams.PublicKey = cryptoParams.PrivateKey.GeneratePublic()
	if len(readPk) > 0 {
		cryptoParams.PublicKeyBytes, err = cryptoParams.PublicKey.ToByteArray()
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(cryptoParams.PublicKeyBytes, readPk) {
			return nil, ErrPublicKeyMismatch
		}
	}

	cryptoParams.PublicKeyString = cspf.pubkeyConverter.Encode(cryptoParams.PublicKeyBytes)

	return cryptoParams, nil
}

func (cspf *cryptoSigningParamsLoader) generateCryptoParams(cryptoParams *CryptoParams) (*CryptoParams, error) {
	log.Warn("the node is in import mode! Will generate a fresh new BLS key")
	cryptoParams.PrivateKey, cryptoParams.PublicKey = cryptoParams.KeyGenerator.GeneratePair()

	var err error
	cryptoParams.PublicKeyBytes, err = cryptoParams.PublicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	cryptoParams.PublicKeyString = cspf.pubkeyConverter.Encode(cryptoParams.PublicKeyBytes)

	return cryptoParams, nil
}

func (cspf *cryptoSigningParamsLoader) getSkPk() ([]byte, []byte, error) {
	skIndex := cspf.skIndex
	encodedSk, pkString, err := core.LoadSkPkFromPemFile(cspf.skPemFileName, skIndex)
	if err != nil {
		return nil, nil, err
	}

	skBytes, err := hex.DecodeString(string(encodedSk))
	if err != nil {
		return nil, nil, fmt.Errorf("%w for encoded secret key", err)
	}

	pkBytes, err := cspf.pubkeyConverter.Decode(pkString)
	if err != nil {
		return nil, nil, fmt.Errorf("%w for encoded public key %s", err, pkString)
	}

	return skBytes, pkBytes, nil
}
