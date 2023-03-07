package peerSignatureHandler_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go"
	errorsErd "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/peerSignatureHandler"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerSignatureHandler_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	peerSigHandler, err := peerSignatureHandler.NewPeerSignatureHandler(
		nil,
		&cryptoMocks.SingleSignerStub{},
		&cryptoMocks.KeyGenStub{},
	)

	assert.True(t, check.IfNil(peerSigHandler))
	assert.Equal(t, errorsErd.ErrNilCacher, err)
}

func TestNewPeerSignatureHandler_NilSingleSignerShouldErr(t *testing.T) {
	t.Parallel()

	peerSigHandler, err := peerSignatureHandler.NewPeerSignatureHandler(
		testscommon.NewCacherMock(),
		nil,
		&cryptoMocks.KeyGenStub{},
	)

	assert.True(t, check.IfNil(peerSigHandler))
	assert.Equal(t, errorsErd.ErrNilSingleSigner, err)
}

func TestNewPeerSignatureHandler_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	peerSigHandler, err := peerSignatureHandler.NewPeerSignatureHandler(
		testscommon.NewCacherMock(),
		&cryptoMocks.SingleSignerStub{},
		nil,
	)

	assert.True(t, check.IfNil(peerSigHandler))
	assert.Equal(t, crypto.ErrNilKeyGenerator, err)
}

func TestNewPeerSignatureHandler_OkParamsShouldWork(t *testing.T) {
	t.Parallel()

	peerSigHandler, err := peerSignatureHandler.NewPeerSignatureHandler(
		testscommon.NewCacherMock(),
		&cryptoMocks.SingleSignerStub{},
		&cryptoMocks.KeyGenStub{},
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(peerSigHandler))
}

func TestPeerSignatureHandler_VerifyPeerSignatureInvalidPk(t *testing.T) {
	t.Parallel()

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		testscommon.NewCacherMock(),
		&cryptoMocks.SingleSignerStub{},
		&cryptoMocks.KeyGenStub{},
	)

	err := peerSigHandler.VerifyPeerSignature(nil, "dummy peer", []byte("signature"))
	assert.Equal(t, crypto.ErrInvalidPublicKey, err)
}

func TestPeerSignatureHandler_VerifyPeerSignatureInvalidPID(t *testing.T) {
	t.Parallel()

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		testscommon.NewCacherMock(),
		&cryptoMocks.SingleSignerStub{},
		&cryptoMocks.KeyGenStub{},
	)

	err := peerSigHandler.VerifyPeerSignature([]byte("public key"), "", []byte("signature"))
	assert.Equal(t, errorsErd.ErrInvalidPID, err)
}

func TestPeerSignatureHandler_VerifyPeerSignatureInvalidSignature(t *testing.T) {
	t.Parallel()

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		testscommon.NewCacherMock(),
		&cryptoMocks.SingleSignerStub{},
		&cryptoMocks.KeyGenStub{},
	)

	err := peerSigHandler.VerifyPeerSignature([]byte("public key"), "dummy peer", nil)
	assert.Equal(t, errorsErd.ErrInvalidSignature, err)
}

func TestPeerSignatureHandler_VerifyPeerSignatureCantGetPubKeyBytes(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("keygen err")
	keyGen := &cryptoMocks.KeyGenStub{
		PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
			return nil, expectedErr
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		testscommon.NewCacherMock(),
		&cryptoMocks.SingleSignerStub{},
		keyGen,
	)

	err := peerSigHandler.VerifyPeerSignature([]byte("public key"), "dummy peer", []byte("signature"))
	assert.Equal(t, expectedErr, err)
}

func TestPeerSignatureHandler_VerifyPeerSignatureSigNotFoundInCache(t *testing.T) {
	t.Parallel()

	pk := []byte("public key")
	verifyCalled := false
	pid := "dummy peer"
	sig := []byte("signature")

	cache := testscommon.NewCacherMock()
	keyGen := &cryptoMocks.KeyGenStub{
		PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
			return &cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return pk, nil
				},
			}, nil
		},
	}
	singleSigner := &cryptoMocks.SingleSignerStub{
		VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			verifyCalled = true
			return nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		keyGen,
	)

	err := peerSigHandler.VerifyPeerSignature(pk, core.PeerID(pid), sig)
	assert.Nil(t, err)
	assert.True(t, verifyCalled)

	val, ok := cache.Get(pk)
	assert.True(t, ok)
	assert.NotNil(t, val)

	recoveredPid, recoveredSig, err := peerSigHandler.GetPIDAndSig(val)
	assert.Nil(t, err)
	assert.Equal(t, core.PeerID(pid), recoveredPid)
	assert.Equal(t, sig, recoveredSig)
}

