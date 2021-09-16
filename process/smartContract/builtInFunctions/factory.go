package builtInFunctions

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	vmcommonBuiltInFunctions "github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
)

// ArgsCreateBuiltInFunctionContainer defines the argument structure to create new built in function container
type ArgsCreateBuiltInFunctionContainer struct {
	GasSchedule                          core.GasScheduleNotifier
	MapDNSAddresses                      map[string]struct{}
	EnableUserNameChange                 bool
	Marshalizer                          marshal.Marshalizer
	Accounts                             state.AccountsAdapter
	ShardCoordinator                     sharding.Coordinator
	EpochNotifier                        vmcommon.EpochNotifier
	ESDTMultiTransferEnableEpoch         uint32
	ESDTTransferRoleEnableEpoch          uint32
	GlobalMintBurnDisableEpoch           uint32
	ESDTTransferMetaEnableEpoch          uint32
	ESDTNFTCreateOnMultiShardEnableEpoch uint32
}

// CreateBuiltInFunctionContainer creates a container that will hold all the available built in functions
func CreateBuiltInFunctionContainer(args ArgsCreateBuiltInFunctionContainer) (vmcommon.BuiltInFunctionContainer, error) {
	if check.IfNil(args.GasSchedule) {
		return nil, process.ErrNilGasSchedule
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if args.MapDNSAddresses == nil {
		return nil, process.ErrNilDnsAddresses
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	vmcommonAccounts, ok := args.Accounts.(vmcommon.AccountsAdapter)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	modifiedArgs := vmcommonBuiltInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:                              args.GasSchedule.LatestGasSchedule(),
		MapDNSAddresses:                     args.MapDNSAddresses,
		EnableUserNameChange:                args.EnableUserNameChange,
		Marshalizer:                         args.Marshalizer,
		Accounts:                            vmcommonAccounts,
		ShardCoordinator:                    args.ShardCoordinator,
		EpochNotifier:                       args.EpochNotifier,
		ESDTNFTImprovementV1ActivationEpoch: args.ESDTMultiTransferEnableEpoch,
		ESDTTransferToMetaEnableEpoch:       args.ESDTTransferMetaEnableEpoch,
		ESDTTransferRoleEnableEpoch:         args.ESDTTransferRoleEnableEpoch,
		GlobalMintBurnDisableEpoch:          args.GlobalMintBurnDisableEpoch,
	}

	bContainerFactory, err := vmcommonBuiltInFunctions.NewBuiltInFunctionsFactory(modifiedArgs)
	if err != nil {
		return nil, err
	}

	container, err := bContainerFactory.CreateBuiltInFunctionContainer()
	if err != nil {
		return nil, err
	}

	args.GasSchedule.RegisterNotifyHandler(bContainerFactory)

	return container, nil
}
