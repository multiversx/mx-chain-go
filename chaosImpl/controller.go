package chaosImpl

import (
	"encoding/json"
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
	config     *chaosConfig
	profile    chaosProfile
	nodeConfig *config.Configs
	node       NodeHandler
	selfShard  uint32
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
	controller.config = config
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
	controller.selfShard = nodeTyped.GetProcessComponents().ShardCoordinator().SelfId()
}

// HandleTransaction -
func (controller *chaosController) HandleTransaction(
	transactionHash []byte,
	transaction interface{},
	senderShard uint32,
	receiverShard uint32,
) {
	if receiverShard != controller.selfShard {
		return
	}

	transactionTyped, ok := transaction.(data.TransactionHandler)
	if !ok {
		log.Error("HandleTransaction: invalid transaction type")
		return
	}

	// A simple check to recognize management transactions
	if transactionTyped.GetGasPrice() != managementTransactionsGasPrice {
		return
	}

	log.Info("HandleTransaction: management transaction", "hash", transactionHash)

	data := transactionTyped.GetData()

	var config managementCommand

	err := json.Unmarshal(data, &config)
	if err != nil {
		log.Error("HandleTransaction: could not unmarshal management command", "error", err, "data", string(data))
	}

	err = controller.handleManagementCommand(config)
	if err != nil {
		log.Error("HandleTransaction: could not handle management command", "error", err)
	}
}

func (controller *chaosController) handleManagementCommand(command managementCommand) error {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	action := managementCommandAction(command.Action)

	switch action {
	case managementToggleChaos:
		log.Info("handleManagementCommand: toggleChaos", "toggle", command.ToggleValue)

		controller.enabled = command.ToggleValue
	case managementSelectProfile:
		log.Info("handleManagementCommand: selectProfile", "profile", command.ProfileName)

		err := controller.config.selectProfile(command.ProfileName)
		if err != nil {
			return err
		}
	case managementToggleFailure:
		log.Info("handleManagementCommand: toggleFailure", "profile", command.ProfileName, "failure", command.FailureName, "toggle", command.ToggleValue)

		err := controller.config.toggleFailure(command.ProfileName, command.FailureName, command.ToggleValue)
		if err != nil {
			return err
		}
	case managementAddFailure:
		log.Info("handleManagementCommand: addFailure", "profile", command.ProfileName, "failure", command.Failure.Name)

		err := controller.config.addFailure(command.ProfileName, command.Failure)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown management command action: %s", action)
	}

	return nil
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
