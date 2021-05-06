package factory

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// GetSkPk -
func (ccf *cryptoComponentsFactory) GetSkPk() ([]byte, []byte, error) {
	return ccf.getSkPk()
}

// CreateSingleSigner -
func (ccf *cryptoComponentsFactory) CreateSingleSigner(importModeNoSigCheck bool) (crypto.SingleSigner, error) {
	return ccf.createSingleSigner(importModeNoSigCheck)
}

// GetMultiSigHasherFromConfig -
func (ccf *cryptoComponentsFactory) GetMultiSigHasherFromConfig() (hashing.Hasher, error) {
	return ccf.getMultiSigHasherFromConfig()
}

// CreateDummyCryptoParams
func (ccf *cryptoComponentsFactory) CreateDummyCryptoParams() *cryptoParams {
	return &cryptoParams{}
}

// CreateCryptoParams -
func (ccf *cryptoComponentsFactory) CreateCryptoParams(blockSignKeyGen crypto.KeyGenerator) (*cryptoParams, error) {
	return ccf.createCryptoParams(blockSignKeyGen)
}

// CreateMultiSigner -
func (ccf *cryptoComponentsFactory) CreateMultiSigner(
	h hashing.Hasher, cp *cryptoParams, blSignKeyGen crypto.KeyGenerator, importModeNoSigCheck bool,
) (crypto.MultiSigner, error) {
	return ccf.createMultiSigner(h, cp, blSignKeyGen, importModeNoSigCheck)
}

// GetSuite -
func (ccf *cryptoComponentsFactory) GetSuite() (crypto.Suite, error) {
	return ccf.getSuite()
}

// SetListenAddress -
func (ncf *networkComponentsFactory) SetListenAddress(address string) {
	ncf.listenAddress = address
}

// NewBlockProcessor calls the unexported method with the same name in order to use it in tests
func (pcf *processComponentsFactory) NewBlockProcessor(
	requestHandler process.RequestHandler,
	forkDetector process.ForkDetector,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
	txSimulatorProcessorArgs *txsimulator.ArgsTxSimulator,
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler,
) (process.BlockProcessor, error) {
	return pcf.newBlockProcessor(
		requestHandler,
		forkDetector,
		epochStartTrigger,
		bootStorer,
		validatorStatisticsProcessor,
		headerValidator,
		blockTracker,
		pendingMiniBlocksHandler,
		txSimulatorProcessorArgs,
		scheduledTxsExecutionHandler,
	)
}

// SetShardCoordinator -
func SetShardCoordinator(shardCoordinator sharding.Coordinator, holder BootstrapComponentsHolder) {
	mbf := holder.(*managedBootstrapComponents)

	mbf.mutBootstrapComponents.Lock()
	defer mbf.mutBootstrapComponents.Unlock()

	mbf.bootstrapComponents.shardCoordinator = shardCoordinator
}
