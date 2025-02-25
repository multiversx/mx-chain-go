package chaosImpl

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/chaos"
	"github.com/multiversx/mx-chain-go/config"
	nodeConfig "github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ chaos.ChaosController = (*chaosController)(nil)

type chaosController struct {
	mutex      sync.RWMutex
	enabled    bool
	profile    chaosProfile
	nodeConfig *config.Configs
	node       NodeHandler
}

// NewChaosController -
func NewChaosController() *chaosController {
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
func (controller *chaosController) HandleNodeConfig(config interface{}) {
	log.Info("HandleNodeConfig")

	configTyped, ok := config.(*nodeConfig.Configs)
	if !ok {
		log.Error("HandleNodeConfig: invalid config type")
		return
	}

	controller.mutex.Lock()
	controller.nodeConfig = configTyped
	controller.mutex.Unlock()
}

// HandleNode -
func (controller *chaosController) HandleNode(node interface{}) {
	log.Info("HandleNode")

	nodeTyped, ok := node.(NodeHandler)
	if !ok {
		log.Error("HandleNode: invalid node type")
		return
	}

	controller.mutex.Lock()
	controller.node = nodeTyped
	controller.mutex.Unlock()

	nodeTyped.GetCoreComponents().EpochNotifier().RegisterNotifyHandler(controller)
}

// HandleTransaction -
func (controller *chaosController) HandleTransaction(
	transactionHash []byte,
	transaction interface{},
	senderShard uint32,
	receiverShard uint32,
) {
	transactionTyped, ok := transaction.(data.TransactionHandler)
	if !ok {
		log.Error("HandleTransaction: invalid transaction type")
		return
	}

	data := string(transactionTyped.GetData())

	log.Info("HandleTransaction",
		"transactionHash", transactionHash,
		"data", data,
	)
}

// EpochConfirmed -
func (controller *chaosController) EpochConfirmed(epoch uint32, timestamp uint64) {
	log.Info("EpochConfirmed", "epoch", epoch, "timestamp", timestamp)

	controller.HandlePoint(chaos.PointInput{
		Name: "epochConfirmed",
	})
}

// HandlePoint -
func (controller *chaosController) HandlePoint(input chaos.PointInput) chaos.PointOutput {
	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	if !controller.enabled {
		return chaos.PointOutput{}
	}

	failures := controller.profile.getFailuresOnPoint(input.Name)
	if len(failures) == 0 {
		return chaos.PointOutput{}
	}

	for _, failure := range failures {
		circumstance := controller.acquireCircumstanceNoLock(failure.Name, input)

		shouldFail := circumstance.anyExpression(failure.Triggers)
		if shouldFail {
			log.Info("HandlePoint FAIL", "failure", failure.Name, "point", input.Name)

			switch failType(failure.Type) {
			case failTypePanic:
				return controller.doFailPanic(failure.Name, input)
			case failTypeReturnError:
				return controller.doFailReturnError(failure.Name, input)
			case failTypeEarlyReturn:
				return controller.doFailEarlyReturn(failure.Name, input)
			case failTypeCorruptVariables:
				return controller.doFailCorruptVariables(failure.Name, input)
			case failTypeSleep:
				return controller.doFailSleep(failure.Name, input)
			default:
				log.Error("HandlePoint: unknown failure type",
					"failure", failure.Name,
					"point", input.Name,
					"failType", failure.Type,
				)

				return chaos.PointOutput{}
			}
		}

		log.Trace("HandlePoint OK", "failure", failure.Name, "point", input.Name)
	}

	return chaos.PointOutput{}
}

func (controller *chaosController) acquireCircumstanceNoLock(failure string, input chaos.PointInput) *failureCircumstance {
	circumstance := newFailureCircumstance()
	circumstance.point = input.Name
	circumstance.failure = failure
	circumstance.nodeDisplayName = controller.nodeConfig.PreferencesConfig.Preferences.NodeDisplayName

	circumstance.enrichWithLoggerCorrelation(logger.GetCorrelation())
	circumstance.enrichWithPointInput(input)

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
