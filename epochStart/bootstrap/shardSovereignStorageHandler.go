package bootstrap

import (
	"fmt"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
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
func (ssh *sovereignShardStorageHandler) SaveDataToStorage(components *ComponentsNeededForBootstrap, notarizedShardHeader data.HeaderHandler, withScheduled bool, syncedMiniBlocks map[string]*block.MiniBlock) error {
	bootStorer, err := ssh.storageService.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return err
	}

	lastHeader, err := ssh.saveLastHeader(components.ShardHeader)
	if err != nil {
		return err
	}

	err = ssh.saveEpochStartMetaHdrs(components)
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

	bootStrapData := bootstrapStorage.BootstrapData{
		LastHeader:                 lastHeader,
		LastCrossNotarizedHeaders:  []bootstrapStorage.BootstrapHeaderInfo{},
		LastSelfNotarizedHeaders:   []bootstrapStorage.BootstrapHeaderInfo{lastHeader},
		ProcessedMiniBlocks:        []bootstrapStorage.MiniBlocksInMeta{},
		PendingMiniBlocks:          []bootstrapStorage.PendingMiniBlocksInfo{},
		NodesCoordinatorConfigKey:  nodesCoordinatorConfigKey,
		EpochStartTriggerConfigKey: triggerConfigKey,
		HighestFinalBlockNonce:     lastHeader.Nonce,
		LastRound:                  0,
	}
	bootStrapDataBytes, err := ssh.marshalizer.Marshal(&bootStrapData)
	if err != nil {
		return err
	}

	roundToUseAsKey := int64(components.ShardHeader.GetRound())
	roundNum := bootstrapStorage.RoundNum{Num: roundToUseAsKey}
	roundNumBytes, err := ssh.marshalizer.Marshal(&roundNum)
	if err != nil {
		return err
	}

	err = bootStorer.Put([]byte(common.HighestRoundFromBootStorage), roundNumBytes)
	if err != nil {
		return err
	}

	log.Info("saved bootstrap data to storage", "round", roundToUseAsKey)
	key := []byte(strconv.FormatInt(roundToUseAsKey, 10))
	err = bootStorer.Put(key, bootStrapDataBytes)
	if err != nil {
		return err
	}

	return nil
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

	bootstrapKey := []byte(fmt.Sprint(sovHeader.GetRound()))
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), bootstrapKey...)

	triggerRegBytes, err := ssh.marshalizer.Marshal(&triggerReg)
	if err != nil {
		return nil, err
	}

	bootstrapStorer, err := ssh.storageService.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return nil, err
	}

	errPut := bootstrapStorer.Put(trigInternalKey, triggerRegBytes)
	if errPut != nil {
		return nil, errPut
	}

	return bootstrapKey, nil
}
