package p2p

import (
	"github.com/pkg/errors"
)

// ErrNilContext signals that a nil context was provided
var ErrNilContext = errors.New("nil context")

// ErrInvalidPort signals that an invalid port was provided
var ErrInvalidPort = errors.New("invalid port provided")

// ErrNilP2PprivateKey signals that a nil P2P private key has been provided
var ErrNilP2PprivateKey = errors.New("nil P2P private key")

// ErrNilMockNet signals that a nil mocknet was provided. Should occur only in testing!!!
var ErrNilMockNet = errors.New("nil mocknet provided")

// ErrNilTopic signals that a nil topic has been provided
var ErrNilTopic = errors.New("nil topic")

// ErrTopicAlreadyExists signals that a topic already exists
var ErrTopicAlreadyExists = errors.New("topic already exists")

// ErrTopicValidatorOperationNotSupported signals that an unsupported validator operation occurred
var ErrTopicValidatorOperationNotSupported = errors.New("topic validator operation is not supported")

// ErrNilDiscoverer signals that a nil discoverer object has been provided
var ErrNilDiscoverer = errors.New("nil discoverer object")

// ErrNilPipeLoadBalancer signals that a nil data throttler object has been provided
var ErrNilPipeLoadBalancer = errors.New("nil pipe load balancer object")

// ErrPipeAlreadyExists signals that the pipe is already defined (and used)
var ErrPipeAlreadyExists = errors.New("pipe already exists")

// ErrPipeDoNotExists signals that a requested pipe does not exists
var ErrPipeDoNotExists = errors.New("pipe does not exists")

// ErrPipeCanNotBeDeleted signals that a pipe can not be deleted (might be the default pipe)
var ErrPipeCanNotBeDeleted = errors.New("pipe can not be deleted")

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrEmptyTopicList signals that a message with empty topic ids has been received
var ErrEmptyTopicList = errors.New("empty topicIDs")

// ErrAlreadySeenMessage signals that the message has already been seen
var ErrAlreadySeenMessage = errors.New("already seen this message")

// ErrNilDirectSendMessageHandler signals that the message handler for new message has not been wired
var ErrNilDirectSendMessageHandler = errors.New("nil direct sender message handler")

// ErrPeerNotDirectlyConnected signals that the peer is not directly connected to self
var ErrPeerNotDirectlyConnected = errors.New("peer is not directly connected")

// ErrNilHost signals that a nil host has been provided
var ErrNilHost = errors.New("nil host")

// ErrNilValidator signals that a validator hasn't been set for the required topic
var ErrNilValidator = errors.New("no validator has been set for this topic")

// ErrPeerDiscoveryProcessAlreadyStarted signals that mdns peer discovery is already turned on
var ErrPeerDiscoveryProcessAlreadyStarted = errors.New("mdns peer discovery is already turned enabled")

// ErrPeerDiscoveryNotImplemented signals that peer discovery is not implemented
var ErrPeerDiscoveryNotImplemented = errors.New("unimplemented peer discovery")
