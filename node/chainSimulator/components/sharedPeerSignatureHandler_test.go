package components

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
)

func TestSharedPeerSignatureVerifier_VerifiesUniqueTupleOnce(t *testing.T) {
	t.Parallel()

	var verifyCalls atomic.Uint32
	delegate := &cryptoMocks.PeerSignatureHandlerStub{
		VerifyPeerSignatureCalled: func(_ []byte, _ core.PeerID, _ []byte) error {
			verifyCalls.Add(1)
			return nil
		},
	}
	verifier := newSharedPeerSignatureVerifier()
	first := verifier.wrap(delegate)
	second := verifier.wrap(delegate)

	wg := sync.WaitGroup{}
	for idx := 0; idx < 50; idx++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			require.NoError(t, first.VerifyPeerSignature([]byte("key"), "peer", []byte("signature")))
		}()
		go func() {
			defer wg.Done()
			require.NoError(t, second.VerifyPeerSignature([]byte("key"), "peer", []byte("signature")))
		}()
	}
	wg.Wait()

	require.Equal(t, uint32(1), verifyCalls.Load())

	require.NoError(t, first.VerifyPeerSignature([]byte("other-key"), "peer", []byte("signature")))
	require.Equal(t, uint32(2), verifyCalls.Load())
}
