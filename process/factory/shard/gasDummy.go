package shard

func getDummyBlockGasLimit() uint64 {
	return uint64(1000000)
}

func getDummyGasSchedule() map[string]uint64 {
	gasSchedule := make(map[string]uint64)
	gasSchedule["StorePerByte"] = 1
	gasSchedule["DataCopyPerByte"] = 1
	gasSchedule["GetOwner"] = 1
	gasSchedule["GetExternalBalance"] = 1
	gasSchedule["GetBlockHash"] = 1
	gasSchedule["TransferValue"] = 1
	gasSchedule["GetArgument"] = 1
	gasSchedule["GetFunction"] = 1
	gasSchedule["GetNumArguments"] = 1
	gasSchedule["StorageStore"] = 1
	gasSchedule["StorageLoad"] = 1
	gasSchedule["GetCaller"] = 1
	gasSchedule["GetCallValue"] = 1
	gasSchedule["Log"] = 1
	gasSchedule["Finish"] = 1
	gasSchedule["SignalError"] = 1
	gasSchedule["GetBlockTimeStamp"] = 1
	gasSchedule["GetGasLeft"] = 1
	gasSchedule["Int64GetArgument"] = 1
	gasSchedule["Int64StorageStore"] = 1
	gasSchedule["Int64StorageLoad"] = 1
	gasSchedule["Int64Finish"] = 1
	gasSchedule["GetStateRootHash"] = 1
	gasSchedule["GetBlockNonce"] = 1
	gasSchedule["GetBlockEpoch"] = 1
	gasSchedule["GetBlockRound"] = 1
	gasSchedule["GetBlockRandomSeed"] = 1
	gasSchedule["UseGas"] = 1
	gasSchedule["GetAddress"] = 1
	gasSchedule["GetExternalBalance"] = 1
	gasSchedule["GetBlockHash"] = 1
	gasSchedule["Call"] = 1
	gasSchedule["CallDataCopy"] = 1
	gasSchedule["GetCallDataSize"] = 1
	gasSchedule["CallCode"] = 1
	gasSchedule["CallDelegate"] = 1
	gasSchedule["CallStatic"] = 1
	gasSchedule["StorageStore"] = 1
	gasSchedule["StorageLoad"] = 1
	gasSchedule["GetCaller"] = 1
	gasSchedule["GetCallValue"] = 1
	gasSchedule["CodeCopy"] = 1
	gasSchedule["GetCodeSize"] = 1
	gasSchedule["GetBlockCoinbase"] = 1
	gasSchedule["Create"] = 1
	gasSchedule["GetBlockDifficulty"] = 1
	gasSchedule["ExternalCodeCopy"] = 1
	gasSchedule["GetExternalCodeSize"] = 1
	gasSchedule["GetGasLeft"] = 1
	gasSchedule["GetBlockGasLimit"] = 1
	gasSchedule["GetTxGasPrice"] = 1
	gasSchedule["Log"] = 1
	gasSchedule["GetBlockNumber"] = 1
	gasSchedule["GetTxOrigin"] = 1
	gasSchedule["Finish"] = 1
	gasSchedule["Revert"] = 1
	gasSchedule["GetReturnDataSize"] = 1
	gasSchedule["ReturnDataCopy"] = 1
	gasSchedule["SelfDestruct"] = 1
	gasSchedule["GetBlockTimeStamp"] = 1
	gasSchedule["BigIntNew"] = 1
	gasSchedule["BigIntByteLength"] = 1
	gasSchedule["BigIntGetBytes"] = 1
	gasSchedule["BigIntSetBytes"] = 1
	gasSchedule["BigIntIsInt64"] = 1
	gasSchedule["BigIntGetInt64"] = 1
	gasSchedule["BigIntSetInt64"] = 1
	gasSchedule["BigIntAdd"] = 1
	gasSchedule["BigIntSub"] = 1
	gasSchedule["BigIntMul"] = 1
	gasSchedule["BigIntCmp"] = 1
	gasSchedule["BigIntFinish"] = 1
	gasSchedule["BigIntStorageLoad"] = 1
	gasSchedule["BigIntStorageStore"] = 1
	gasSchedule["BigIntGetArgument"] = 1
	gasSchedule["BigIntGetCallValue"] = 1
	gasSchedule["BigIntGetExternalBalance"] = 1
	gasSchedule["SHA256"] = 1
	gasSchedule["Keccak256"] = 1
	return gasSchedule
}
