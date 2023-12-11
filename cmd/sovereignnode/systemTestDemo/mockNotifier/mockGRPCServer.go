package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type ConfirmedBridgeOp struct {
	HashOfHashes []byte
	BridgeOpHash []byte
}

type mockServer struct {
	mut       sync.RWMutex
	cachedOps []*sovereign.BridgeOperations
	*sovereign.UnimplementedBridgeTxSenderServer
}

// NewMockServer -
func NewMockServer() *mockServer {
	return &mockServer{
		cachedOps: make([]*sovereign.BridgeOperations, 0),
	}
}

// Send should handle receiving data bridge operations from sovereign shard and forward transactions to main chain
func (s *mockServer) Send(_ context.Context, data *sovereign.BridgeOperations) (*sovereign.BridgeOperationsResponse, error) {
	s.mut.Lock()
	s.cachedOps = append(s.cachedOps, data)
	s.mut.Unlock()

	hashes, err := generateRandomHashes(data)
	if err != nil {
		return nil, err
	}

	logTxHashes(hashes)

	return &sovereign.BridgeOperationsResponse{
		TxHashes: hashes,
	}, nil
}

func (s *mockServer) ExtractRandomBridgeTopicsForConfirmation() ([]*ConfirmedBridgeOp, error) {
	percent := int64(70)
	ret := make([]*ConfirmedBridgeOp, 0)

	s.mut.Lock()
	cachedOpsCopy := make([]*sovereign.BridgeOperations, len(s.cachedOps))
	copy(cachedOpsCopy, s.cachedOps)
	s.mut.Unlock()

	for _, cachedOp := range cachedOpsCopy {
		for _, outGoingData := range cachedOp.Data {
			for _, outGoingOp := range outGoingData.OutGoingOperations {
				index, err := rand.Int(rand.Reader, big.NewInt(100))
				if err != nil {
					return nil, err
				}
				if index.Int64() > percent {
					ret = append(ret, &ConfirmedBridgeOp{
						HashOfHashes: outGoingData.Hash,
						BridgeOpHash: outGoingOp.Hash,
					})
				}
			}
		}
	}

	return ret, nil
}

func generateRandomHashes(bridgeOps *sovereign.BridgeOperations) ([]string, error) {
	numHashes := len(bridgeOps.Data) + 1 // one register tx  + one tx for each bridge op
	hashes := make([]string, numHashes)

	size := 32
	for i := 0; i < numHashes; i++ {
		randomBytes := make([]byte, size)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return nil, err
		}
		hashes[i] = hex.EncodeToString(randomBytes)[:size]
	}

	return hashes, nil
}

func logTxHashes(hashes []string) {
	for _, hash := range hashes {
		log.Info("sent tx", "hash", hash)
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (s *mockServer) IsInterfaceNil() bool {
	return s == nil
}
