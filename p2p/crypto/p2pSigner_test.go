package crypto

import (
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func generatePrivateKey() *libp2pCrypto.Secp256k1PrivateKey {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)

	return (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
}

func TestNewP2PSigner(t *testing.T) {
	t.Parallel()

	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		signer, err := NewP2PSigner(nil)

		assert.Nil(t, signer)
		assert.Equal(t, errNilPrivateKey, err)
	})
	t.Run("valid private key should work", func(t *testing.T) {
		t.Parallel()

		signer, err := NewP2PSigner(generatePrivateKey())

		assert.NotNil(t, signer)
		assert.Nil(t, err)
	})
}

func TestP2pSigner_Sign(t *testing.T) {
	t.Parallel()

	signer, _ := NewP2PSigner(generatePrivateKey())

	sig, err := signer.Sign([]byte("payload"))
	assert.Nil(t, err)
	assert.NotNil(t, sig)
}

func TestP2pSigner_Verify(t *testing.T) {
	t.Parallel()

	sk := generatePrivateKey()
	pk := sk.GetPublic()
	payload := []byte("payload")
	signer, _ := NewP2PSigner(sk)
	libp2pPid, _ := peer.IDFromPublicKey(pk)

	t.Run("invalid public key should error", func(t *testing.T) {
		t.Parallel()

		sig, err := signer.Sign(payload)
		assert.Nil(t, err)

		err = signer.Verify(payload, "invalid PK", sig)
		assert.NotNil(t, err)
		assert.Equal(t, "length greater than remaining number of bytes in buffer", err.Error())
	})
	t.Run("malformed signature header should error", func(t *testing.T) {
		t.Parallel()

		sig, err := signer.Sign(payload)
		assert.Nil(t, err)

		sig[0] = sig[0] ^ sig[1] ^ sig[2]

		err = signer.Verify(payload, core.PeerID(libp2pPid), sig)
		assert.NotNil(t, err)
		assert.Equal(t, "malformed signature: no header magic", err.Error())
	})
	t.Run("altered signature should error", func(t *testing.T) {
		t.Parallel()

		sig, err := signer.Sign(payload)
		assert.Nil(t, err)

		sig[len(sig)-1] = sig[0] ^ sig[1] ^ sig[2]

		err = signer.Verify(payload, core.PeerID(libp2pPid), sig)
		assert.Equal(t, crypto.ErrInvalidSignature, err)
	})
	t.Run("sign and verify should work", func(t *testing.T) {
		t.Parallel()

		sig, err := signer.Sign(payload)
		assert.Nil(t, err)

		err = signer.Verify(payload, core.PeerID(libp2pPid), sig)
		assert.Nil(t, err)
	})
}

func TestP2PSigner_SignUsingPrivateKey(t *testing.T) {
	t.Parallel()

	payload := []byte("payload")

	generator := NewIdentityGenerator()
	skBytes1, pid1, err := generator.CreateRandomP2PIdentity()
	assert.Nil(t, err)

	skBytes2, pid2, err := generator.CreateRandomP2PIdentity()
	assert.Nil(t, err)
	assert.NotEqual(t, skBytes1, skBytes2)

	p2pSigner := &p2pSigner{}

	sig1, err := p2pSigner.SignUsingPrivateKey(skBytes1, payload)
	assert.Nil(t, err)

	sig2, err := p2pSigner.SignUsingPrivateKey(skBytes2, payload)
	assert.Nil(t, err)
	assert.NotEqual(t, sig1, sig2)

	assert.Nil(t, p2pSigner.Verify(payload, pid1, sig1))
	assert.Nil(t, p2pSigner.Verify(payload, pid2, sig2))
}

func TestP2pSigner_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	numOps := 1000
	wg := sync.WaitGroup{}
	wg.Add(numOps)

	sk := generatePrivateKey()
	pk := sk.GetPublic()
	payload1 := []byte("payload1")
	payload2 := []byte("payload2")
	payload3 := []byte("payload3")
	signer, _ := NewP2PSigner(sk)
	libp2pPid, _ := peer.IDFromPublicKey(pk)
	pid := core.PeerID(libp2pPid)

	sig1, _ := signer.Sign(payload1)

	generator := NewIdentityGenerator()
	skBytes, _, err := generator.CreateRandomP2PIdentity()
	assert.Nil(t, err)

	for i := 0; i < numOps; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)

			switch idx {
			case 0:
				_, errSign := signer.Sign(payload2)
				assert.Nil(t, errSign)
			case 1:
				errVerify := signer.Verify(payload1, pid, sig1)
				assert.Nil(t, errVerify)
			case 2:
				_, errSignWithSK := signer.SignUsingPrivateKey(skBytes, payload3)
				assert.Nil(t, errSignWithSK)
			}

			wg.Done()
		}(i % 3)
	}

	wg.Wait()
}
