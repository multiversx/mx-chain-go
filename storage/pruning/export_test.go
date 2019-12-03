package pruning

func (ps *PruningStorer) ChangeEpoch(epochNum uint32) error {
	return ps.changeEpoch(epochNum)
}

func RemoveDirectoryIfEmpty(path string) {
	removeDirectoryIfEmpty(path)
}
