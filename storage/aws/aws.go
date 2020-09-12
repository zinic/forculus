package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3manager"
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/errors"
	"io"
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
	s3Uploader *s3manager.Uploader
}

func (s *S3Provider) Configure(cfg config.StorageProvider) error {
	if err := s.Validate(cfg); err != nil {
		return err
	}

	s.s3Uploader = s3manager.NewUploader(newAWSConfig(cfg))
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
