package chainSimulator

import (
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

func loadEpochConfig() (*config.EpochConfig, error) {
	epochConfig, err := common.LoadEpochConfig("../config/enableEpochs.toml")
	if err != nil {
		return nil, err
	}

	return epochConfig, nil
}

func loadEconomicsConfig() (*config.EconomicsConfig, error) {
	economicsConfig, err := common.LoadEconomicsConfig("../config/economics.toml")
	if err != nil {
		return nil, err
	}

	return economicsConfig, nil
}

func loadSovereignConfig() (*config.SovereignConfig, error) {
	sovereignExtraConfig, err := sovereignConfig.LoadSovereignGeneralConfig("../config/sovereignConfig.toml")
	if err != nil {
		return nil, err
	}

	sovereignExtraConfig.OutGoingBridgeCertificate = config.OutGoingBridgeCertificate{
		CertificatePath:   "/home/ubuntu/MultiversX/testnet/config/certificate.crt",
		CertificatePkPath: "/home/ubuntu/MultiversX/testnet/config/private_key.pem",
	}

	return sovereignExtraConfig, nil
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

func createRunTypeComponents(coreComponents process.CoreComponentsHolder, cryptoComponents factory.CryptoComponentsHolder, sovereignExtraConfig config.SovereignConfig) (factory.RunTypeComponentsHolder, error) {
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

// NewSovereignChainSimulator will create a new instance of sovereign chain simulator
func NewSovereignChainSimulator(args chainSimulator.ArgsChainSimulator) (*chainSimulator.Simulator, error) {
	sovereignExtraConfig, err := loadSovereignConfig()
	if err != nil {
		return nil, err
	}

	args.CreateIncomingHeaderHandler = func(config *config.NotifierConfig, dataPool dataRetriever.PoolsHolder, mainChainNotarizationStartRound uint64, runTypeComponents factory.RunTypeComponentsHolder) (process.IncomingHeaderSubscriber, error) {
		return incomingHeader.CreateIncomingHeaderProcessor(config, dataPool, mainChainNotarizationStartRound, runTypeComponents)
	}
	args.GetRunTypeComponents = func(coreComponents factory.CoreComponentsHolder, cryptoComponents factory.CryptoComponentsHolder) (factory.RunTypeComponentsHolder, error) {
		return createRunTypeComponents(coreComponents, cryptoComponents, *sovereignExtraConfig)
	}

	return chainSimulator.NewChainSimulator(args)
}
