package scrCommon

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/smartcontract/scrCommon")

type scProcessorHelper struct {
	accounts        state.AccountsAdapter
	coordinator     sharding.Coordinator
	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	pubKeyConverter core.PubkeyConverter
}

// SCProcessorHelperArgs represents the arguments needed to create a new sc processor helper
type SCProcessorHelperArgs struct {
	Accounts         state.AccountsAdapter
	ShardCoordinator sharding.Coordinator
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	PubkeyConverter  core.PubkeyConverter
}

// NewSCProcessorHelper is a helper that provides methods for checking and processing SCs
func NewSCProcessorHelper(args SCProcessorHelperArgs) (*scProcessorHelper, error) {
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.PubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}

	return &scProcessorHelper{
		accounts:        args.Accounts,
		coordinator:     args.ShardCoordinator,
		marshalizer:     args.Marshalizer,
		hasher:          args.Hasher,
		pubKeyConverter: args.PubkeyConverter,
	}, nil
}

// GetAccountFromAddress returns the account from the given address
func (scph *scProcessorHelper) GetAccountFromAddress(address []byte) (state.UserAccountHandler, error) {
	shardForCurrentNode := scph.coordinator.SelfId()
	shardForSrc := scph.coordinator.ComputeId(address)
	if shardForCurrentNode != shardForSrc {
		return nil, nil
	}

	acnt, err := scph.accounts.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	stAcc, ok := acnt.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return stAcc, nil
}

// CheckSCRBeforeProcessing checks the smart contract result before processing it
func (scph *scProcessorHelper) CheckSCRBeforeProcessing(scr *smartContractResult.SmartContractResult) (process.ScrProcessingDataHandler, error) {
	if check.IfNil(scr) {
		return nil, process.ErrNilSmartContractResult
	}

	scrHash, err := core.CalculateHash(scph.marshalizer, scph.hasher, scr)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return nil, err
	}

	dstAcc, err := scph.GetAccountFromAddress(scr.RcvAddr)
	if err != nil {
		return nil, err
	}
	sndAcc, err := scph.GetAccountFromAddress(scr.SndAddr)
	if err != nil {
		return nil, err
	}

	if check.IfNil(dstAcc) {
		err = process.ErrNilSCDestAccount
		return nil, err
	}

	snapshot := scph.accounts.JournalLen()
	process.DisplayProcessTxDetails(
		"ProcessSmartContractResult: receiver account details",
		dstAcc,
		scr,
		scrHash,
		scph.pubKeyConverter,
	)

	return &ScrProcessingData{
		Hash:        scrHash,
		Snapshot:    snapshot,
		Sender:      sndAcc,
		Destination: dstAcc,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scph *scProcessorHelper) IsInterfaceNil() bool {
	return scph == nil
}
