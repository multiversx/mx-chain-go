package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/pubkeyConverter"
)

const hex = "hex"
const bech32 = "bech32"

// NewPubkeyConverter will create a new pubkey converter based on the config provided
func NewPubkeyConverter(config config.PubkeyConfig) (state.PubkeyConverter, error) {
	switch config.Type {
	case hex:
		return pubkeyConverter.NewHexPubkeyConverter(config.Length)
	case bech32:
		return pubkeyConverter.NewBech32PubkeyConverter(config.Length)
	default:
		return nil, fmt.Errorf("%w unrecognized type %s", state.ErrInvalidPubkeyConverterType, config.Type)
	}
}
