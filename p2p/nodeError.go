package p2p

type NodeError struct {
	PeerRecv     string
	PeerSend     string
	Err          string
	NestedErrors []NodeError
}

func (e *NodeError) Error() string {
	return e.Err
}

func NewNodeError(peerRecv string, peerSend string, err string) *NodeError {
	return &NodeError{PeerRecv: peerRecv, PeerSend: peerSend, Err: err}
}
