package disabled

type peerHonesty struct {
}

// NewPeerHonesty creates a new instance of disabled peer honesty
func NewPeerHonesty() *peerHonesty {
	return &peerHonesty{}
}

// ChangeScore does nothing
func (p *peerHonesty) ChangeScore(_ string, _ string, _ int) {
}

// Close does nothing and returns nil
func (p *peerHonesty) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *peerHonesty) IsInterfaceNil() bool {
	return p == nil
}
