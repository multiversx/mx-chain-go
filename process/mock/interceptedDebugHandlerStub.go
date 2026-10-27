package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptedDebugHandlerStub -
type InterceptedDebugHandlerStub struct {
	LogReceivedHashesCalled  func(topic string, hashes [][]byte)
	LogProcessedHashesCalled func(topic string, hashes [][]byte, err error)
	LogReceivedDataCalled    func(data process.InterceptedData, msg p2p.MessageP2P, fromConnectedPeer core.PeerID)
}

// LogReceivedData -
func (idhs *InterceptedDebugHandlerStub) LogReceivedData(data process.InterceptedData, msg p2p.MessageP2P, fromConnectedPeer core.PeerID) {
	if idhs.LogReceivedDataCalled != nil {
		idhs.LogReceivedDataCalled(data, msg, fromConnectedPeer)
	}
}

// LogReceivedHashes -
func (idhs *InterceptedDebugHandlerStub) LogReceivedHashes(topic string, hashes [][]byte) {
	if idhs.LogReceivedHashesCalled != nil {
		idhs.LogReceivedHashesCalled(topic, hashes)
	}
}

// LogProcessedHashes -
func (idhs *InterceptedDebugHandlerStub) LogProcessedHashes(topic string, hashes [][]byte, err error) {
	if idhs.LogProcessedHashesCalled != nil {
		idhs.LogProcessedHashesCalled(topic, hashes, err)
	}
}

// IsInterfaceNil -
func (idhs *InterceptedDebugHandlerStub) IsInterfaceNil() bool {
	return idhs == nil
}
