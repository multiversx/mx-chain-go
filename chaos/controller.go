package chaos

import (
	"fmt"
	"sync"

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
	log.Trace("HandlePoint", "point", input.Name)

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(input)

	for _, failure := range controller.profile.Failures {
		if !failure.isOnPoint(input.Name) {
			continue
		}

		shouldFail := controller.shouldFailNoLock(failure.Name, circumstance)
		if shouldFail {
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
	}

	return nil
}

func (controller *chaosController) acquireCircumstanceNoLock(input PointInput) *failureCircumstance {
	circumstance := newFailureCircumstance()
	circumstance.point = input.Name
	circumstance.nodeDisplayName = controller.nodeConfig.PreferencesConfig.Preferences.NodeDisplayName
	circumstance.enrichWithLoggerCorrelation(logger.GetCorrelation())
	circumstance.enrichWithConsensusState(input.ConsensusState, input.NodePublicKey)

	// Provide header on a best-effort basis.
	circumstance.enrichWithBlockHeader(input.ConsensusState.GetHeader())
	circumstance.enrichWithBlockHeader(input.Header)

	return circumstance
}

func (controller *chaosController) shouldFailNoLock(failureName string, circumstance *failureCircumstance) bool {
	if !controller.enabled {
		return false
	}

	failure, configured := controller.profile.getFailureByName(failureName)
	if !configured {
		return false
	}
	if !failure.Enabled {
		return false
	}

	shouldFail := circumstance.anyExpression(failure.Triggers)
	if shouldFail {
		log.Info("shouldFail", "failureName", failureName)
		return true
	}

	return false
}

// IsInterfaceNil -
func (controller *chaosController) IsInterfaceNil() bool {
	return controller == nil
}
