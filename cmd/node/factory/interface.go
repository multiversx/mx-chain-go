package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// HeaderSigVerifierHandler is the interface needed to check that a header's signature is correct
type HeaderSigVerifierHandler interface {
	VerifyRandSeed(header data.HeaderHandler) error
	VerifyLeaderSignature(header data.HeaderHandler) error
	VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error
	VerifySignature(header data.HeaderHandler) error
	IsInterfaceNil() bool
}

// HeaderIntegrityVerifierHandler is the interface needed to check that a header's integrity is correct
type HeaderIntegrityVerifierHandler interface {
	Verify(header data.HeaderHandler) error
	GetVersion(epoch uint32) string
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	IsInterfaceNil() bool
}

// FileLoggingHandler will handle log file rotation
type FileLoggingHandler interface {
	ChangeFileLifeSpan(logger logger.LogLifeSpanner) error
	Close() error
	IsInterfaceNil() bool
}

// TODO: find a better naming
// StatusHandlersUtils provides some functionality for statusHandlers
type StatusHandlersUtils interface {
	StatusHandler() core.AppStatusHandler
	Metrics() external.StatusMetricsHandler
	UpdateStorerAndMetricsForPersistentHandler(store storage.Storer) error
	SignalStartViews()
	SignalLogRewrite()
	IsInterfaceNil() bool
}

// StatusHandlerUtilsFactory is the factory for statusHandler utils
type StatusHandlerUtilsFactory interface {
	Create(marshalizer marshal.Marshalizer, converter typeConverters.Uint64ByteSliceConverter) (StatusHandlersUtils, error)
}
