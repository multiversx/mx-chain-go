package chronology

import "github.com/davecgh/go-spew/spew"

type RoundServiceImpl struct {
}

func (r RoundServiceImpl) Print(round *Round) {
	spew.Dump(round)
}
