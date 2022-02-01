package libp2p

import (
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
)

func generatePrivateKey() *libp2pCrypto.Secp256k1PrivateKey {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)

	return (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
}

func TestP2pSigner_Sign(t *testing.T) {
	t.Parallel()

	signer := &p2pSigner{
		privateKey: generatePrivateKey(),
	}

	sig, err := signer.Sign([]byte("payload"))
	assert.Nil(t, err)
	assert.NotNil(t, sig)
}

func TestP2pSigner_Verify(t *testing.T) {
	t.Parallel()

	sk := generatePrivateKey()
	pk := sk.GetPublic()
	payload := []byte("payload")
	signer := &p2pSigner{
		privateKey: sk,
	}

	t.Run("invalid public key should error", func(t *testing.T) {
		t.Parallel()

		sig, err := signer.Sign(payload)
		assert.Nil(t, err)

		err = signer.Verify(payload, core.PeerID("invalid PK"), sig)
		assert.NotNil(t, err)
		assert.Equal(t, "cannot extract signing key: unexpected EOF", err.Error())
	})
	t.Run("malformed signature header should error", func(t *testing.T) {
		t.Parallel()

		sig, err := signer.Sign(payload)
		assert.Nil(t, err)

		buffPk, err := pk.Bytes()
		assert.Nil(t, err)

		sig[0] = sig[0] ^ sig[1] ^ sig[2]

		err = signer.Verify(payload, core.PeerID(buffPk), sig)
		assert.NotNil(t, err)
		assert.Equal(t, "malformed signature: no header magic", err.Error())
	})
	t.Run("altered signature should error", func(t *testing.T) {
		t.Parallel()

		sig, err := signer.Sign(payload)
		assert.Nil(t, err)

		buffPk, err := pk.Bytes()
		assert.Nil(t, err)

		sig[len(sig)-1] = sig[0] ^ sig[1] ^ sig[2]

		err = signer.Verify(payload, core.PeerID(buffPk), sig)
		assert.Equal(t, crypto.ErrInvalidSignature, err)
	})
	t.Run("sign and verify should work", func(t *testing.T) {
		t.Parallel()

		sig, err := signer.Sign(payload)
		assert.Nil(t, err)

		buffPk, err := pk.Bytes()
		assert.Nil(t, err)

		err = signer.Verify(payload, core.PeerID(buffPk), sig)
		assert.Nil(t, err)
	})
}
