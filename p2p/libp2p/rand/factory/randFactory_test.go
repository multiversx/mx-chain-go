package factory_test

import (
	"crypto/rand"
	"reflect"
	"testing"

	rand2 "github.com/ElrondNetwork/elrond-go/p2p/libp2p/rand"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/rand/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewRandFactory_EmptySeedShouldReturnCryptoRand(t *testing.T) {
	t.Parallel()

	r, err := factory.NewRandFactory("")

	assert.Nil(t, err)
	assert.True(t, r == rand.Reader)
}

func TestNewRandFactory_NotEmptySeedShouldSeedRandReader(t *testing.T) {
	t.Parallel()

	seed := "seed"
	srrExpected, _ := rand2.NewSeedRandReader([]byte(seed))

	r, err := factory.NewRandFactory(seed)

	assert.Nil(t, err)
	assert.Equal(t, reflect.TypeOf(r), reflect.TypeOf(srrExpected))
}
