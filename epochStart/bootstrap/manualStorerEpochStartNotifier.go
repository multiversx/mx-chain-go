package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

type manualStorerEpochStartNotifier struct {
	dataRetriever.ManualEpochStartNotifier
}

// NewManualStorerEpochStartNotifier creates an extended manualEpochStartNotifier instance that can be used by
// the storer service factory
func NewManualStorerEpochStartNotifier(
	manualEpochStartNotifier dataRetriever.ManualEpochStartNotifier,
) (*manualStorerEpochStartNotifier, error) {
	if check.IfNil(manualEpochStartNotifier) {
		return nil, dataRetriever.ErrNilManualEpochStartNotifier
	}

	return &manualStorerEpochStartNotifier{
		ManualEpochStartNotifier: manualEpochStartNotifier,
	}, nil
}

// RegisterHandler does nothing, it is here just to implement the storage.EpochStartNotifier
func (msesn *manualStorerEpochStartNotifier) RegisterHandler(_ epochStart.ActionHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (msesn *manualStorerEpochStartNotifier) IsInterfaceNil() bool {
	return msesn == nil || msesn.ManualEpochStartNotifier == nil
}
