package builtInFunctions

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	vmcommonBuiltInFunctions "github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
)

// ArgsCreateBuiltInFunctionContainer -
type ArgsCreateBuiltInFunctionContainer struct {
	GasSchedule          core.GasScheduleNotifier
	MapDNSAddresses      map[string]struct{}
	EnableUserNameChange bool
	Marshalizer          marshal.Marshalizer
	Accounts             state.AccountsAdapter
	ShardCoordinator     sharding.Coordinator
}

// CreateBuiltInFunctionContainer creates a factory which will instantiate the built in functions contracts
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

	vmcommonAccounts, ok := args.Accounts.(vmcommon.AccountsAdapter)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	modifiedArgs := vmcommonBuiltInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:               args.GasSchedule.LatestGasSchedule(),
		MapDNSAddresses:      args.MapDNSAddresses,
		EnableUserNameChange: args.EnableUserNameChange,
		Marshalizer:          args.Marshalizer,
		Accounts:             vmcommonAccounts,
		ShardCoordinator:     args.ShardCoordinator,
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
