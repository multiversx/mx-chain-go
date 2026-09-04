package factory

import (
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptorDebugHandler hold information about requested and received information
type InterceptorDebugHandler interface {
	EpochStartEventHandler() epochStart.ActionHandler
	LogRequestedData(topic string, hashes [][]byte, numReqIntra int, numReqCross int)
	LogReceivedHashes(topic string, hashes [][]byte)
	LogProcessedHashes(topic string, hashes [][]byte, err error)
	LogFailedToResolveData(topic string, hash []byte, err error)
	LogSucceededToResolveData(topic string, hash []byte)
	LogReceivedData(data process.InterceptedData, msg p2p.MessageP2P, fromConnectedPeer core.PeerID)
	Query(topic string) []string
	Close() error
	IsInterfaceNil() bool
}

// ProcessDebugger defines what a process debugger implementation should do
type ProcessDebugger interface {
	SetLastCommittedBlockRound(round uint64)
	Close() error
	IsInterfaceNil() bool
}
