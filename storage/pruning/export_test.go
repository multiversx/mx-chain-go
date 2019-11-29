package pruning

func (pd *PruningStorer) ChangeEpoch(epochNum uint32) error {
	return pd.changeEpoch(epochNum)
}
