package mock

import (
	"bytes"
	"io"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type DatabaseWriterMock struct {
	DoRequestCalled     func(req esapi.IndexRequest) error
	DoBulkRequestCalled func(buff *bytes.Buffer, index string) error
}

func (dwm *DatabaseWriterMock) DoRequest(req esapi.IndexRequest) error {
	if dwm.DoRequestCalled != nil {
		return dwm.DoRequestCalled(req)
	}
	return nil
}

func (dwm *DatabaseWriterMock) DoBulkRequest(buff *bytes.Buffer, index string) error {
	if dwm.DoBulkRequestCalled != nil {
		return dwm.DoBulkRequestCalled(buff, index)
	}
	return nil
}

func (dwm *DatabaseWriterMock) CheckAndCreateIndex(_ string, _ io.Reader) error {
	return nil
}
