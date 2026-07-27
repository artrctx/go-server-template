// R2 Storage
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/artrctx/shuffle-core/internal/env"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Service struct {
	s3      *s3.Client
	presign *s3.PresignClient
}

func New(ctx context.Context, accountId, accessKey, accessSecretKey string) (*Service, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, accessSecretKey, "")),
		config.WithRegion("auto"),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// europe get's their own unique eu sub domain between account id and r2
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	return &Service{client, s3.NewPresignClient(client)}, nil
}

// new service with env
func Get(ctx context.Context) (*Service, error) {
	return New(ctx, env.R2AccountID, env.R2AccessKey, env.R2AccessSecretKey)
}

type GeneratePresignedPutOpt struct {
	Bucket      string
	Key         string
	ContentType *string
	Expires     time.Duration
}

func (s *Service) GeneratePresignedPut(ctx context.Context, opt GeneratePresignedPutOpt) (*v4.PresignedHTTPRequest, error) {
	ps, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(opt.Bucket),
		Key:         aws.String(opt.Key),
		ContentType: opt.ContentType,
	}, s3.WithPresignExpires(opt.Expires))

	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (s *Service) GetObject(ctx context.Context, bucket, key string) (*s3.GetObjectOutput, error) {
	return s.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}
