package common_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/require"
)

func TestPeerAuthenticationPublicKeyIdentifier(t *testing.T) {
	t.Parallel()

	t.Run("short public key should be copied unchanged", func(t *testing.T) {
		t.Parallel()

		publicKey := []byte("public key")

		identifier := common.PeerAuthenticationPublicKeyIdentifier(publicKey)

		require.Equal(t, publicKey, identifier)
		require.NotSame(t, &publicKey[0], &identifier[0])
	})
	t.Run("long public key should be trimmed and copied", func(t *testing.T) {
		t.Parallel()

		publicKey := bytes.Repeat([]byte("p"), common.MaxPeerAuthenticationPublicKeyIdentifierLen+8)

		identifier := common.PeerAuthenticationPublicKeyIdentifier(publicKey)

		require.Equal(t, publicKey[:common.MaxPeerAuthenticationPublicKeyIdentifierLen], identifier)
		require.NotSame(t, &publicKey[0], &identifier[0])
	})
}
