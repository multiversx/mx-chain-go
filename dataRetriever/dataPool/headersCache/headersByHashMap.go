package headersCache

type headersByHashMap map[string]headerInfo

func (hhm headersByHashMap) addElement(hash []byte, info headerInfo) bool {
	if _, ok := hhm[string(hash)]; ok {
		return true
	}

	hhm[string(hash)] = info
	return false
}

func (hhm headersByHashMap) deleteElement(hash []byte) {
	delete(hhm, string(hash))
}

func (hhm headersByHashMap) deleteBulk(hashes [][]byte) {
	for _, hash := range hashes {
		delete(hhm, string(hash))
	}
}

func (hhm headersByHashMap) getElement(hash []byte) (headerInfo, bool) {
	if element, ok := hhm[string(hash)]; ok {
		return element, true
	}

	return headerInfo{}, false
}
