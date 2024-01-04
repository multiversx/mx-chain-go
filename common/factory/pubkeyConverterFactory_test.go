package factory

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/stretchr/testify/assert"
)

func TestNewPubkeyConverter_HexShouldWork(t *testing.T) {
	t.Parallel()

	pc, err := NewPubkeyConverter(
		config.PubkeyConfig{
			Length: 32,
			Type:   "hex",
		},
	)

	assert.Nil(t, err)
	expected, _ := pubkeyConverter.NewHexPubkeyConverter(32)
	assert.IsType(t, expected, pc)
}

func TestNewPubkeyConverter_Bech32ShouldWork(t *testing.T) {
	t.Parallel()

	pc, err := NewPubkeyConverter(
		config.PubkeyConfig{
			Length: 32,
			Type:   "bech32",
			Hrp:    "erd",
		},
	)

	assert.Nil(t, err)
	expected, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	assert.Nil(t, err)
	assert.IsType(t, expected, pc)
}

func TestNewPubkeyConverter_UnknownTypeShouldErr(t *testing.T) {
	t.Parallel()

	pc, err := NewPubkeyConverter(
		config.PubkeyConfig{
			Length: 32,
			Type:   "unknown",
		},
	)

	assert.Nil(t, pc)
	assert.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}
