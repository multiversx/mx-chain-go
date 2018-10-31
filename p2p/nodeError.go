package p2p

// NodeError is used inside p2p package and contains the receiver peer, sender peer and (if any) other nested errors
type NodeError struct {
	PeerRecv     string
	PeerSend     string
	Err          string
	NestedErrors []NodeError
}

func (e *NodeError) Error() string {
	return e.Err
}

// NewNodeError returns a new instance of the struct NodeError
func NewNodeError(peerRecv string, peerSend string, err string) *NodeError {
	return &NodeError{PeerRecv: peerRecv, PeerSend: peerSend, Err: err}
}
