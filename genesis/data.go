package genesis

import "math/big"

// DelegationData specify the delegation address and the balance provided
type DelegationData struct {
	Address string `json:"address,omitempty"`
	Balance string `json:"balance,omitempty"`
	address []byte
	balance *big.Int
}

// InitialBalance provides information about one entry in the genesis file
type InitialBalance struct {
	Address        string          `json:"address"`
	Supply         string          `json:"supply"`
	Balance        string          `json:"balance"`
	StakingBalance string          `json:"stakingbalance"`
	Delegation     *DelegationData `json:"delegation"`

	address        []byte
	supply         *big.Int
	balance        *big.Int
	stakingBalance *big.Int
}

// InitialNode holds data from json
type InitialNode struct {
	PubKey  string `json:"pubkey"`
	Address string `json:"address"`
}

// NodesSetup holds the data read from the nodes setup file
type NodesSetup struct {
	StartTime                   int64   `json:"startTime"`
	RoundDuration               uint64  `json:"roundDuration"`
	ConsensusGroupSize          uint32  `json:"consensusGroupSize"`
	MinNodesPerShard            uint32  `json:"minNodesPerShard"`
	ChainID                     string  `json:"chainID"`
	MetaChainConsensusGroupSize uint32  `json:"metaChainConsensusGroupSize"`
	MetaChainMinNodes           uint32  `json:"metaChainMinNodes"`
	Hysteresis                  float32 `json:"hysteresis"`
	Adaptivity                  bool    `json:"adaptivity"`
	InitialNodes                []*InitialNode
}
