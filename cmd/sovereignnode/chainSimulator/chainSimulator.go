package chainSimulator

import (
	"path"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-abi-incubator/golang/abi"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	sovereignConfig "github.com/multiversx/mx-chain-go/sovereignnode/config"
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"
	"github.com/multiversx/mx-chain-go/sovereignnode/incomingHeader"
)

type ArgsSovereignChainSimulator struct {
	sovereignExtraConfig config.SovereignConfig
	chainSimulatorArgs   chainSimulator.ArgsChainSimulator
}

// NewSovereignChainSimulator will create a new instance of sovereign chain simulator
func NewSovereignChainSimulator(args ArgsSovereignChainSimulator) (*chainSimulator.Simulator, error) {
	args.chainSimulatorArgs.CreateIncomingHeaderHandler = func(config *config.NotifierConfig, dataPool dataRetriever.PoolsHolder, mainChainNotarizationStartRound uint64, runTypeComponents factory.RunTypeComponentsHolder) (process.IncomingHeaderSubscriber, error) {
		return incomingHeader.CreateIncomingHeaderProcessor(config, dataPool, mainChainNotarizationStartRound, runTypeComponents)
	}
	args.chainSimulatorArgs.GetRunTypeComponents = func(coreComponents factory.CoreComponentsHolder, cryptoComponents factory.CryptoComponentsHolder) (factory.RunTypeComponentsHolder, error) {
		return createSovereignRunTypeComponents(coreComponents, cryptoComponents, args.sovereignExtraConfig)
	}

	return chainSimulator.NewChainSimulator(args.chainSimulatorArgs)
}

// LoadSovereignConfigs - load sovereign configs
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

func createArgsRunTypeComponents(blockSigner crypto.SingleSigner, sovereignExtraConfig config.SovereignConfig) (*runType.ArgsSovereignRunTypeComponents, error) {
	codec := abi.NewDefaultCodec()
	argsDataCodec := dataCodec.ArgsDataCodec{
		Serializer: abi.NewSerializer(codec),
	}

	dataCodecHandler, err := dataCodec.NewDataCodec(argsDataCodec)
	if err != nil {
		return nil, err
	}

	topicsCheckerHandler := incomingHeader.NewTopicsChecker()

	sovHeaderSigVerifier, err := headerCheck.NewSovereignHeaderSigVerifier(blockSigner)
	if err != nil {
		return nil, err
	}

	return &runType.ArgsSovereignRunTypeComponents{
		Config:        sovereignExtraConfig,
		DataCodec:     dataCodecHandler,
		TopicsChecker: topicsCheckerHandler,
		ExtraVerifier: sovHeaderSigVerifier,
	}, nil
}

func createSovereignRunTypeComponents(coreComponents process.CoreComponentsHolder, cryptoComponents factory.CryptoComponentsHolder, sovereignExtraConfig config.SovereignConfig) (factory.RunTypeComponentsHolder, error) {
	runTypeComponentsFactory, _ := runType.NewRunTypeComponentsFactory(coreComponents)
	sovRunTypeArgs, err := createArgsRunTypeComponents(cryptoComponents.BlockSigner(), sovereignExtraConfig)
	if err != nil {
		return nil, err
	}
	sovereignComponentsFactory, _ := runType.NewSovereignRunTypeComponentsFactory(runTypeComponentsFactory, *sovRunTypeArgs)
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
