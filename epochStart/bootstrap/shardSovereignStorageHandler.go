package bootstrap

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
)

type sovereignShardStorageHandler struct {
	*shardStorageHandler
}

// internal constructor, no need to check for nils
func newSovereignShardStorageHandler(shardStorageHandler *shardStorageHandler) *sovereignShardStorageHandler {
	return &sovereignShardStorageHandler{
		shardStorageHandler,
	}
}

// SaveDataToStorage will save the fetched data to storage, so it will be used by the storage bootstrap component
func (ssh *sovereignShardStorageHandler) SaveDataToStorage(components *ComponentsNeededForBootstrap, _ data.HeaderHandler, _ bool, syncedMiniBlocks map[string]*block.MiniBlock) error {
	lastHeader, err := ssh.saveLastHeader(components.ShardHeader)
	if err != nil {
		return err
	}

	err = ssh.saveEpochStartMetaHdrs(components, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return err
	}

	ssh.saveMiniblocksFromComponents(components)

	log.Debug("saving synced miniblocks", "num miniblocks", len(syncedMiniBlocks))
	ssh.saveMiniblocks(syncedMiniBlocks)

	triggerConfigKey, err := ssh.saveTriggerRegistry(components)
	if err != nil {
		return err
	}

	components.NodesConfig.SetCurrentEpoch(components.ShardHeader.GetEpoch())
	nodesCoordinatorConfigKey, err := ssh.saveNodesCoordinatorRegistry(components.EpochStartMetaBlock, components.NodesConfig)
	if err != nil {
		return err
	}

	lastCrossNotarizedHdrs, err := ssh.saveLastCrossChainNotarizedHeaders(components.EpochStartMetaBlock, components.Headers)
	if err != nil {
		return err
	}

	bootStrapData := bootstrapStorage.BootstrapData{
		LastHeader:                 lastHeader,
		LastCrossNotarizedHeaders:  lastCrossNotarizedHdrs,
		LastSelfNotarizedHeaders:   []bootstrapStorage.BootstrapHeaderInfo{lastHeader},
		ProcessedMiniBlocks:        []bootstrapStorage.MiniBlocksInMeta{},
		PendingMiniBlocks:          []bootstrapStorage.PendingMiniBlocksInfo{},
		NodesCoordinatorConfigKey:  nodesCoordinatorConfigKey,
		EpochStartTriggerConfigKey: triggerConfigKey,
		HighestFinalBlockNonce:     lastHeader.Nonce,
		LastRound:                  0,
	}

	return ssh.saveBootStrapData(components, bootStrapData)
}

func (ssh *sovereignShardStorageHandler) saveTriggerRegistry(components *ComponentsNeededForBootstrap) ([]byte, error) {
	sovHeader, castOk := components.EpochStartMetaBlock.(*block.SovereignChainHeader)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignShardStorageHandler.saveTriggerRegistry", process.ErrWrongTypeAssertion)
	}

	metaBlockHash, err := core.CalculateHash(ssh.marshalizer, ssh.hasher, sovHeader)
	if err != nil {
		return nil, err
	}

	triggerReg := block.SovereignShardTriggerRegistry{
		Epoch:                       sovHeader.GetEpoch(),
		CurrentRound:                sovHeader.GetRound(),
		EpochFinalityAttestingRound: sovHeader.GetRound(),
		CurrEpochStartRound:         sovHeader.GetRound(),
		PrevEpochStartRound:         components.PreviousEpochStart.GetRound(),
		EpochStartMetaHash:          metaBlockHash,
		SovereignChainHeader:        sovHeader,
	}

	return ssh.baseSaveTriggerRegistry(&triggerReg, sovHeader.GetRound())
}

func (ssh *sovereignShardStorageHandler) saveLastCrossChainNotarizedHeaders(
	meta data.MetaHeaderHandler,
	headers map[string]data.HeaderHandler,
) ([]bootstrapStorage.BootstrapHeaderInfo, error) {
	log.Debug("saveLastCrossChainNotarizedHeaders")

	shardData, err := getEpochStartShardData(meta, core.MainChainShardId)
	if errors.Is(err, epochStart.ErrEpochStartDataForShardNotFound) {
		log.Error("NO CROSS CHAIN EPOCH START DATA FOUND")
		return []bootstrapStorage.BootstrapHeaderInfo{}, nil
	}

	log.Error("CROSS CHAIN EPOCH START DATA FOUND")

	neededHdr, ok := headers[string(shardData.GetHeaderHash())]
	if !ok {
		return nil, fmt.Errorf("%w in saveLastCrossChainNotarizedHeaders: hash: %s",
			epochStart.ErrMissingHeader,
			hex.EncodeToString(shardData.GetHeaderHash()))
	}

	extendedShardHeader, ok := neededHdr.(*block.ShardHeaderExtended)
	if !ok {
		return nil, fmt.Errorf("%w in sovereignShardStorageHandler.saveLastCrossChainNotarizedHeaders", epochStart.ErrWrongTypeAssertion)
	}

	_, err = ssh.saveExtendedHeaderToStorage(extendedShardHeader)
	if err != nil {
		return nil, err
	}

	crossNotarizedHdrs := make([]bootstrapStorage.BootstrapHeaderInfo, 0)
	crossNotarizedHdrs = append(crossNotarizedHdrs, bootstrapStorage.BootstrapHeaderInfo{
		ShardId: core.MainChainShardId,
		Nonce:   shardData.GetNonce(),
		Hash:    shardData.GetHeaderHash(),
	})

	return crossNotarizedHdrs, nil
}

func (bsh *sovereignShardStorageHandler) saveExtendedHeaderToStorage(extendedShardHeader data.HeaderHandler) ([]byte, error) {
	headerBytes, err := bsh.marshalizer.Marshal(extendedShardHeader)
	if err != nil {
		return nil, err
	}

	headerHash := bsh.hasher.Compute(string(headerBytes))

	metaHdrStorage, err := bsh.storageService.GetStorer(dataRetriever.ExtendedShardHeadersUnit)
	if err != nil {
		return nil, err
	}

	err = metaHdrStorage.Put(headerHash, headerBytes)
	if err != nil {
		return nil, err
	}

	nonceToByteSlice := bsh.uint64Converter.ToByteSlice(extendedShardHeader.GetNonce())
	metaHdrNonceStorage, err := bsh.storageService.GetStorer(dataRetriever.ExtendedShardHeadersNonceHashDataUnit)
	if err != nil {
		return nil, err
	}

	err = metaHdrNonceStorage.Put(nonceToByteSlice, headerHash)
	if err != nil {
		return nil, err
	}

	return headerHash, nil
}
