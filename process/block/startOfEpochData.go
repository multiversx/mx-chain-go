package block

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type startOfEpochData struct {
	peerMiniblocks [][]byte
	headerHash     []byte
	metaHeader     data.HeaderHandler
	stillMissing   int
}
