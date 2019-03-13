package sharding

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

var log = logger.NewDefaultLogger()

// InitialNode holds data from json and decoded data from genesis process
type InitialNode struct {
	PubKey  string `json:"pubkey"`
	Balance string `json:"balance"`
	shard   uint32
	pubKey  []byte
	balance *big.Int
}

// Genesis hold data for decoded data from json file
type Genesis struct {
	StartTime          int64          `json:"startTime"`
	RoundDuration      uint64         `json:"roundDuration"`
	ConsensusGroupSize uint32         `json:"consensusGroupSize"`
	MinNodesPerShard   uint32         `json:"minNodesPerShard"`
	ElasticSubrounds   bool           `json:"elasticSubrounds"`
	InitialNodes       []*InitialNode `json:"initialNodes"`
	nrOfShards         uint32
	nrOfNodes          uint32
	allNodesPubKeys    [][]string
}

// NewGenesisConfig creates a new decoded genesis structure from json config file
func NewGenesisConfig(genesisFilePath string) (*Genesis, error) {
	genesis := &Genesis{}

	err := core.LoadJsonFile(genesis, genesisFilePath, log)
	if err != nil {
		return nil, err
	}

	err = genesis.processConfig()
	if err != nil {
		return nil, err
	}

	genesis.processShardAssignment()
	genesis.createInitialNodesPubKeys()

	return genesis, nil
}

func (g *Genesis) processConfig() error {
	var err error
	var ok bool

	g.nrOfNodes = 0
	for i := 0; i < len(g.InitialNodes); i++ {
		g.InitialNodes[i].pubKey, err = hex.DecodeString(g.InitialNodes[i].PubKey)

		// decoder treats empty string as correct, it is not allowed to have empty string as public key
		if g.InitialNodes[i].PubKey == "" || err != nil {
			g.InitialNodes[i].pubKey = nil
			return errors.ErrCouldNotParsePubKey
		}

		g.InitialNodes[i].balance, ok = new(big.Int).SetString(g.InitialNodes[i].Balance, 10)
		if !ok {
			log.Warn(fmt.Sprintf("error decoding balance %s for public key %s - setting to 0",
				g.InitialNodes[i].Balance, g.InitialNodes[i].PubKey))
			g.InitialNodes[i].balance = big.NewInt(0)
		}

		g.nrOfNodes++
	}

	return nil
}

func (g *Genesis) processShardAssignment() {
	// initial verification - should not happen.
	if g.ConsensusGroupSize < 1 {
		g.ConsensusGroupSize = 1
	}
	if g.MinNodesPerShard < g.ConsensusGroupSize {
		g.MinNodesPerShard = g.ConsensusGroupSize
	}

	// initial implementation - as there is no other info than public key, we allocate first nodes in FIFO order to shards
	g.nrOfShards = g.nrOfNodes / g.MinNodesPerShard
	currentShard := uint32(0)
	countSetNodes := uint32(0)
	for ; currentShard < g.nrOfShards; currentShard++ {
		for id := countSetNodes; id < (currentShard+1)*g.MinNodesPerShard; id++ {
			// consider only nodes with valid public key
			if g.InitialNodes[id].pubKey != nil {
				g.InitialNodes[id].shard = currentShard
				countSetNodes++
			}
		}
	}

	// allocate the rest
	currentShard = 0
	for i := countSetNodes; i < g.nrOfNodes; i++ {
		g.InitialNodes[i].shard = currentShard
		currentShard = (currentShard + 1) % g.nrOfShards
	}
}

func (g *Genesis) createInitialNodesPubKeys() {
	g.allNodesPubKeys = make([][]string, g.nrOfShards)
	for _, in := range g.InitialNodes {
		if in.pubKey != nil {
			g.allNodesPubKeys[in.shard] = append(g.allNodesPubKeys[in.shard], string(in.pubKey))
		}
	}
}

// InitialNodesPubKeys - gets initial public keys
func (g *Genesis) InitialNodesPubKeys() [][]string {
	return g.allNodesPubKeys
}

// InitialNodesPubKeysForShard - gets initial public keys
func (g *Genesis) InitialNodesPubKeysForShard(shardId uint32) ([]string, error) {
	if shardId >= g.nrOfShards {
		return nil, errors.ErrShardIdOutOfRange
	}

	if len(g.allNodesPubKeys[shardId]) == 0 {
		return nil, errors.ErrNoPubKeys
	}

	return g.allNodesPubKeys[shardId], nil
}

// InitialNodesBalances - gets the initial balances of the nodes
func (g *Genesis) InitialNodesBalances(shardId uint32) (map[string]*big.Int, error) {
	if shardId >= g.nrOfShards {
		return nil, errors.ErrShardIdOutOfRange
	}

	var balances = make(map[string]*big.Int)
	for _, in := range g.InitialNodes {
		if in.shard == shardId {
			balances[string(in.pubKey)] = in.balance
		}
	}

	if len(balances) == 0 {
		return nil, errors.ErrNoPubKeys
	}

	return balances, nil
}
