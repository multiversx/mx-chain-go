package systemSmartContracts

import (
	"math/big"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func (d *delegation) createAndAddLogEntry(contractCallInput *vmcommon.ContractCallInput, topics ...[]byte) {
	d.createAndAddLogEntryCustom(contractCallInput.Function, contractCallInput.CallerAddr, topics...)
}

func (d *delegation) createAndAddLogEntryCustom(identifier string, address []byte, topics ...[]byte) {
	entry := &vmcommon.LogEntry{
		Identifier: []byte(identifier),
		Address:    address,
		Topics:     topics,
	}

	d.eei.AddLogEntry(entry)
}

func (d *delegation) createAndAddLogEntryForWithdraw(
	function string,
	address []byte,
	actualUserUnBond *big.Int,
	globalFund *GlobalFundData,
	delegator *DelegatorData,
	numUsers uint64,
	wasDeleted bool,
	withdrawFundKeys [][]byte,
) {
	activeFund := d.getFundForLogEntry(delegator.ActiveFund)
	topics := append(make([][]byte, 0), actualUserUnBond.Bytes(), activeFund.Bytes(), big.NewInt(0).SetUint64(numUsers).Bytes(), globalFund.TotalActive.Bytes(), boolToSlice(wasDeleted))
	topics = append(topics, withdrawFundKeys...)
	d.createAndAddLogEntryCustom(function, address, topics...)
}

func (d *delegation) createAndAddLogEntryForDelegate(
	contractCallInput *vmcommon.ContractCallInput,
	delegationValue *big.Int,
	globalFund *GlobalFundData,
	delegator *DelegatorData,
	dStatus *DelegationContractStatus,
	isNew bool,
) {
	numUsersWithCurrent := dStatus.NumUsers
	if isNew {
		numUsersWithCurrent++
	}

	activeFund := d.getFundForLogEntry(delegator.ActiveFund)
	numUsers := big.NewInt(0).SetUint64(numUsersWithCurrent)
	numActiveWithCurrentValue := big.NewInt(0).Add(globalFund.TotalActive, delegationValue)
	delegatorActiveWithCurrent := big.NewInt(0).Add(activeFund, delegationValue)

	topics := [][]byte{delegationValue.Bytes(), delegatorActiveWithCurrent.Bytes(), numUsers.Bytes(), numActiveWithCurrentValue.Bytes()}

	address := contractCallInput.CallerAddr
	function := contractCallInput.Function
	if function == initFromValidatorData ||
		function == mergeValidatorDataToDelegation ||
		function == changeOwner {
		address = contractCallInput.Arguments[0]

		topics = append(topics, contractCallInput.RecipientAddr)
	}
	if function == core.SCDeployInitFunctionName {
		topics = append(topics, contractCallInput.RecipientAddr)
	}

	entry := &vmcommon.LogEntry{
		Identifier: []byte("delegate"),
		Address:    address,
		Topics:     topics,
	}

	d.eei.AddLogEntry(entry)
}

func (d *delegation) getFundForLogEntry(activeFund []byte) *big.Int {
	if len(activeFund) == 0 {
		return big.NewInt(0)
	}

	fund, err := d.getFund(activeFund)
	if err != nil {
		log.Warn("d.getFundForLogEntry cannot get fund", "error", err.Error())

		return big.NewInt(0)
	}

	return fund.Value
}

func (d *delegation) createLogEventsForChangeOwner(
	args *vmcommon.ContractCallInput,
	ownerDelegatorData *DelegatorData,
) {
	globalFund, err := d.getGlobalFundData()
	if err != nil {
		globalFund = &GlobalFundData{
			TotalActive: big.NewInt(0),
		}

		log.Warn("d.changeOwner cannot get global fund data", "error", err)
	}
	dStatus, err := d.getDelegationStatus()
	if err != nil {
		dStatus = &DelegationContractStatus{}

		log.Warn("d.changeOwner cannot get delegation status", "error", err)
	}

	d.createAndAddLogEntryForDelegate(args, big.NewInt(0), globalFund, ownerDelegatorData, dStatus, false)
	d.createAndAddLogEntryForWithdraw(withdraw, args.CallerAddr, big.NewInt(0), globalFund, ownerDelegatorData, d.numUsers(), true, nil)

	entry := &vmcommon.LogEntry{
		Identifier: []byte(core.BuiltInFunctionChangeOwnerAddress),
		Address:    args.RecipientAddr,
		Topics:     [][]byte{args.Arguments[0]},
	}
	d.eei.AddLogEntry(entry)
}

func boolToSlice(b bool) []byte {
	return []byte(strconv.FormatBool(b))
}
