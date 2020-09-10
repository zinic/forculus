package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zinic/forculus/config"
)

func NewS3Client(cfg config.S3Uploader) *s3.Client {
	return s3.New(NewAWSConfig(cfg))
}

func NewAWSConfig(cfg config.S3Uploader) aws.Config {
	awsCfg := defaults.Config()
	awsCfg.Region = cfg.Region
	awsCfg.Credentials = &aws.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     cfg.AccessKeyID,
			SecretAccessKey: cfg.SecretAccessKey,
		},
	}

	return awsCfg
}
