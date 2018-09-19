package chronology

import "github.com/davecgh/go-spew/spew"

type EpochServiceImpl struct {
}

func (e EpochServiceImpl) Print(epoch *Epoch) {
	spew.Dump(epoch)
}
