package memp2p

import "errors"

// ErrNilNetwork signals that a nil was given where a memp2p.Network instance was expected
var ErrNilNetwork = errors.New("nil network")

// ErrNotConnectedToNetwork signals that a peer tried to perform a network-related operation, but is not connected to any network
var ErrNotConnectedToNetwork = errors.New("not connected to network")

// ErrReceivingPeerNotConnected signals that the receiving peer of a sending operation is not connected to the network
var ErrReceivingPeerNotConnected = errors.New("receiving peer not connected to network")
