package block

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Client struct {
	s3         *s3.S3
	bucketName string
}

func newS3Client(accessKey, secretKey, bucketName, url string) (*s3Client, error) {
	region := "fra1"
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(url),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return &s3Client{
		bucketName: bucketName,
		s3:         s3.New(sess),
	}, nil
}

func (s *s3Client) Put(body []byte, key string) error {
	_, err := s.s3.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(body),
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
	})

	return err
}
