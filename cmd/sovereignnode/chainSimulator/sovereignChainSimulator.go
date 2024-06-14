package chainSimulator

import (
	"path"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
	sovereignConfig "github.com/multiversx/mx-chain-go/sovereignnode/config"
	"github.com/multiversx/mx-chain-go/sovereignnode/incomingHeader"
	sovRunType "github.com/multiversx/mx-chain-go/sovereignnode/runType"

	"github.com/multiversx/mx-chain-core-go/core"
)

const (
	numOfShards = 1
)

// ArgsSovereignChainSimulator holds the arguments for sovereign chain simulator
type ArgsSovereignChainSimulator struct {
	SovereignConfigPath string
	*chainSimulator.ArgsChainSimulator
}

// NewSovereignChainSimulator will create a new instance of sovereign chain simulator
func NewSovereignChainSimulator(args ArgsSovereignChainSimulator) (chainSimulatorIntegrationTests.ChainSimulator, error) {
	args.NumOfShards = numOfShards

	alterConfigs := args.AlterConfigsFunction
	configs, err := loadSovereignConfigs(args.SovereignConfigPath)
	if err != nil {
		return nil, err
	}

	args.AlterConfigsFunction = func(cfg *config.Configs) {
		cfg.EconomicsConfig = configs.EconomicsConfig
		cfg.EpochConfig = configs.EpochConfig
		cfg.GeneralConfig.SovereignConfig = *configs.SovereignExtraConfig
		cfg.GeneralConfig.VirtualMachine.Execution.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
		cfg.GeneralConfig.VirtualMachine.Querying.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}

		if alterConfigs != nil {
			alterConfigs(cfg)
			configs.SovereignExtraConfig = &cfg.GeneralConfig.SovereignConfig
		}
	}

	args.CreateGenesisNodesSetup = func(nodesFilePath string, addressPubkeyConverter core.PubkeyConverter, validatorPubkeyConverter core.PubkeyConverter, _ uint32) (sharding.GenesisNodesSetupHandler, error) {
		return sharding.NewSovereignNodesSetup(&sharding.SovereignNodesSetupArgs{
			NodesFilePath:            nodesFilePath,
			AddressPubKeyConverter:   addressPubkeyConverter,
			ValidatorPubKeyConverter: validatorPubkeyConverter,
		})
	}
	args.CreateRatingsData = func(arg rating.RatingsDataArg) (process.RatingsInfoHandler, error) {
		return rating.NewSovereignRatingsData(arg)
	}
	args.CreateIncomingHeaderSubscriber = func(config *config.NotifierConfig, dataPool dataRetriever.PoolsHolder, mainChainNotarizationStartRound uint64, runTypeComponents factory.RunTypeComponentsHolder) (process.IncomingHeaderSubscriber, error) {
		return incomingHeader.CreateIncomingHeaderProcessor(config, dataPool, mainChainNotarizationStartRound, runTypeComponents)
	}
	args.CreateRunTypeComponents = func(argsRunType runType.ArgsRunTypeComponents) (factory.RunTypeComponentsHolder, error) {
		return createSovereignRunTypeComponents(argsRunType, *configs.SovereignExtraConfig)
	}
	args.NodeFactory = node.NewSovereignNodeFactory()
	args.ChainProcessorFactory = NewSovereignChainHandlerFactory()

	return chainSimulator.NewChainSimulator(*args.ArgsChainSimulator)
}

// loadSovereignConfigs loads sovereign configs
func loadSovereignConfigs(configsPath string) (*sovereignConfig.SovereignConfig, error) {
	epochConfig, err := common.LoadEpochConfig(path.Join(configsPath, "enableEpochs.toml"))
	if err != nil {
		return nil, err
	}

	economicsConfig, err := common.LoadEconomicsConfig(path.Join(configsPath, "economics.toml"))
	if err != nil {
		return nil, err
	}

	sovereignExtraConfig, err := sovereignConfig.LoadSovereignGeneralConfig(path.Join(configsPath, "sovereignConfig.toml"))
	if err != nil {
		return nil, err
	}

	return &sovereignConfig.SovereignConfig{
		Configs: &config.Configs{
			EpochConfig:     epochConfig,
			EconomicsConfig: economicsConfig,
		},
		SovereignExtraConfig: sovereignExtraConfig,
	}, nil
}

func createSovereignRunTypeComponents(args runType.ArgsRunTypeComponents, sovereignExtraConfig config.SovereignConfig) (factory.RunTypeComponentsHolder, error) {
	argsSovRunType, err := sovRunType.CreateSovereignArgsRunTypeComponents(args, sovereignExtraConfig)
	if err != nil {
		return nil, err
	}

	sovereignComponentsFactory, err := runType.NewSovereignRunTypeComponentsFactory(*argsSovRunType)
	if err != nil {
		return nil, err
	}

	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(sovereignComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedRunTypeComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedRunTypeComponents, nil
}
