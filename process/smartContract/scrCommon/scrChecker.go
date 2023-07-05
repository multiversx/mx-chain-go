package scrCommon

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/scrChecker")

type ArgsSCRChecker struct {
	Hasher     hashing.Hasher
	Marshaller marshal.Marshalizer
	PubKeyConv core.PubkeyConverter
	Accounts   state.AccountsAdapter
	AccGetter  AccountGetter
}

type scrChecker struct {
	hasher     hashing.Hasher
	marshaller marshal.Marshalizer
	pubKeyConv core.PubkeyConverter
	accounts   state.AccountsAdapter
	accGetter  AccountGetter
}

func NewSCRChecker(args *ArgsSCRChecker) (SCRChecker, error) {
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.PubKeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.AccGetter) {
		return nil, errNilAccGetter
	}

	return &scrChecker{
		hasher:     args.Hasher,
		marshaller: args.Marshaller,
		pubKeyConv: args.PubKeyConv,
		accounts:   args.Accounts,
		accGetter:  args.AccGetter,
	}, nil
}

func (sc *scrChecker) CheckSCRBeforeProcessing(scr *smartContractResult.SmartContractResult) (*ScrProcessingData, error) {
	scrHash, err := core.CalculateHash(sc.marshaller, sc.hasher, scr)
	if err != nil {
		log.Debug("CalculateHash error", "error", err)
		return nil, err
	}

	dstAcc, err := sc.accGetter.GetAccountFromAddress(scr.RcvAddr)
	if err != nil {
		return nil, err
	}
	sndAcc, err := sc.accGetter.GetAccountFromAddress(scr.SndAddr)
	if err != nil {
		return nil, err
	}

	if check.IfNil(dstAcc) {
		err = process.ErrNilSCDestAccount
		return nil, err
	}

	snapshot := sc.accounts.JournalLen()
	process.DisplayProcessTxDetails(
		"ProcessSmartContractResult: receiver account details",
		dstAcc,
		scr,
		scrHash,
		sc.pubKeyConv,
	)

	return &ScrProcessingData{
		Hash:        scrHash,
		Snapshot:    snapshot,
		Sender:      sndAcc,
		Destination: dstAcc,
	}, nil
}

func (sc *scrChecker) IsInterfaceNil() bool {
	return sc == nil
}
