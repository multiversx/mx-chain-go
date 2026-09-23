package common

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/config"
)

var ErrInvalidRecoveryCheckpoint = errors.New("invalid recovery checkpoint")

const minRecoveryChains = 2

type RecoveryCheckpoint struct {
	Round        uint64
	ExcludedEnd  uint64
	headerHashes map[uint32][]byte
}

func NewRecoveryCheckpoint(cfg *config.Config) (*RecoveryCheckpoint, error) {
	if cfg == nil {
		return nil, ErrInvalidRecoveryCheckpoint
	}
	checkpoint := cfg.HardforkRecoveryCheckpoint
	if !checkpoint.Enabled {
		return nil, nil
	}
	if checkpoint.Round == 0 || checkpoint.Round > math.MaxInt64-1 {
		return nil, fmt.Errorf("%w: round %d", ErrInvalidRecoveryCheckpoint, checkpoint.Round)
	}

	var excludedEnd uint64
	for _, interval := range cfg.HardforkRoundExclusions {
		if interval.StartRound <= checkpoint.Round && checkpoint.Round <= interval.EndRound {
			return nil, fmt.Errorf("%w: round %d is excluded", ErrInvalidRecoveryCheckpoint, checkpoint.Round)
		}
		if interval.StartRound == checkpoint.Round+1 {
			excludedEnd = interval.EndRound
		}
	}
	if excludedEnd < checkpoint.Round+1 {
		return nil, fmt.Errorf("%w: no exclusion begins after round %d", ErrInvalidRecoveryCheckpoint, checkpoint.Round)
	}

	hashes := make(map[uint32][]byte, len(checkpoint.Headers))
	for _, entry := range checkpoint.Headers {
		if _, exists := hashes[entry.ShardID]; exists {
			return nil, fmt.Errorf("%w: duplicate shard %d", ErrInvalidRecoveryCheckpoint, entry.ShardID)
		}
		hash, err := hex.DecodeString(entry.Hash)
		if err != nil || len(hash) != HashSize {
			return nil, fmt.Errorf("%w: hash for shard %d", ErrInvalidRecoveryCheckpoint, entry.ShardID)
		}
		hashes[entry.ShardID] = hash
	}
	if len(hashes) < minRecoveryChains || len(hashes[core.MetachainShardId]) == 0 {
		return nil, fmt.Errorf("%w: missing chain hashes", ErrInvalidRecoveryCheckpoint)
	}

	return &RecoveryCheckpoint{
		Round:        checkpoint.Round,
		ExcludedEnd:  excludedEnd,
		headerHashes: hashes,
	}, nil
}

func (rc *RecoveryCheckpoint) HeaderHash(shardID uint32) ([]byte, bool) {
	if rc == nil {
		return nil, false
	}
	hash, ok := rc.headerHashes[shardID]
	if !ok {
		return nil, false
	}

	return append([]byte(nil), hash...), true
}

func (rc *RecoveryCheckpoint) HasAllShards(numShards uint32) bool {
	if rc == nil || len(rc.headerHashes) != int(numShards)+1 {
		return false
	}
	for shardID := uint32(0); shardID < numShards; shardID++ {
		if len(rc.headerHashes[shardID]) == 0 {
			return false
		}
	}

	return len(rc.headerHashes[core.MetachainShardId]) != 0
}

func (rc *RecoveryCheckpoint) IsDiscarded(round uint64, shardID uint32, hash []byte) bool {
	if rc == nil {
		return false
	}
	if round > rc.Round && round <= rc.ExcludedEnd {
		return true
	}
	if round != rc.Round {
		return false
	}

	expected, ok := rc.headerHashes[shardID]
	return !ok || !bytes.Equal(expected, hash)
}
