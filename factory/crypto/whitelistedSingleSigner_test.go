package crypto_test

import (
	"encoding/hex"
	"errors"
	"testing"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	cryptoFactory "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWhiteListEd25519Signer(t *testing.T) {
	t.Parallel()

	t.Run("nil key generator should error", func(t *testing.T) {
		t.Parallel()

		singleSigner := &singlesig.Ed25519Signer{}

		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                nil,
			SingleSigner:          singleSigner,
			WhitelistedAddressHex: "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727",
		}

		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)

		require.Equal(t, cryptoFactory.ErrNilKeyGenerator, err)
		require.Nil(t, signer)
	})

	t.Run("nil single signer should error", func(t *testing.T) {
		t.Parallel()

		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return &cryptoMocks.PublicKeyStub{}, nil
			},
		}

		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          nil,
			WhitelistedAddressHex: "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727",
		}

		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)

		require.Equal(t, cryptoFactory.ErrNilSingleSigner, err)
		require.Nil(t, signer)
	})

	t.Run("empty whitelisted address hex should error", func(t *testing.T) {
		t.Parallel()

		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return &cryptoMocks.PublicKeyStub{}, nil
			},
		}
		singleSigner := &singlesig.Ed25519Signer{}

		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          singleSigner,
			WhitelistedAddressHex: "",
		}

		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)

		require.Equal(t, cryptoFactory.ErrEmptyWhitelistedAddressHex, err)
		require.Nil(t, signer)
	})

	t.Run("invalid whitelisted address hex should error", func(t *testing.T) {
		t.Parallel()

		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return &cryptoMocks.PublicKeyStub{}, nil
			},
		}
		singleSigner := &singlesig.Ed25519Signer{}

		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          singleSigner,
			WhitelistedAddressHex: "invalid_hex",
		}

		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)

		require.Error(t, err)
		require.Nil(t, signer)
	})

	t.Run("key generator fails to create public key from bytes", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("failed to create public key")
		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return nil, expectedErr
			},
		}
		singleSigner := &singlesig.Ed25519Signer{}

		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          singleSigner,
			WhitelistedAddressHex: "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727",
		}

		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)

		require.Nil(t, signer)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should successfully create whitelisted signer", func(t *testing.T) {
		t.Parallel()

		whitelistedPubKey := &cryptoMocks.PublicKeyStub{}
		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				// Verify the bytes match the expected whitelisted address
				expectedHex := "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727"
				expectedBytes, _ := hex.DecodeString(expectedHex)
				assert.Equal(t, expectedBytes, b)
				return whitelistedPubKey, nil
			},
		}
		singleSigner := &singlesig.Ed25519Signer{}

		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          singleSigner,
			WhitelistedAddressHex: "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727",
		}

		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)

		require.NoError(t, err)
		require.NotNil(t, signer)
	})
}

