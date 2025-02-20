package chaos

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type chaosController struct {
	mutex      sync.RWMutex
	enabled    bool
	profile    chaosProfile
	nodeConfig *config.Configs
	node       NodeHandler
}

func newChaosController() *chaosController {
	return &chaosController{enabled: false}
}

// Setup sets up the chaos controller. Make sure to call this only after logging components (file logging, as well) are set up.
func (controller *chaosController) Setup() error {
	config, err := newChaosConfigFromFile(defaultConfigFilePath)
	if err != nil {
		return fmt.Errorf("could not load chaos config: %w", err)
	}

	controller.mutex.Lock()
	controller.profile = config.selectedProfile
	controller.enabled = true
	controller.mutex.Unlock()

	return nil
}

// HandleNodeConfig -
func (controller *chaosController) HandleNodeConfig(config *config.Configs) {
	log.Info("HandleNodeConfig")

	controller.mutex.Lock()
	controller.nodeConfig = config
	controller.mutex.Unlock()
}

// HandleNode -
func (controller *chaosController) HandleNode(node NodeHandler) {
	log.Info("HandleNode")

	controller.mutex.Lock()
	controller.node = node
	controller.mutex.Unlock()

	node.GetCoreComponents().EpochNotifier().RegisterNotifyHandler(controller)
}

// EpochConfirmed -
func (controller *chaosController) EpochConfirmed(epoch uint32, timestamp uint64) {
	log.Info("EpochConfirmed", "epoch", epoch, "timestamp", timestamp)

	controller.HandlePoint(PointInput{
		Name: string(pointEpochConfirmed),
	})
}

// HandlePoint -
func (controller *chaosController) HandlePoint(input PointInput) error {
	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	if !controller.enabled {
		return nil
	}

	failures := controller.profile.getFailuresOnPoint(input.Name)
	if len(failures) == 0 {
		return nil
	}

	for _, failure := range failures {
		circumstance := controller.acquireCircumstanceNoLock(failure.Name, input)

		shouldFail := circumstance.anyExpression(failure.Triggers)
		if shouldFail {
			log.Info("HandlePoint FAIL",
				"failure", failure.Name,
				"point", input.Name,
			)

			switch failType(failure.Type) {
			case failTypePanic:
				return controller.doFailPanic(failure.Name, input)
			case failTypeReturnError:
				return controller.doFailReturnError(failure.Name, input)
			case failTypeEarlyReturn:
				return controller.doFailEarlyReturn(failure.Name, input)
			case failTypeCorruptSignature:
				return controller.doFailCorruptSignature(failure.Name, input)
			case failTypeSleep:
				return controller.doFailSleep(failure.Name, input)
			default:
				return fmt.Errorf("unknown failure type: %s", failure.Type)
			}
		}

		log.Trace("HandlePoint OK",
			"failure", failure.Name,
			"point", input.Name,
			"hasConsensusState", !check.IfNil(input.ConsensusState),
			"hasNodePublicKey", len(input.NodePublicKey) > 0,
			"hasHeader", !check.IfNil(input.Header),
			"hasSignature", len(input.Signature) > 0,
		)
	}

	return nil
}

func (controller *chaosController) acquireCircumstanceNoLock(failure string, input PointInput) *failureCircumstance {
	circumstance := newFailureCircumstance()
	circumstance.point = input.Name
	circumstance.failure = failure
	circumstance.nodeDisplayName = controller.nodeConfig.PreferencesConfig.Preferences.NodeDisplayName

	circumstance.enrichWithLoggerCorrelation(logger.GetCorrelation())
	circumstance.enrichWithConsensusState(input.ConsensusState, input.NodePublicKey)
	circumstance.enrichWithBlockHeader(input.Header)

	log.Trace("circumstance",
		"failure", circumstance.failure,
		"point", circumstance.point,
		"nodeDisplayName", circumstance.nodeDisplayName,
		"randomNumber", circumstance.randomNumber,
		"now", circumstance.now,
		"uptime", circumstance.uptime,
		"shard", circumstance.shard,
		"epoch", circumstance.epoch,
		"round", circumstance.round,
		"nodeIndex", circumstance.nodeIndex,
		"nodePublicKey", circumstance.nodePublicKey,
		"consensusSize", circumstance.consensusSize,
		"amILeader", circumstance.amILeader,
		"blockNonce", circumstance.blockNonce,
		"blockIsStartOfEpoch", circumstance.blockIsStartOfEpoch,
	)

	return circumstance
}

// IsInterfaceNil -
func (controller *chaosController) IsInterfaceNil() bool {
	return controller == nil
}
