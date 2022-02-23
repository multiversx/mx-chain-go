package epochActivation

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/prometheus/common/log"
)

type epochActivation struct {
	config        config.EnableEpochs
	epoch         uint32
	epochNotifier vm.EpochNotifier
	mutex         sync.RWMutex
}

func (ea *epochActivation) EpochConfirmed(epoch uint32, _ uint64) {
	ea.mutex.Lock()
	defer ea.mutex.Unlock()
	ea.epoch = epoch
}

func NewEpochActivation(config config.EnableEpochs) (*epochActivation, error) {

	log.Debug("esdt: enable epoch for esdt", "epoch", config.ESDTEnableEpoch)
	log.Debug("esdt: enable epoch for contract global mint and burn", "epoch", config.GlobalMintBurnDisableEpoch)
	log.Debug("esdt: enable epoch for contract transfer role", "epoch", config.ESDTTransferRoleEnableEpoch)
	log.Debug("esdt: enable epoch for esdt NFT create on multiple shards", "epoch", config.ESDTNFTCreateOnMultiShardEnableEpoch)
	log.Debug("esdt: enable epoch for meta tokens, financial SFTs", "epoch", config.MetaESDTSetEnableEpoch)
	log.Debug("esdt: enable epoch for transform to multi shard create", "epoch", config.TransformToMultiShardCreateEnableEpoch)
	log.Debug("esdt: enable epoch for esdt register and set all roles function", "epoch", config.ESDTRegisterAndSetAllRolesEnableEpoch)

	return &epochActivation{
		config: config,
		mutex:  sync.RWMutex{},
	}, nil
}

func (ea *epochActivation) IsESDTEnabled() bool {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	return ea.config.ESDTEnableEpoch >= ea.epoch
}

func (ea *epochActivation) IsGlobalMintBurnEnabled() bool {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	return ea.config.GlobalMintBurnDisableEpoch >= ea.epoch
}

func (ea *epochActivation) IsESDTTransferRoleEnabled() bool {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	return ea.config.ESDTTransferRoleEnableEpoch >= ea.epoch
}

func (ea *epochActivation) IsESDTNFTCreateOnMultiShardEnabled() bool {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	return ea.config.ESDTNFTCreateOnMultiShardEnableEpoch >= ea.epoch
}

func (ea *epochActivation) IsMetaESDTSetEnabled() bool {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	return ea.config.MetaESDTSetEnableEpoch >= ea.epoch
}

func (ea *epochActivation) IsTransformToMultiShardCreateEnabled() bool {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	return ea.config.TransformToMultiShardCreateEnableEpoch >= ea.epoch
}

func (ea *epochActivation) IsESDTRegisterAndSetAllRolesEnabled() bool {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()
	return ea.config.ESDTRegisterAndSetAllRolesEnableEpoch >= ea.epoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (ea *epochActivation) IsInterfaceNil() bool {
	return ea == nil
}