func TestPeerSignatureHandler_VerifyPeerSignatureWrongEntryInCache(t *testing.T) {
	t.Parallel()

	wrongType := []byte("wrong type")
	pk := []byte("public key")
	verifyCalled := false
	pid := "dummy peer"
	sig := []byte("signature")

	cache := testscommon.NewCacherMock()
	cache.Put(pk, wrongType, len(wrongType))

	keyGen := &cryptoMocks.KeyGenStub{
		PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
			return &cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return pk, nil
				},
			}, nil
		},
	}
	singleSigner := &cryptoMocks.SingleSignerStub{
		VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			verifyCalled = true
			return nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		keyGen,
	)

	err := peerSigHandler.VerifyPeerSignature(pk, core.PeerID(pid), sig)
	assert.Nil(t, err)
	assert.True(t, verifyCalled)

	val, ok := cache.Get(pk)
	assert.True(t, ok)
	assert.NotNil(t, val)

	recoveredPid, recoveredSig, err := peerSigHandler.GetPIDAndSig(val)
	assert.Nil(t, err)
	assert.Equal(t, core.PeerID(pid), recoveredPid)
	assert.Equal(t, sig, recoveredSig)
}

func TestPeerSignatureHandler_VerifyPeerSignatureNewPidAndSig(t *testing.T) {
	t.Parallel()

	pk := []byte("public key")
	verifyCalled := false
	pid := core.PeerID("dummy peer")
	sig := []byte("signature")
	newPid := core.PeerID("new dummy peer")
	newSig := []byte("new sig")

	cache := testscommon.NewCacherMock()
	keyGen := &cryptoMocks.KeyGenStub{
		PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
			return &cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return pk, nil
				},
			}, nil
		},
	}
	singleSigner := &cryptoMocks.SingleSignerStub{
		VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			verifyCalled = true
			return nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		keyGen,
	)

	cacheEntry := peerSigHandler.GetCacheEntry(pid, sig)
	cache.Put(pk, cacheEntry, len(pid)+len(sig))

	err := peerSigHandler.VerifyPeerSignature(pk, newPid, newSig)
	assert.Nil(t, err)
	assert.True(t, verifyCalled)

	val, ok := cache.Get(pk)
	assert.True(t, ok)
	assert.NotNil(t, val)

	recoveredPid, recoveredSig, err := peerSigHandler.GetPIDAndSig(val)
	assert.Nil(t, err)
	assert.Equal(t, newPid, recoveredPid)
	assert.Equal(t, newSig, recoveredSig)
}

func TestPeerSignatureHandler_VerifyPeerSignatureDifferentPid(t *testing.T) {
	t.Parallel()

	pk := []byte("public key")
	verifyCalled := false
	pid := core.PeerID("dummy peer")
	sig := []byte("signature")
	newPid := core.PeerID("new dummy peer")

	cache := testscommon.NewCacherMock()
	keyGen := &cryptoMocks.KeyGenStub{
		PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
			return &cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return pk, nil
				},
			}, nil
		},
	}
	singleSigner := &cryptoMocks.SingleSignerStub{
		VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			verifyCalled = true
			return nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		keyGen,
	)

	cacheEntry := peerSigHandler.GetCacheEntry(pid, sig)
	cache.Put(pk, cacheEntry, len(pid)+len(sig))

	err := peerSigHandler.VerifyPeerSignature(pk, newPid, sig)
	assert.Equal(t, errorsErd.ErrPIDMismatch, err)
	assert.False(t, verifyCalled)
}

func TestPeerSignatureHandler_VerifyPeerSignatureDifferentSig(t *testing.T) {
	t.Parallel()

	pk := []byte("public key")
	verifyCalled := false
	pid := core.PeerID("dummy peer")
	sig := []byte("signature")
	newSig := []byte("new signature")

	cache := testscommon.NewCacherMock()
	keyGen := &cryptoMocks.KeyGenStub{
		PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
			return &cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return pk, nil
				},
			}, nil
		},
	}
	singleSigner := &cryptoMocks.SingleSignerStub{
		VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			verifyCalled = true
			return nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		keyGen,
	)

	cacheEntry := peerSigHandler.GetCacheEntry(pid, sig)
	cache.Put(pk, cacheEntry, len(pid)+len(sig))

	err := peerSigHandler.VerifyPeerSignature(pk, pid, newSig)
	assert.Equal(t, errorsErd.ErrSignatureMismatch, err)
	assert.False(t, verifyCalled)
}

func TestPeerSignatureHandler_VerifyPeerSignatureGetFromCache(t *testing.T) {
	t.Parallel()

	pk := []byte("public key")
	verifyCalled := false
	pid := core.PeerID("dummy peer")
	sig := []byte("signature")

	cache := testscommon.NewCacherMock()
	keyGen := &cryptoMocks.KeyGenStub{
		PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
			return &cryptoMocks.PublicKeyStub{
				ToByteArrayStub: func() ([]byte, error) {
					return pk, nil
				},
			}, nil
		},
	}
	singleSigner := &cryptoMocks.SingleSignerStub{
		VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			verifyCalled = true
			return nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		keyGen,
	)

	cacheEntry := peerSigHandler.GetCacheEntry(pid, sig)
	cache.Put(pk, cacheEntry, len(pid)+len(sig))

	err := peerSigHandler.VerifyPeerSignature(pk, pid, sig)
	assert.Nil(t, err)
	assert.False(t, verifyCalled)
}

