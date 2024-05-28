package builtInFunctions

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
)

var log = logger.GetOrCreate("process/smartcontract/builtInFunctions")

// ArgsCreateBuiltInFunctionContainer defines the argument structure to create new built in function container
type ArgsCreateBuiltInFunctionContainer struct {
	GasSchedule                           core.GasScheduleNotifier
	MapDNSAddresses                       map[string]struct{}
	MapDNSV2Addresses                     []string
	MapWhiteListedCrossChainMintAddresses []string
	EnableUserNameChange                  bool
	Marshalizer                           marshal.Marshalizer
	Accounts                              state.AccountsAdapter
	ShardCoordinator                      sharding.Coordinator
	EpochNotifier                         vmcommon.EpochNotifier
	EnableEpochsHandler                   vmcommon.EnableEpochsHandler
	GuardedAccountHandler                 vmcommon.GuardedAccountHandler
	PubKeyConverter                       core.PubkeyConverter
	AutomaticCrawlerAddresses             [][]byte
	MaxNumNodesInTransferRole             uint32
	SelfESDTPrefix                        []byte
}

// CreateBuiltInFunctionsFactory creates a container that will hold all the available built in functions
func CreateBuiltInFunctionsFactory(args ArgsCreateBuiltInFunctionContainer) (vmcommon.BuiltInFunctionFactory, error) {
	if check.IfNil(args.GasSchedule) {
		return nil, process.ErrNilGasSchedule
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if args.MapDNSAddresses == nil || args.MapDNSV2Addresses == nil {
		return nil, process.ErrNilDnsAddresses
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.GuardedAccountHandler) {
		return nil, process.ErrNilGuardedAccountHandler
	}
	if check.IfNil(args.PubKeyConverter) {
		return nil, core.ErrNilPubkeyConverter
	}

	vmcommonAccounts, ok := args.Accounts.(vmcommon.AccountsAdapter)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	crawlerAllowedAddress, err := GetAllowedAddress(
		args.ShardCoordinator,
		args.AutomaticCrawlerAddresses)
	if err != nil {
		return nil, err
	}

	log.Debug("createBuiltInFunctionsFactory",
		"shardId", args.ShardCoordinator.SelfId(),
		"crawlerAllowedAddress", crawlerAllowedAddress,
	)

	dnsV2AddressesStrings := args.MapDNSV2Addresses
	convertedDNSV2Addresses, errDecode := factory.DecodeAddresses(args.PubKeyConverter, dnsV2AddressesStrings)
	if errDecode != nil {
		return nil, errDecode
	}

	crossChainWhiteListedAddressesStrings := args.MapWhiteListedCrossChainMintAddresses
	convertedCrossChainWhiteListedAddresses, errDecode := factory.DecodeAddresses(args.PubKeyConverter, crossChainWhiteListedAddressesStrings)
	if errDecode != nil {
		return nil, errDecode
	}

	mapDNSV2Addresses := make(map[string]struct{})
	for _, address := range convertedDNSV2Addresses {
		mapDNSV2Addresses[string(address)] = struct{}{}
	}

	mapWhiteListedCrossChain := make(map[string]struct{})
	for _, address := range convertedCrossChainWhiteListedAddresses {
		mapWhiteListedCrossChain[string(address)] = struct{}{}
	}

	modifiedArgs := vmcommonBuiltInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:                                args.GasSchedule.LatestGasSchedule(),
		MapDNSAddresses:                       args.MapDNSAddresses,
		MapDNSV2Addresses:                     mapDNSV2Addresses,
		MapWhiteListedCrossChainMintAddresses: mapWhiteListedCrossChain,
		EnableUserNameChange:                  args.EnableUserNameChange,
		Marshalizer:                           args.Marshalizer,
		Accounts:                              vmcommonAccounts,
		ShardCoordinator:                      args.ShardCoordinator,
		EnableEpochsHandler:                   args.EnableEpochsHandler,
		GuardedAccountHandler:                 args.GuardedAccountHandler,
		MaxNumOfAddressesForTransferRole:      args.MaxNumNodesInTransferRole,
		ConfigAddress:                         crawlerAllowedAddress,
		SelfESDTPrefix:                        args.SelfESDTPrefix,
	}

	bContainerFactory, err := vmcommonBuiltInFunctions.NewBuiltInFunctionsCreator(modifiedArgs)
	if err != nil {
		return nil, err
	}

	err = bContainerFactory.CreateBuiltInFunctionContainer()
	if err != nil {
		return nil, err
	}

	args.GasSchedule.RegisterNotifyHandler(bContainerFactory)

	return bContainerFactory, nil
}

// GetAllowedAddress returns the allowed crawler address on the current shard
func GetAllowedAddress(coordinator sharding.Coordinator, addresses [][]byte) ([]byte, error) {
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("%w for shard %d, provided count is %d", process.ErrNilCrawlerAllowedAddress, coordinator.SelfId(), len(addresses))
	}

	if coordinator.SelfId() == core.MetachainShardId {
		return core.SystemAccountAddress, nil
	}

	for _, address := range addresses {
		allowedAddressShardId := coordinator.ComputeId(address)
		if allowedAddressShardId == coordinator.SelfId() {
			return address, nil
		}
	}

	return nil, fmt.Errorf("%w for shard %d, provided count is %d", process.ErrNilCrawlerAllowedAddress, coordinator.SelfId(), len(addresses))
}
