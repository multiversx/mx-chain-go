package trie

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

func shouldTestNode(n node, key []byte) bool {
	hasher := n.getHasher()
	randomness := string(key) + core.GetAnonymizedMachineID("") + fmt.Sprintf("%d", time.Now().UnixNano())
	buff := hasher.Compute(randomness)
	checkVal := binary.BigEndian.Uint32(buff)
	if checkVal%1000 == 0 {
		log.Debug("deliberately not saving hash", "hash", key)
		return true
	}

	return false
}