func TestPeerSignatureHandler_GetPeerSignatureErrInConvertingPrivateKeyToByteArray(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("converting error")
	privateKey := &cryptoMocks.PrivateKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return nil, expectedErr
		},
	}
	pid := []byte("dummy peer")

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		testscommon.NewCacherMock(),
		&cryptoMocks.SingleSignerStub{},
		&cryptoMocks.KeyGenStub{},
	)

	sig, err := peerSigHandler.GetPeerSignature(privateKey, pid)
	assert.Nil(t, sig)
	assert.Equal(t, expectedErr, err)
}

func TestPeerSignatureHandler_GetPeerSignatureNotPresentInCache(t *testing.T) {
	t.Parallel()

	privateKeyBytes := []byte("private key")
	privateKey := &cryptoMocks.PrivateKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return privateKeyBytes, nil
		},
	}
	signCalled := false
	pid := []byte("dummy peer")
	sig := []byte("signature")

	cache := testscommon.NewCacherMock()
	singleSigner := &cryptoMocks.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			signCalled = true
			return sig, nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		&cryptoMocks.KeyGenStub{},
	)

	recoveredSig, err := peerSigHandler.GetPeerSignature(privateKey, pid)
	assert.Equal(t, recoveredSig, sig)
	assert.Nil(t, err)
	assert.True(t, signCalled)

	val, ok := cache.Get(privateKeyBytes)
	assert.True(t, ok)
	assert.NotNil(t, val)

	recoveredPid, recoveredSig, err := peerSigHandler.GetPIDAndSig(val)
	assert.Nil(t, err)
	assert.Equal(t, recoveredPid, core.PeerID(pid))
	assert.Equal(t, recoveredSig, sig)
}

func TestPeerSignatureHandler_GetPeerSignatureWrongEntryInCache(t *testing.T) {
	t.Parallel()

	privateKeyBytes := []byte("private key")
	privateKey := &cryptoMocks.PrivateKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return privateKeyBytes, nil
		},
	}
	signCalled := false
	pid := []byte("dummy peer")
	sig := []byte("signature")
	wrongEntry := []byte("wrong entry")

	cache := testscommon.NewCacherMock()
	singleSigner := &cryptoMocks.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			signCalled = true
			return sig, nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		&cryptoMocks.KeyGenStub{},
	)

	cache.Put(privateKeyBytes, wrongEntry, len(wrongEntry))

	recoveredSig, err := peerSigHandler.GetPeerSignature(privateKey, pid)
	assert.Equal(t, recoveredSig, sig)
	assert.Nil(t, err)
	assert.True(t, signCalled)

	val, ok := cache.Get(privateKeyBytes)
	assert.True(t, ok)
	assert.NotNil(t, val)

	recoveredPid, recoveredSig, err := peerSigHandler.GetPIDAndSig(val)
	assert.Nil(t, err)
	assert.Equal(t, recoveredPid, core.PeerID(pid))
	assert.Equal(t, recoveredSig, sig)
}

func TestPeerSignatureHandler_GetPeerSignatureDifferentPidInCache(t *testing.T) {
	t.Parallel()

	privateKeyBytes := []byte("private key")
	privateKey := &cryptoMocks.PrivateKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return privateKeyBytes, nil
		},
	}
	signCalled := false
	pid := core.PeerID("dummy peer")
	newPid := []byte("new dummy peer")
	sig := []byte("signature")
	newSig := []byte("new signature")

	cache := testscommon.NewCacherMock()
	singleSigner := &cryptoMocks.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			signCalled = true
			return newSig, nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		&cryptoMocks.KeyGenStub{},
	)

	entry := peerSigHandler.GetCacheEntry(pid, sig)
	cache.Put(privateKeyBytes, entry, len(pid)+len(sig))

	recoveredSig, err := peerSigHandler.GetPeerSignature(privateKey, newPid)
	assert.Equal(t, recoveredSig, newSig)
	assert.Nil(t, err)
	assert.True(t, signCalled)

	val, ok := cache.Get(privateKeyBytes)
	assert.True(t, ok)
	assert.NotNil(t, val)

	recoveredPid, recoveredSig, err := peerSigHandler.GetPIDAndSig(val)
	assert.Nil(t, err)
	assert.Equal(t, recoveredPid, core.PeerID(newPid))
	assert.Equal(t, recoveredSig, newSig)
}

func TestPeerSignatureHandler_GetPeerSignatureGetFromCache(t *testing.T) {
	t.Parallel()

	privateKeyBytes := []byte("private key")
	privateKey := &cryptoMocks.PrivateKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return privateKeyBytes, nil
		},
	}
	pid := []byte("dummy peer")
	sig := []byte("signature")

	cache := testscommon.NewCacherMock()
	singleSigner := &cryptoMocks.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return nil, nil
		},
	}

	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(
		cache,
		singleSigner,
		&cryptoMocks.KeyGenStub{},
	)

	entry := peerSigHandler.GetCacheEntry(core.PeerID(pid), sig)
	cache.Put(privateKeyBytes, entry, len(pid)+len(sig))

	recoveredSig, err := peerSigHandler.GetPeerSignature(privateKey, pid)
	assert.Equal(t, recoveredSig, sig)
	assert.Nil(t, err)
}
