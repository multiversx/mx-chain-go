package crypto

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewIdentityGenerator(t *testing.T) {
	t.Parallel()

	generator := NewIdentityGenerator()
	assert.False(t, check.IfNil(generator))
}

func TestIdentityGenerator_CreateP2PPrivateKey(t *testing.T) {
	t.Parallel()

	generator := NewIdentityGenerator()
	seed1 := "secret seed 1"
	seed2 := "secret seed 2"
	t.Run("same seed should produce the same private key", func(t *testing.T) {

		sk1, err := generator.CreateP2PPrivateKey(seed1)
		assert.Nil(t, err)

		sk2, err := generator.CreateP2PPrivateKey(seed1)
		assert.Nil(t, err)

		assert.Equal(t, sk1, sk2)
	})
	t.Run("different seed should produce different private key", func(t *testing.T) {
		sk1, err := generator.CreateP2PPrivateKey(seed1)
		assert.Nil(t, err)

		sk2, err := generator.CreateP2PPrivateKey(seed2)
		assert.Nil(t, err)

		assert.NotEqual(t, sk1, sk2)
	})
	t.Run("empty seed should produce different private key", func(t *testing.T) {
		sk1, err := generator.CreateP2PPrivateKey("")
		assert.Nil(t, err)

		sk2, err := generator.CreateP2PPrivateKey("")
		assert.Nil(t, err)

		assert.NotEqual(t, sk1, sk2)
	})
}

func TestIdentityGenerator_CreateRandomP2PIdentity(t *testing.T) {
	t.Parallel()

	generator := NewIdentityGenerator()
	sk1, pid1, err := generator.CreateRandomP2PIdentity()
	assert.Nil(t, err)

	sk2, pid2, err := generator.CreateRandomP2PIdentity()
	assert.Nil(t, err)

	assert.NotEqual(t, sk1, sk2)
	assert.NotEqual(t, pid1, pid2)
	assert.Equal(t, 36, len(sk1))
	assert.Equal(t, 39, len(pid1))
	assert.Equal(t, 36, len(sk2))
	assert.Equal(t, 39, len(pid2))
}
