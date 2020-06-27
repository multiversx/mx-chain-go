package p2p

import (
	"errors"
)

// ErrNilContext signals that a nil context was provided
var ErrNilContext = errors.New("nil context")

// ErrNilMockNet signals that a nil mocknet was provided. Should occur only in testing!!!
var ErrNilMockNet = errors.New("nil mocknet provided")

// ErrNilTopic signals that a nil topic has been provided
var ErrNilTopic = errors.New("nil topic")

// ErrTopicAlreadyExists signals that a topic already exists
var ErrTopicAlreadyExists = errors.New("topic already exists")

// ErrTopicValidatorOperationNotSupported signals that an unsupported validator operation occurred
var ErrTopicValidatorOperationNotSupported = errors.New("topic validator operation is not supported")

// ErrChannelDoesNotExist signals that a requested channel does not exist
var ErrChannelDoesNotExist = errors.New("channel does not exist")

// ErrChannelCanNotBeDeleted signals that a channel can not be deleted (might be the default channel)
var ErrChannelCanNotBeDeleted = errors.New("channel can not be deleted")

// ErrChannelCanNotBeReAdded signals that a channel can not be re added as it is the default channel
var ErrChannelCanNotBeReAdded = errors.New("channel can not be re added")

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrEmptyTopicList signals that a message with empty topic ids has been received
var ErrEmptyTopicList = errors.New("empty topicIDs")

// ErrAlreadySeenMessage signals that the message has already been seen
var ErrAlreadySeenMessage = errors.New("already seen this message")

// ErrOldMessage signals that the message is too old
var ErrOldMessage = errors.New("message too old")

// ErrNilDirectSendMessageHandler signals that the message handler for new message has not been wired
var ErrNilDirectSendMessageHandler = errors.New("nil direct sender message handler")

// ErrPeerNotDirectlyConnected signals that the peer is not directly connected to self
var ErrPeerNotDirectlyConnected = errors.New("peer is not directly connected")

// ErrNilHost signals that a nil host has been provided
var ErrNilHost = errors.New("nil host")

// ErrNilValidator signals that a validator hasn't been set for the required topic
var ErrNilValidator = errors.New("no validator has been set for this topic")

// ErrPeerDiscoveryProcessAlreadyStarted signals that a peer discovery is already turned on
var ErrPeerDiscoveryProcessAlreadyStarted = errors.New("peer discovery is already turned on")

// ErrMessageTooLarge signals that the message provided is too large
var ErrMessageTooLarge = errors.New("buffer too large")

// ErrEmptyBufferToSend signals that an empty buffer was provided for sending to other peers
var ErrEmptyBufferToSend = errors.New("empty buffer to send")

// ErrNilFetchPeersOnTopicHandler signals that a nil handler was provided
var ErrNilFetchPeersOnTopicHandler = errors.New("nil fetch peers on topic handler")

// ErrInvalidDurationProvided signals that an invalid time.Duration has been provided
var ErrInvalidDurationProvided = errors.New("invalid time.Duration provided")

// ErrTooManyGoroutines is raised when the number of goroutines has exceeded a threshold
var ErrTooManyGoroutines = errors.New(" number of goroutines exceeded")

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")

// ErrInvalidPortValue signals that an invalid port value has been provided
var ErrInvalidPortValue = errors.New("invalid port value")

// ErrInvalidPortsRangeString signals that an invalid ports range string has been provided
var ErrInvalidPortsRangeString = errors.New("invalid ports range string")

// ErrInvalidStartingPortValue signals that an invalid starting port value has been provided
var ErrInvalidStartingPortValue = errors.New("invalid starting port value")

// ErrInvalidEndingPortValue signals that an invalid ending port value has been provided
var ErrInvalidEndingPortValue = errors.New("invalid ending port value")

// ErrEndPortIsSmallerThanStartPort signals that the ending port value is smaller than the starting port value
var ErrEndPortIsSmallerThanStartPort = errors.New("ending port value is smaller than the starting port value")

// ErrNoFreePortInRange signals that no free port was found from provided range
var ErrNoFreePortInRange = errors.New("no free port in range")

// ErrNilSharder signals that the provided sharder is nil
var ErrNilSharder = errors.New("nil sharder")

// ErrNilPeerShardResolver signals that the peer shard resolver provided is nil
var ErrNilPeerShardResolver = errors.New("nil PeerShardResolver")

// ErrNilNetworkShardingCollector signals that the network sharding collector provided is nil
var ErrNilNetworkShardingCollector = errors.New("nil network sharding collector")

// ErrNilSignerVerifier signals that the signer-verifier instance provided is nil
var ErrNilSignerVerifier = errors.New("nil signer-verifier")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil marshalizer implementation
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrWrongTypeAssertion signals that a wrong type assertion occurred
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilReconnecter signals that a nil reconnecter has been provided
var ErrNilReconnecter = errors.New("nil reconnecter")

// ErrUnwantedPeer signals that the provided peer has a longer kademlia distance in respect with the already connected
// peers and a connection to this peer will result in an immediate disconnection
var ErrUnwantedPeer = errors.New("unwanted peer: will not initiate connection as it will get disconnected")

// ErrEmptySeed signals that an empty seed has been provided
var ErrEmptySeed = errors.New("empty seed")

// ErrEmptyBuffer signals that an empty buffer has been provided
var ErrEmptyBuffer = errors.New("empty buffer")

// ErrNilPeerDenialEvaluator signals that a nil peer denial evaluator was provided
var ErrNilPeerDenialEvaluator = errors.New("nil peer denial evaluator")

// ErrNilStatusHandler signals that a nil status handler has been provided
var ErrNilStatusHandler = errors.New("nil status handler")

// ErrMessageUnmarshalError signals that an invalid message was received from a peer. There is no way to communicate
// with such a peer as it does not respect the protocol
var ErrMessageUnmarshalError = errors.New("message unmarshal error")
