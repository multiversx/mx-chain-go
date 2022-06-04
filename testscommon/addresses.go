package testscommon

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
)

var RealWorldBech32PubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, &mock.LoggerMock{})

var (
	AddressOfAlice   = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	PubKeyOfAlice, _ = RealWorldBech32PubkeyConverter.Decode(AddressOfAlice)
	PubKeyOfAliceHex = hex.EncodeToString(PubKeyOfAlice)

	AddressOfBob   = "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	PubKeyOfBob, _ = RealWorldBech32PubkeyConverter.Decode(AddressOfBob)
	PubKeyOfBobHex = hex.EncodeToString(PubKeyOfBob)
)
