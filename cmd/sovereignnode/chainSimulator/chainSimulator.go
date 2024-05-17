package chainSimulator

import (
	"path"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
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

type ArgsSovereignChainSimulator struct {
	SovereignExtraConfig config.SovereignConfig
	ChainSimulatorArgs   chainSimulator.ArgsChainSimulator
}

// NewSovereignChainSimulator will create a new instance of sovereign chain simulator
func NewSovereignChainSimulator(args ArgsSovereignChainSimulator) (*chainSimulator.Simulator, error) {
	args.ChainSimulatorArgs.CreateGenesisNodesSetup = func(nodesFilePath string, addressPubkeyConverter core.PubkeyConverter, validatorPubkeyConverter core.PubkeyConverter, _ uint32) (sharding.GenesisNodesSetupHandler, error) {
		return sharding.NewSovereignNodesSetup(&sharding.SovereignNodesSetupArgs{
			NodesFilePath:            nodesFilePath,
			AddressPubKeyConverter:   addressPubkeyConverter,
			ValidatorPubKeyConverter: validatorPubkeyConverter,
		})
	}
	args.ChainSimulatorArgs.CreateRatingsData = func(arg rating.RatingsDataArg) (process.RatingsInfoHandler, error) {
		return rating.NewSovereignRatingsData(arg)
	}
	args.ChainSimulatorArgs.CreateIncomingHeaderHandler = func(config *config.NotifierConfig, dataPool dataRetriever.PoolsHolder, mainChainNotarizationStartRound uint64, runTypeComponents factory.RunTypeComponentsHolder) (process.IncomingHeaderSubscriber, error) {
		return incomingHeader.CreateIncomingHeaderProcessor(config, dataPool, mainChainNotarizationStartRound, runTypeComponents)
	}
	args.ChainSimulatorArgs.GetRunTypeComponents = func(argsRunType runType.ArgsRunTypeComponents) (factory.RunTypeComponentsHolder, error) {
		return createSovereignRunTypeComponents(argsRunType, args.SovereignExtraConfig)
	}
	args.ChainSimulatorArgs.NodeFactory = node.NewSovereignNodeFactory()
	args.ChainSimulatorArgs.ChainProcessorFactory = NewSovereignProcessorFactory()

	return chainSimulator.NewChainSimulator(args.ChainSimulatorArgs)
}

// LoadSovereignConfigs loads sovereign configs
func LoadSovereignConfigs(configsPath string) (*config.EpochConfig, *config.EconomicsConfig, *config.SovereignConfig, error) {
	epochConfig, err := common.LoadEpochConfig(path.Join(configsPath, "enableEpochs.toml"))
	if err != nil {
		return nil, nil, nil, err
	}

	economicsConfig, err := common.LoadEconomicsConfig(path.Join(configsPath, "economics.toml"))
	if err != nil {
		return nil, nil, nil, err
	}

	sovereignExtraConfig, err := sovereignConfig.LoadSovereignGeneralConfig(path.Join(configsPath, "sovereignConfig.toml"))
	sovereignExtraConfig.OutGoingBridgeCertificate = config.OutGoingBridgeCertificate{
		CertificatePath:   "/home/ubuntu/MultiversX/testnet/config/certificate.crt",
		CertificatePkPath: "/home/ubuntu/MultiversX/testnet/config/private_key.pem",
	}
	if err != nil {
		return nil, nil, nil, err
	}

	return epochConfig, economicsConfig, sovereignExtraConfig, nil
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
