package aws

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zinic/forculus/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3manager"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/errors"
)

const (
	ErrNotConfigured = errors.New("provider not configured")

	regionProperty          = "region"
	bucketProperty          = "bucket"
	accessKeyIDProperty     = "access_key_id"
	secretAccessKeyProperty = "secret_access_key"
)

func listRequiredKeys() []string {
	return []string{regionProperty, bucketProperty, accessKeyIDProperty, secretAccessKeyProperty}
}

func newAWSConfig(cfg config.StorageProvider) aws.Config {
	awsCfg := defaults.Config()
	awsCfg.Region = cfg.Properties[regionProperty]
	awsCfg.Credentials = &aws.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     cfg.Properties[accessKeyIDProperty],
			SecretAccessKey: cfg.Properties[secretAccessKeyProperty],
		},
	}

	return awsCfg
}

type S3Provider struct {
	cfg        config.StorageProvider
	s3Client   *s3.Client
	s3Uploader *s3manager.Uploader
}

func (s *S3Provider) Configure(cfg config.StorageProvider) error {
	if err := s.Validate(cfg); err != nil {
		return err
	}

	awsCfg := newAWSConfig(cfg)
	s.s3Uploader = s3manager.NewUploader(awsCfg)
	s.s3Client = s3.New(awsCfg)
	s.cfg = cfg
	return nil
}

func (s *S3Provider) Validate(cfg config.StorageProvider) error {
	var (
		requiredKeys          = listRequiredKeys()
		expectedNumProperties = len(requiredKeys)
		numProperties         = len(cfg.Properties)
	)

	if numProperties != expectedNumProperties {
		return fmt.Errorf("found %d properties for S3 storage provider but expected %d", numProperties, expectedNumProperties)
	}

	for _, requiredKey := range listRequiredKeys() {
		if value, found := cfg.Properties[requiredKey]; !found {
			return fmt.Errorf("missing required S3 property \"%s\"", requiredKey)
		} else if len(value) == 0 {
			return fmt.Errorf("zero-length value found for S3 property \"%s\"", requiredKey)
		}
	}

	return nil
}

func (s *S3Provider) Write(key string, reader io.Reader) error {
	if s.s3Uploader == nil {
		return ErrNotConfigured
	}

	input := &s3manager.UploadInput{
		Body:   reader,
		Bucket: aws.String(s.cfg.Properties[bucketProperty]),
		Key:    aws.String(key),
	}

	_, err := s.s3Uploader.Upload(input)
	return err
}

func (s *S3Provider) Read(key string) (io.ReadCloser, error) {
	if s.s3Uploader == nil {
		return nil, ErrNotConfigured
	}

	req := s.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Properties[bucketProperty]),
		Key:    aws.String(key),
	})

	if resp, err := req.Send(context.Background()); err != nil {
		return nil, err
	} else {
		return resp.Body, nil
	}
}

func (s *S3Provider) Stat(key string) (storage.Details, error) {
	var details storage.Details

	if s.s3Uploader == nil {
		return details, ErrNotConfigured
	}

	req := s.s3Client.HeadObjectRequest(&s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.Properties[bucketProperty]),
		Key:    aws.String(key),
	})

	if resp, err := req.Send(context.Background()); err != nil {
		return details, err
	} else {
		details.Size = *resp.ContentLength
		return details, nil
	}
}
