package storage

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/Kasita-Inc/gadget/errors"
	"github.com/Kasita-Inc/gadget/log"
)

var publicACL = "public-read"

// Bucket wraps the S3 downloader with an in memory cache
type Bucket struct {
	bucket string
	key    string
	// ACL Amazon S3 access control lists value
	ACL string
}

// NewS3 returns a Bucket with an S3 downloader
func NewS3(bucket, key string) *Bucket {
	return &Bucket{
		bucket: bucket,
		key:    key,
		ACL:    publicACL,
	}
}

func newSession() (*session.Session, errors.TracerError) {
	session, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	return session, errors.Wrap(err)
}

// ReadObject downloads a file from s3 into a byte array
func (b *Bucket) ReadObject() ([]byte, errors.TracerError) {
	data := aws.NewWriteAtBuffer([]byte{})

	session, tracerError := newSession()
	if tracerError != nil {
		return nil, tracerError
	}

	downloader := s3manager.NewDownloader(session)

	_, err := downloader.Download(data,
		&s3.GetObjectInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(b.key),
		})

	if err != nil {
		log.Errorf("Issue loading from S3, %s/%s (%s)", b.bucket, b.key, err)
		return nil, errors.Wrap(err)
	}
	return data.Bytes(), nil
}

// WriteObject writes an object to a file in s3
func (b *Bucket) WriteObject(p []byte) errors.TracerError {
	session, tracerError := newSession()
	if tracerError != nil {
		return tracerError
	}

	uploader := s3manager.NewUploader(session)
	upParams := &s3manager.UploadInput{
		Bucket: &b.bucket,
		Key:    &b.key,
		Body:   bytes.NewReader(p),
		ACL:    &b.ACL,
	}
	_, err := uploader.Upload(upParams)
	if nil != err {
		log.Errorf("Issue writing to S3, %s/%s (%s)", b.bucket, b.key, err)
		return errors.Wrap(err)
	}
	return nil
}

// List the contents of a bucket with the given prefix
func (b *Bucket) List(prefix string, startAfter string) (*s3.ListObjectsV2Output, errors.TracerError) {
	session, tracerError := newSession()
	if tracerError != nil {
		return nil, tracerError
	}

	svc := s3.New(session)
	input := &s3.ListObjectsV2Input{
		Bucket:     aws.String(b.bucket),
		Prefix:     aws.String(prefix),
		StartAfter: aws.String(startAfter),
	}

	result, err := svc.ListObjectsV2(input)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return result, nil
}

// PruneByPrefix will remove objects over maxObjects from an S3 bucket with the given prefix
func PruneByPrefix(bucket string, prefix string, maxObjects int) {
	svc := s3.New(session.New())
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	objects, err := svc.ListObjectsV2(input)
	if nil != err {
		return
	}

	if len(objects.Contents) > maxObjects {
		maxIdx := len(objects.Contents) - maxObjects
		for _, obj := range objects.Contents[:maxIdx] {
			delInp := &s3.DeleteObjectInput{
				Bucket: aws.String(bucket),
				Key:    obj.Key,
			}
			_, err = svc.DeleteObject(delInp)
			if nil != err {
				log.Errorf("prune failed for %s/%s\n%#v", bucket, *obj.Key, err)
			}
		}
	}
}
