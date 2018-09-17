package chronology

import "github.com/davecgh/go-spew/spew"

type RoundServiceImpl struct {
}

func (RoundServiceImpl) Print(round *Round) {
	spew.Dump(round)
}
