package metachain

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
)

type sovereignTrigger struct {
	*trigger
}

// NewSovereignTrigger creates a new sovereign epoch start trigger
func NewSovereignTrigger(args *ArgsNewMetaEpochStartTrigger) (*sovereignTrigger, error) {
	metaTrigger, err := newTrigger(args, &block.SovereignChainHeader{}, &sovereignTriggerRegistryCreator{}, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return &sovereignTrigger{
		trigger: metaTrigger,
	}, nil
}

// SetProcessed sets start of epoch to false and cleans underlying structure
func (st *sovereignTrigger) SetProcessed(header data.HeaderHandler, body data.BodyHandler) {
	st.mutTrigger.Lock()
	defer st.mutTrigger.Unlock()

	sovChainHeader, ok := header.(*block.SovereignChainHeader)
	if !ok {
		log.Error("sovereignTrigger.trigger", "error", data.ErrInvalidTypeAssertion)
		return
	}

	st.baseSetProcessed(sovChainHeader, body)
}

// RevertStateToBlock will revert the state of the trigger to the current block
func (st *sovereignTrigger) RevertStateToBlock(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return epochStart.ErrNilHeaderHandler
	}

	if header.IsStartOfEpochBlock() {
		log.Debug("RevertStateToBlock with epoch start block called")
		st.SetProcessed(header, nil)
		return nil
	}

	st.mutTrigger.RLock()
	prevMeta := st.epochStartMeta
	st.mutTrigger.RUnlock()

	currentHeaderHash, err := core.CalculateHash(st.marshaller, st.hasher, header)
	if err != nil {
		log.Warn("RevertStateToBlock error on hashing", "error", err)
		return err
	}

	if !bytes.Equal(prevMeta.GetPrevHash(), currentHeaderHash) {
		return nil
	}

	log.Debug("RevertStateToBlock to revert behind epoch start block is called")
	err = st.revert(prevMeta)
	if err != nil {
		return err
	}

	st.mutTrigger.Lock()
	st.currentRound = header.GetRound()
	st.mutTrigger.Unlock()

	return nil
}

func (st *sovereignTrigger) revert(header data.HeaderHandler) error {
	if check.IfNil(header) || !header.IsStartOfEpochBlock() || header.GetEpoch() == 0 {
		return nil
	}

	sovMetaHdr, ok := header.(*block.SovereignChainHeader)
	if !ok {
		log.Warn("wrong type assertion in Revert sovereign metachain trigger")
		return epochStart.ErrWrongTypeAssertion
	}

	st.mutTrigger.Lock()
	defer st.mutTrigger.Unlock()

	prevEpochStartIdentifier := core.EpochStartIdentifier(sovMetaHdr.GetEpoch() - 1)
	epochStartMetaBuff, err := st.metaHeaderStorage.SearchFirst([]byte(prevEpochStartIdentifier))
	if err != nil {
		log.Warn("Revert get previous sovereign meta from storage", "error", err)
		return err
	}

	epochStartSovMeta := &block.SovereignChainHeader{}
	err = st.marshaller.Unmarshal(epochStartSovMeta, epochStartMetaBuff)
	if err != nil {
		log.Warn("Revert unmarshal previous sovereign meta", "error", err)
		return err
	}

	st.baeRevert(epochStartSovMeta, sovMetaHdr)
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (st *sovereignTrigger) IsInterfaceNil() bool {
	return st == nil
}
