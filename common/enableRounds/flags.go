package enableRounds

type flagsHolder struct {
	checkValueOnExecByCaller *roundFlag
}

// IsCheckValueOnExecByCallerEnabled returns true if checkValueOnExecByCaller is enabled
func (fh *flagsHolder) IsCheckValueOnExecByCallerEnabled() bool {
	return fh.checkValueOnExecByCaller.flag.IsSet()
}
