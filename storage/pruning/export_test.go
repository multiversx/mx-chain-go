package pruning

func (ps *PruningStorer) ChangeEpoch(epochNum uint32) error {
	return ps.changeEpoch(epochNum)
}

func (ps *PruningStorer) ChangeEpochWithExisting(epoch uint32) error {
	return ps.changeEpochWithExisting(epoch)
}

func RemoveDirectoryIfEmpty(path string) {
	removeDirectoryIfEmpty(path)
}
