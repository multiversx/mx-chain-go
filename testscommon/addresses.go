package testscommon

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// RealWorldBech32PubkeyConverter is a bech32 converter, to be used in tests
var RealWorldBech32PubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, logger.GetOrCreate("testscommon"))

var (
	// TestAddressAlice is a test address
	TestAddressAlice = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	// TestPubKeyAlice is a test pubkey
	TestPubKeyAlice, _ = RealWorldBech32PubkeyConverter.Decode(TestAddressAlice)
	// TestPubKeyHexAlice is a test pubkey
	TestPubKeyHexAlice = hex.EncodeToString(TestPubKeyAlice)

	// TestAddressBob is a test address
	TestAddressBob = "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	// TestPubKeyBob is a test pubkey
	TestPubKeyBob, _ = RealWorldBech32PubkeyConverter.Decode(TestAddressBob)
	// TestPubKeyHexBob is a test pubkey
	TestPubKeyHexBob = hex.EncodeToString(TestPubKeyBob)
)
