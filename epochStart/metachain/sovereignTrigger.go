package metachain

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignTrigger struct {
	*trigger
	currentEpochValidatorInfoPool epochStart.ValidatorInfoCacher
	PeerMiniBlocksSyncer          process.ValidatorInfoSyncer
}

// NewSovereignTrigger creates a new sovereign epoch start trigger
func NewSovereignTrigger(args *ArgsNewMetaEpochStartTrigger, PeerMiniBlocksSyncer process.ValidatorInfoSyncer) (*sovereignTrigger, error) {
	metaTrigger, err := newTrigger(args, &block.SovereignChainHeader{}, &sovereignTriggerRegistryCreator{}, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	st := &sovereignTrigger{
		trigger:                       metaTrigger,
		currentEpochValidatorInfoPool: args.DataPool.CurrentEpochValidatorInfo(),
		PeerMiniBlocksSyncer:          PeerMiniBlocksSyncer,
	}

	args.DataPool.Headers().RegisterHandler(st.receivedMetaBlock)

	return st, nil
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
		log.Debug("sovereignTrigger.RevertStateToBlock with epoch start block called")
		st.SetProcessed(header, nil)
		return nil
	}

	st.mutTrigger.RLock()
	prevMeta := st.epochStartMeta
	st.mutTrigger.RUnlock()

	currentHeaderHash, err := core.CalculateHash(st.marshaller, st.hasher, header)
	if err != nil {
		log.Warn("sovereignTrigger.RevertStateToBlock error on hashing", "error", err)
		return err
	}

	if !bytes.Equal(prevMeta.GetPrevHash(), currentHeaderHash) {
		return nil
	}

	log.Debug("sovereignTrigger.RevertStateToBlock to revert behind epoch start block is called")
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

	st.baseRevert(epochStartSovMeta, sovMetaHdr)
	return nil
}

// receivedMetaBlock is a callback function when a new metablock was received
// upon receiving checks if trigger can be updated
func (t *sovereignTrigger) receivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	metaHdr, ok := headerHandler.(*block.SovereignChainHeader)
	if !ok {
		return
	}

	if !metaHdr.IsStartOfEpochBlock() {
		return
	}

	log.Error("##############################receivedMetaBlock", "epoch", headerHandler.GetEpoch())

	isMetaStartOfEpochForCurrentEpoch := metaHdr.GetEpoch() == t.epoch && metaHdr.IsStartOfEpochBlock()
	if isMetaStartOfEpochForCurrentEpoch {
		log.Error("##############################isMetaStartOfEpochForCurrentEpoch")
		return
	}

	var err error
	defer func() {
		log.LogIfError(err)
	}()

	missingMiniBlocksHashes, blockBody, err := t.PeerMiniBlocksSyncer.SyncMiniBlocks(metaHdr)
	if err != nil {
		//	t.addMissingMiniBlocks(metaHdr.GetEpoch(), missingMiniBlocksHashes)
		log.Debug("checkIfTriggerCanBeActivated.SyncMiniBlocks", "num missing mini blocks", len(missingMiniBlocksHashes), "error", err)
		return
	}

	missingValidatorsInfoHashes, validatorsInfo, err := t.PeerMiniBlocksSyncer.SyncValidatorsInfo(blockBody)
	if err != nil {
		//t.addMissingValidatorsInfo(metaHdr.GetEpoch(), missingValidatorsInfoHashes)
		log.Debug("checkIfTriggerCanBeActivated.SyncValidatorsInfo", "num missing validators info", len(missingValidatorsInfoHashes), "error", err)
		return
	}

	for validatorInfoHash, validatorInfo := range validatorsInfo {
		t.currentEpochValidatorInfoPool.AddValidatorInfo([]byte(validatorInfoHash), validatorInfo)
	}

	log.Error("##############################receivedMetaBlock", "NotifyAllPrepare", "true")

	t.epochStartNotifier.NotifyAllPrepare(metaHdr, blockBody)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (st *sovereignTrigger) IsInterfaceNil() bool {
	return st == nil
}
