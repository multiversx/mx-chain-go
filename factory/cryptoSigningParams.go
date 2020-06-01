package factory

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
)

// TODO: merge this with Crypto Components

type cryptoSigningParamsLoader struct {
	pubkeyConverter     core.PubkeyConverter
	skIndex             int
	skPemFileName       string
	suite               crypto.Suite
	skPkProviderHandler func() ([]byte, []byte, error)
}

// NewCryptoSigningParamsLoader returns a new instance of cryptoSigningParamsLoader
func NewCryptoSigningParamsLoader(
	pubkeyConverter core.PubkeyConverter,
	skIndex int,
	skPemFileName string,
	suite crypto.Suite,
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
	}
	cspf.skPkProviderHandler = cspf.getSkPk

	return cspf, nil
}

// Get returns a key generator, a private key, and a public key
func (cspf *cryptoSigningParamsLoader) Get() (*CryptoParams, error) {
	cryptoParams := &CryptoParams{}
	sk, readPk, err := cspf.skPkProviderHandler()
	if err != nil {
		return nil, err
	}

	cryptoParams.KeyGenerator = signing.NewKeyGenerator(cspf.suite)
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
