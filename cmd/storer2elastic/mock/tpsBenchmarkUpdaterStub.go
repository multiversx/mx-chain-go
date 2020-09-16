package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// TPSBenchmarkUpdaterStub -
type TPSBenchmarkUpdaterStub struct {
	IndexTPSForMetaBlockCalled func(metaBlock *block.MetaBlock)
}

// IndexTPSForMetaBlock -
func (t *TPSBenchmarkUpdaterStub) IndexTPSForMetaBlock(metaBlock *block.MetaBlock) {
	if t.IndexTPSForMetaBlockCalled != nil {
		t.IndexTPSForMetaBlockCalled(metaBlock)
	}
}

// IsInterfaceNil -
func (t *TPSBenchmarkUpdaterStub) IsInterfaceNil() bool {
	return t == nil
}
