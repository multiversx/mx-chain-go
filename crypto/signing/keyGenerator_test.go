package signing

import (
	"testing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
	"github.com/stretchr/testify/assert"
)

func TestNewKeyGenerator_OK(t *testing.T) {
	suite := kv2.NewBlakeSHA256Ed25519()
	kg := NewKeyGenerator(suite)
	assert.NotNil(t, kg)

	s2 := kg.Suite()
	assert.Equal(t, suite, s2)
}

func TestKeyGenerator_GeneratePair(t *testing.T) {

}