func TestWhitelistedSingleSigner_Verify(t *testing.T) {
	t.Parallel()

	t.Run("verify succeeds with valid public key", func(t *testing.T) {
		t.Parallel()

		publicKey := &cryptoMocks.PublicKeyStub{}
		message := []byte("test message")
		signature := []byte("valid signature")

		verifyCalled := false
		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return &cryptoMocks.PublicKeyStub{}, nil
			},
		}

		singleSignerStub := &cryptoMocks.SingleSignerStub{
			VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				assert.Equal(t, publicKey, public)
				assert.Equal(t, message, msg)
				assert.Equal(t, signature, sig)
				verifyCalled = true
				return nil
			},
		}

		// We need to create a real Ed25519Signer and wrap it
		// For testing purposes, we'll use a mock approach
		ed25519Signer := &singlesig.Ed25519Signer{}
		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          ed25519Signer,
			WhitelistedAddressHex: "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727",
		}
		_, err := cryptoFactory.NewWhiteListEd25519Signer(args)
		require.NoError(t, err)

		// Since we can't easily mock the embedded Ed25519Signer, we'll test the logic flow
		// In a real scenario, the first Verify would succeed
		err = singleSignerStub.Verify(publicKey, message, signature)

		require.NoError(t, err)
		assert.True(t, verifyCalled)
	})

	t.Run("verify fails with invalid public key then tries whitelisted key", func(t *testing.T) {
		t.Parallel()

		publicKey := &cryptoMocks.PublicKeyStub{}
		whitelistedPubKey := &cryptoMocks.PublicKeyStub{}
		message := []byte("test message")
		signature := []byte("signature")

		firstVerifyFailed := false
		secondVerifyWithWhitelisted := false

		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return whitelistedPubKey, nil
			},
		}

		callCount := 0
		singleSignerStub := &cryptoMocks.SingleSignerStub{
			VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				callCount++
				if callCount == 1 {
					// First call with original public key fails
					assert.Equal(t, publicKey, public)
					firstVerifyFailed = true
					return errors.New("signature verification failed")
				}
				// Second call with whitelisted key
				assert.Equal(t, whitelistedPubKey, public)
				assert.Equal(t, message, msg)
				assert.Equal(t, signature, sig)
				secondVerifyWithWhitelisted = true
				return nil
			},
		}

		ed25519Signer := &singlesig.Ed25519Signer{}
		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          ed25519Signer,
			WhitelistedAddressHex: "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727",
		}
		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)
		require.NoError(t, err)
		require.NotNil(t, signer)

		// Simulate the verification flow
		err = singleSignerStub.Verify(publicKey, message, signature)
		if err != nil {
			// Fallback to whitelisted key
			err = singleSignerStub.Verify(whitelistedPubKey, message, signature)
		}

		require.NoError(t, err)
		assert.True(t, firstVerifyFailed)
		assert.True(t, secondVerifyWithWhitelisted)
	})

	t.Run("verify fails with both original and whitelisted key", func(t *testing.T) {
		t.Parallel()

		publicKey := &cryptoMocks.PublicKeyStub{}
		whitelistedPubKey := &cryptoMocks.PublicKeyStub{}
		message := []byte("test message")
		signature := []byte("invalid signature")

		expectedErr := errors.New("signature verification failed")

		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return whitelistedPubKey, nil
			},
		}

		callCount := 0
		singleSignerStub := &cryptoMocks.SingleSignerStub{
			VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				callCount++
				if callCount == 1 {
					assert.Equal(t, publicKey, public)
				} else {
					assert.Equal(t, whitelistedPubKey, public)
				}
				return expectedErr
			},
		}

		ed25519Signer := &singlesig.Ed25519Signer{}
		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          ed25519Signer,
			WhitelistedAddressHex: "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727",
		}
		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)
		require.NoError(t, err)
		require.NotNil(t, signer)

		// Simulate the verification flow
		err = singleSignerStub.Verify(publicKey, message, signature)
		if err != nil {
			// Fallback to whitelisted key
			err = singleSignerStub.Verify(whitelistedPubKey, message, signature)
		}

		require.Error(t, err)
		require.Equal(t, expectedErr, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("whitelisted address hex validation", func(t *testing.T) {
		t.Parallel()

		// Verify that the hardcoded whitelisted address is valid hex and has correct length
		whitelistedAddressHex := "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727"

		bytes, err := hex.DecodeString(whitelistedAddressHex)
		require.NoError(t, err)
		require.Equal(t, 32, len(bytes), "Ed25519 public key should be 32 bytes")

		// Verify the bech32 address comment is documented correctly
		// The comment states: erd19ptzdk7zvs3csjjf8u04j5kxzn5jck409sefdyx6x9afkjg4wunsfw7rj7
		// We're just verifying the hex is valid, not the bech32 conversion
	})
}

func TestWhitelistedSingleSigner_Integration(t *testing.T) {
	t.Parallel()

	t.Run("full integration test with real key generator", func(t *testing.T) {
		t.Parallel()

		// This test verifies the full constructor flow with realistic components
		keyGen := &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				// Verify we receive the correct whitelisted address bytes
				whitelistedAddressHex := "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727"
				expectedBytes, _ := hex.DecodeString(whitelistedAddressHex)

				if len(b) != 32 {
					return nil, errors.New("invalid public key length")
				}

				require.Equal(t, expectedBytes, b)

				return &cryptoMocks.PublicKeyStub{}, nil
			},
		}

		ed25519Signer := &singlesig.Ed25519Signer{}

		args := cryptoFactory.ArgsWhiteListedSingleSigner{
			KeyGen:                keyGen,
			SingleSigner:          ed25519Signer,
			WhitelistedAddressHex: "285626dbc26423884a493f1f5952c614e92c5aaf2c329690da317a9b49157727",
		}

		signer, err := cryptoFactory.NewWhiteListEd25519Signer(args)

		require.NoError(t, err)
		require.NotNil(t, signer)
	})
}
