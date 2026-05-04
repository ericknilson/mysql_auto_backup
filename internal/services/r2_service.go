package services

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/erick_nilson/mysql_auto_backup/internal/config"
)

type R2Service struct {
	client *s3.Client
	bucket string
}

func NewR2Service(ctx context.Context, cfg *config.Config) (*R2Service, error) {
	creds := credentials.NewStaticCredentialsProvider(
		cfg.R2AccessKeyID, cfg.R2SecretAccessKey, "",
	)

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &R2Service{client: client, bucket: cfg.R2Bucket}, nil
}

func (s *R2Service) Upload(ctx context.Context, key string, body io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return fmt.Errorf("upload R2: %w", err)
	}
	return nil
}
