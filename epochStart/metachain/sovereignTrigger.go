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

// ArgsSovereignTrigger defines args needed to create a sovereign trigger
type ArgsSovereignTrigger struct {
	*ArgsNewMetaEpochStartTrigger
	ValidatorInfoSyncer process.ValidatorInfoSyncer
}

type sovereignTrigger struct {
	*trigger
	currentEpochValidatorInfoPool epochStart.ValidatorInfoCacher
	validatorInfoSyncer           process.ValidatorInfoSyncer
}

// NewSovereignTrigger creates a new sovereign epoch start trigger
func NewSovereignTrigger(args ArgsSovereignTrigger) (*sovereignTrigger, error) {
	metaTrigger, err := newTrigger(args.ArgsNewMetaEpochStartTrigger, &block.SovereignChainHeader{}, &sovereignTriggerRegistryCreator{}, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	if check.IfNil(args.ValidatorInfoSyncer) {
		return nil, epochStart.ErrNilValidatorInfoSyncer
	}
	if check.IfNil(args.DataPool.Headers()) {
		return nil, process.ErrNilHeadersDataPool
	}

	st := &sovereignTrigger{
		trigger:                       metaTrigger,
		currentEpochValidatorInfoPool: args.DataPool.CurrentEpochValidatorInfo(),
		validatorInfoSyncer:           args.ValidatorInfoSyncer,
	}

	args.DataPool.Headers().RegisterHandler(st.receivedBlock)

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

// receivedBlock is a callback function when a new block is received
// upon receiving checks if trigger can be updated to notify subscribed epoch change handlers
func (st *sovereignTrigger) receivedBlock(headerHandler data.HeaderHandler, _ []byte) {
	st.mutTrigger.Lock()
	defer st.mutTrigger.Unlock()

	header, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		return
	}

	if !header.IsStartOfEpochBlock() {
		return
	}

	isMetaStartOfEpochForCurrentEpoch := header.GetEpoch() == st.epoch
	if isMetaStartOfEpochForCurrentEpoch {
		return
	}

	st.updateTrigger(header)
}

func (st *sovereignTrigger) updateTrigger(header data.MetaHeaderHandler) {
	if !st.checkIfTriggerCanBeActivated(header) {
		return
	}

	st.epochStartNotifier.NotifyEpochChangeConfirmed(header.GetEpoch())
}

func (st *sovereignTrigger) checkIfTriggerCanBeActivated(hdr data.HeaderHandler) bool {
	missingMiniBlocksHashes, blockBody, err := st.validatorInfoSyncer.SyncMiniBlocks(hdr)
	if err != nil {
		log.Error("sovereignTrigger.checkIfTriggerCanBeActivated.SyncMiniBlocks", "num missing mini blocks", len(missingMiniBlocksHashes), "error", err)
		return false
	}

	missingValidatorsInfoHashes, validatorsInfo, err := st.validatorInfoSyncer.SyncValidatorsInfo(blockBody)
	if err != nil {
		log.Error("sovereignTrigger.checkIfTriggerCanBeActivated.SyncValidatorsInfo", "num missing validators info", len(missingValidatorsInfoHashes), "error", err)
		return false
	}

	for validatorInfoHash, validatorInfo := range validatorsInfo {
		st.currentEpochValidatorInfoPool.AddValidatorInfo([]byte(validatorInfoHash), validatorInfo)
	}

	st.epochStartNotifier.NotifyAllPrepare(hdr, blockBody)
	return true
}

// IsInterfaceNil checks if the underlying pointer is nil
func (st *sovereignTrigger) IsInterfaceNil() bool {
	return st == nil
}
