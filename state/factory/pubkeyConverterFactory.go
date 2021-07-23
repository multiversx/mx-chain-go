package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/state"
)

var log = logger.GetOrCreate("state/factory")

// HexFormat defines the hex format for the pubkey converter
const HexFormat = "hex"

// Bech32Format defines the bech32 format for the pubkey converter
const Bech32Format = "bech32"

// NewPubkeyConverter will create a new pubkey converter based on the config provided
func NewPubkeyConverter(config config.PubkeyConfig) (core.PubkeyConverter, error) {
	switch config.Type {
	case HexFormat:
		return pubkeyConverter.NewHexPubkeyConverter(config.Length)
	case Bech32Format:
		return pubkeyConverter.NewBech32PubkeyConverter(config.Length, log)
	default:
		return nil, fmt.Errorf("%w unrecognized type %s", state.ErrInvalidPubkeyConverterType, config.Type)
	}
}
