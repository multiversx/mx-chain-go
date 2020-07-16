package indexer

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

type firehoseConfig struct {
	Region string
}

type firehoseClient struct {
	fClient *firehose.Firehose
}

func newFirehoseClient(cfg *firehoseConfig) (*firehoseClient, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
	}))

	return &firehoseClient{
		fClient: firehose.New(sess),
	}, nil
}

func (fc *firehoseClient) PutBulkRecord(bulk *bytes.Buffer) {
	streamName := "corcotx"

	pco, err := fc.fClient.PutRecord(&firehose.PutRecordInput{
		DeliveryStreamName: &streamName,
		Record: &firehose.Record{
			Data: bulk.Bytes(),
		},
	})

	if err != nil {
		log.Warn("put err", "this is it", err.Error())
	}

	fmt.Println(pco)
}
