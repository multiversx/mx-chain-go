package factory

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
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
