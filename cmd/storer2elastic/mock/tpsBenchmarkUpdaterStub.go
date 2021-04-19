package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// TPSBenchmarkUpdaterStub -
type TPSBenchmarkUpdaterStub struct {
	IndexTPSForMetaBlockCalled func(metaBlock data.HeaderHandler)
}

// IndexTPSForMetaBlock -
func (t *TPSBenchmarkUpdaterStub) IndexTPSForMetaBlock(metaBlock data.HeaderHandler) {
	if t.IndexTPSForMetaBlockCalled != nil {
		t.IndexTPSForMetaBlockCalled(metaBlock)
	}
}

// IsInterfaceNil -
func (t *TPSBenchmarkUpdaterStub) IsInterfaceNil() bool {
	return t == nil
}
