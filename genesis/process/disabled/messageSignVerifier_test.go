package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/stretchr/testify/assert"
)

func TestMessageSignVerifier_Verify(t *testing.T) {
	t.Parallel()

	suite := mcl.NewSuiteBLS12()
	keygen := signing.NewKeyGenerator(suite)

	sv, _ := NewMessageSignVerifier(keygen)

	err := sv.Verify(nil, nil, make([]byte, 97))
	assert.NotNil(t, err)
}
