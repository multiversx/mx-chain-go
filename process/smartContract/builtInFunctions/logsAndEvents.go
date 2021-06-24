package builtInFunctions

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
)

func newEntryForNFT(identifier string, caller []byte, tokenID []byte, nonce uint64) *vmcommon.LogEntry {
	nonceBig := big.NewInt(0).SetUint64(nonce)

	logEntry := &vmcommon.LogEntry{
		Identifier: []byte(identifier),
		Address:    caller,
		Topics:     [][]byte{tokenID, nonceBig.Bytes()},
	}

	return logEntry
}
