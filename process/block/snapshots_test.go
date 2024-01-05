package block

import (
	"bytes"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"
)

func TestTryToWriteInS3(t *testing.T) {
	bucketName := ""
	accessKey := ""
	secretKey := ""
	region := "fra1"

	// Create an AWS session with your DigitalOcean Spaces credentials
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(""), // Set the DigitalOcean Spaces endpoint
		S3ForcePathStyle: aws.Bool(true), // Set to true for DigitalOcean Spaces

	})
	require.Nil(t, err)

	s3Client := s3.New(sess)

	res, err := s3Client.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader([]byte("hello hello")),
		Bucket: aws.String(bucketName),
		Key:    aws.String("test.txt"),
		ACL:    aws.String("public-read"),
	})
	require.NotNil(t, res)
	require.Nil(t, err)
}

func TestReadState(t *testing.T) {
	os.Setenv(envNodePIInterface, "")
	os.Setenv(envS3AccessKey, "")
	os.Setenv(envS3SecretKey, "")
	os.Setenv(envBucketName, "")
	os.Setenv(envS3URL, "")

	ad, err := newAccountsDumper(nil, nil)
	require.Nil(t, err)

	state, err := ad.getLegacyDelegationState()
	require.Nil(t, err)

	err = ad.sClient.Put(state, "test.txt")
	require.Nil(t, err)
}
