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

const selectBridgeOpChance = int64(40) // 40% chance

type mockServer struct {
	mut       sync.RWMutex
	cachedOps map[string]map[string]struct{}
	*sovereign.UnimplementedBridgeTxSenderServer
}

// NewMockServer creates a mock grpc bridge server
func NewMockServer() *mockServer {
	return &mockServer{
		cachedOps: make(map[string]map[string]struct{}),
	}
}

// Send will save internally in a cache the received outgoing bridge operations.
// As a response, it will generate random tx hashes, since no bridge tx will actually be sent.
func (s *mockServer) Send(_ context.Context, data *sovereign.BridgeOperations) (*sovereign.BridgeOperationsResponse, error) {
	s.mut.Lock()
	for _, bridgeData := range data.Data {
		bridgeOpHashes := make(map[string]struct{})
		for _, outGoingOp := range bridgeData.OutGoingOperations {
			bridgeOpHashes[string(outGoingOp.Hash)] = struct{}{}
		}

		s.cachedOps[string(bridgeData.Hash)] = bridgeOpHashes
	}
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

// ExtractRandomBridgeTopicsForConfirmation will randomly select (40% chance for an event to be selected) some of the
// internal saved bridge operations and remove them from the cache. These events shall be used by the notifier to
// send confirmedBridgeOperation events
func (s *mockServer) ExtractRandomBridgeTopicsForConfirmation() ([]*ConfirmedBridgeOp, error) {
	ret := make([]*ConfirmedBridgeOp, 0)

	s.mut.Lock()
	defer s.mut.Unlock()

	for hash, cachedOp := range s.cachedOps {
		selectedBridgeOps, err := selectRandomBridgeOps([]byte(hash), cachedOp)
		if err != nil {
			return nil, err
		}

		ret = append(ret, selectedBridgeOps...)
	}

	s.removeBridgeOpsFromCache(ret)

	return ret, nil
}

func selectRandomBridgeOps(hash []byte, outGoingOps map[string]struct{}) ([]*ConfirmedBridgeOp, error) {
	ret := make([]*ConfirmedBridgeOp, 0)
	for outGoingOpHash := range outGoingOps {
		index, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			return nil, err
		}

		if index.Int64() < selectBridgeOpChance {
			ret = append(ret, &ConfirmedBridgeOp{
				HashOfHashes: hash,
				BridgeOpHash: []byte(outGoingOpHash),
			})
		}

	}

	return ret, nil
}

func (s *mockServer) removeBridgeOpsFromCache(bridgeOps []*ConfirmedBridgeOp) {
	for _, bridgeOp := range bridgeOps {
		delete(s.cachedOps[string(bridgeOp.HashOfHashes)], string(bridgeOp.BridgeOpHash))
		if len(s.cachedOps[string(bridgeOp.HashOfHashes)]) == 0 {
			delete(s.cachedOps, string(bridgeOp.HashOfHashes))
		}
	}
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

		hashes[i] = hex.EncodeToString(randomBytes)
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
