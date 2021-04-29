package mock

import "github.com/ElrondNetwork/elrond-go/process"

// PeerHeartbeatProcessorStub -
type PeerHeartbeatProcessorStub struct {
	ProcessCalled func(peerHeartbeat process.InterceptedPeerHeartbeat) error
}

// Process -
func (proc *PeerHeartbeatProcessorStub) Process(peerHeartbeat process.InterceptedPeerHeartbeat) error {
	if proc.ProcessCalled != nil {
		return proc.ProcessCalled(peerHeartbeat)
	}

	return nil
}

// IsInterfaceNil -
func (proc *PeerHeartbeatProcessorStub) IsInterfaceNil() bool {
	return proc == nil
}
