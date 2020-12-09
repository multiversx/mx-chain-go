package factory

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	IsInterfaceNil() bool
}

// NodesSetupHandler defines which actions should be done for handling initial nodes setup
type NodesSetupHandler interface {
	InitialNodesPubKeys() map[uint32][]string
	InitialEligibleNodesPubKeysForShard(shardId uint32) ([]string, error)
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	SetDebugger(debugger process.AntifloodDebugger) error
	SetPeerValidatorMapper(validatorMapper process.PeerValidatorMapper) error
	SetTopicsForAll(topics ...string)
	ApplyConsensusSize(size int)
	BlacklistPeer(peer core.PeerID, reason string, duration time.Duration)
	IsOriginatorEligibleForTopic(pid core.PeerID, topic string) error
	IsInterfaceNil() bool
}

// EconomicsHandler provides some economics related computation and read access to economics data
type EconomicsHandler interface {
	LeaderPercentage() float64
	ProtocolSustainabilityPercentage() float64
	ProtocolSustainabilityAddress() string
	MinInflationRate() float64
	MaxInflationRate(year uint32) float64
	DeveloperPercentage() float64
	GenesisTotalSupply() *big.Int
	MaxGasLimitPerBlock(shardID uint32) uint64
	ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFee(tx process.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValues(tx process.TransactionWithFeeHandler) error
	MinGasPrice() uint64
	MinGasLimit() uint64
	GasPerDataByte() uint64
	GasPriceModifier() float64
	IsInterfaceNil() bool
}
