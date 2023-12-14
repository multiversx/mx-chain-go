package trie

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

const faultyChance = 1000000

func shouldTestNode(n node, key []byte) bool {
	hasher := n.getHasher()
	randomness := string(key) + core.GetAnonymizedMachineID("") + fmt.Sprintf("%d", time.Now().UnixNano())
	buff := hasher.Compute(randomness)
	checkVal := binary.BigEndian.Uint32(buff)
	if checkVal%faultyChance == 0 {
		log.Debug("deliberately not saving hash", "hash", key)
		return true
	}

	return false
}

func snapshotGetTestPoint(key []byte, faultyChance int) error {
	rand.NewSource(time.Now().UnixNano())
	checkVal := rand.Intn(math.MaxInt)
	if checkVal%faultyChance == 0 {
		log.Debug("deliberately not returning hash", "hash", key)
		return fmt.Errorf("snapshot get error")
	}

	return nil
}
