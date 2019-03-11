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
	Address string `json:"address"`
	PubKey  string `json:"pubkey"`
	Balance string `json:"balance"`
	shard   uint32
	address []byte
	pubKey  []byte
	balance *big.Int
}

// Genesis hold data for decoded data from json file
type Genesis struct {
	StartTime           int64         `json:"startTime"`
	RoundDuration       uint64        `json:"roundDuration"`
	ConsensusGroupSize  int           `json:"consensusGroupSize"`
	ShardConsBufferSize uint32        `json:"shardConsBufferSize"`
	ElasticSubrounds    bool          `json:"elasticSubrounds"`
	InitialNodes        []InitialNode `json:"initialNodes"`
	nrOfShards          uint32
	nrOfNodes           uint32
}

// NewGenesisConfig creates a new decoded genesis structure from json config file
func NewGenesisConfig(genesisFilePath string) (*Genesis, error) {
	if &genesisFilePath == nil {
		return nil, errors.ErrFilePathNil
	}

	genesis := &Genesis{ShardConsBufferSize: 1,
		ConsensusGroupSize: 1}

	err := core.LoadJsonFile(genesis, genesisFilePath, log)
	if err != nil {
		return nil, err
	}

	genesis.processConfig()
	genesis.processShardAssigment()

	return genesis, nil
}

func (g *Genesis) processConfig() {
	var err error
	var ok bool

	g.nrOfNodes = 0
	for _, in := range g.InitialNodes {
		in.pubKey, err = hex.DecodeString(in.PubKey)

		if err != nil {
			log.Error(fmt.Sprintf("%s is not a valid public key. Ignored", in.PubKey))
			continue
		}

		in.balance, ok = new(big.Int).SetString(in.Balance, 10)
		if !ok {
			log.Warn(fmt.Sprintf("error decoding balance %s for public key %s - setting to 0", in.Balance, in.PubKey))
			in.balance = big.NewInt(0)
		}

		g.nrOfNodes++

		if &in.Address != nil {
			in.address, err = hex.DecodeString(in.Address)

			if err != nil {
				log.Error(fmt.Sprintf("%s is not a valid address. Ignored", in.Address))
				continue
			}
		}

	}
}

func (g *Genesis) processShardAssigment() {
	// initial implementation - as there is no other info than public key, we allocate first N nodes to first shard and so on.
	g.nrOfShards = g.nrOfNodes / g.ShardConsBufferSize
	currentShard := uint32(0)
	countSetNodes := uint32(0)
	for currentShard = uint32(0); currentShard < g.nrOfShards; currentShard++ {
		for id := countSetNodes; id < (currentShard+1)*g.ShardConsBufferSize; id++ {
			// consider only nodes with valid public key
			if g.InitialNodes[id].pubKey != nil {
				g.InitialNodes[id].shard = currentShard
				countSetNodes++
			}
		}
	}

	// allocate the rest :D
	for i := countSetNodes; i < g.nrOfNodes; i++ {
		g.InitialNodes[i].shard = currentShard
		currentShard++
		if currentShard > g.nrOfShards {
			currentShard = 0
		}
	}
}

// InitialNodesPubKeys - gets initial public keys
func (g *Genesis) InitialNodesPubKeys() [][]string {
	pubKeys := make([][]string, g.nrOfShards)

	for _, in := range g.InitialNodes {
		pubKeys[in.shard] = append(pubKeys[in.shard], string(in.pubKey))
	}

	return pubKeys
}

// InitialNodesAddress - gets initial addresses keys
func (g *Genesis) InitialNodesAddress() [][]string {
	pubAdrs := make([][]string, g.nrOfShards)

	for _, in := range g.InitialNodes {
		pubAdrs[in.shard] = append(pubAdrs[in.shard], string(in.address))
	}

	return pubAdrs
}

// InitialNodesBalances - gets the initial balances of the nodes
func (g *Genesis) InitialNodesBalances(shardID uint32) map[string]*big.Int {
	var balances = make(map[string]*big.Int)
	for _, in := range g.InitialNodes {
		if in.shard == shardID {
			balances[string(in.pubKey)] = in.balance
		}
	}
	return balances
}
