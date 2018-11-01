// +build integration

package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/generator"
)

const testBucket = "kasita-integration.test"

func TestRead(t *testing.T) {
	assert := assert.New(t)
	s3ReaderWriter := NewS3(testBucket, "s3-reader-writer-closer/read.txt")

	expected := []byte("This is a test file.\n")
	actual, err := s3ReaderWriter.ReadObject()
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func TestReadError(t *testing.T) {
	assert := assert.New(t)
	s3ReaderWriter := NewS3(testBucket, "s3-reader-writer-closer/read-does-not-exist.txt")

	actual, err := s3ReaderWriter.ReadObject()
	assert.Error(err)
	assert.Nil(actual)
}

func TestWrite(t *testing.T) {
	assert := assert.New(t)
	key := fmt.Sprintf("s3-reader-writer-closer/%s/test.txt", generator.String(10))
	s3ReaderWriter := NewS3(testBucket, key)

	expected := []byte(fmt.Sprintf("This is a test file.\n%s\n", generator.String(40)))
	err := s3ReaderWriter.WriteObject(expected)
	assert.NoError(err)

	actual, err := s3ReaderWriter.ReadObject()
	assert.NoError(err)
	assert.Equal(expected, actual)

	list, err := s3ReaderWriter.List("s3-reader-writer-closer", "")
	assert.NoError(err)
	assert.Equal(int64(3), *list.KeyCount)
	keys := make([]string, len(list.Contents))
	for i, res := range list.Contents {
		keys[i] = *res.Key
	}
	assert.Contains(keys, key)

	PruneAlpha(testBucket, key, 0)
}

func TestPrune(t *testing.T) {
	assert := assert.New(t)

	prefix := "PruneAlpha"
	numFiles := 5

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(testBucket),
		Prefix: aws.String(prefix),
	}
	svc := s3.New(session.New())

	var lastUploadedFilename string
	for i := 0; i < numFiles; i++ {
		lastUploadedFilename = fmt.Sprintf("%s/%d", prefix, int32(time.Now().Unix())+int32(i))
		s3 := NewS3(testBucket, lastUploadedFilename)
		err := s3.WriteObject([]byte(generator.String(100)))
		assert.NoError(err)
	}

	objects, err := svc.ListObjectsV2(input)

	assert.Equal(numFiles, len(objects.Contents))
	assert.NoError(err)

	PruneAlpha(testBucket, prefix, numFiles-2)

	objects, err = svc.ListObjectsV2(input)
	assert.Equal(numFiles-2, len(objects.Contents))

	var objectKeys = map[string]bool{}
	for _, value := range objects.Contents {
		objectKeys[*value.Key] = true
	}
	assert.True(objectKeys[lastUploadedFilename])
	assert.NoError(err)

	PruneAlpha(testBucket, prefix, 0)

	PruneAlpha("foo", "bar", numFiles)
}
