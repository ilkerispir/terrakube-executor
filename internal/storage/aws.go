package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSStorageService struct {
	client     *s3.Client
	uploader   *manager.Uploader
	bucketName string
}

func getEnvWithFallback(primary, fallback string) string {
	val := os.Getenv(primary)
	if val == "" {
		return os.Getenv(fallback)
	}
	return val
}

func NewAWSStorageService() (*AWSStorageService, error) {
	region := getEnvWithFallback("AWS_REGION", "AwsTerraformStateRegion")
	bucketName := getEnvWithFallback("AWS_BUCKET_NAME", "AwsTerraformStateBucketName")
	endpoint := getEnvWithFallback("AWS_ENDPOINT", "AwsEndpoint")
	accessKey := getEnvWithFallback("AWS_ACCESS_KEY_ID", "AwsTerraformStateAccessKey")
	secretKey := getEnvWithFallback("AWS_SECRET_ACCESS_KEY", "AwsTerraformStateSecretKey")
	enableRoleAuth := getEnvWithFallback("AWS_ENABLE_ROLE_AUTH", "AwsEnableRoleAuth")

	log.Printf("Initializing AWS Storage Service")
	log.Printf("Region: %s", region)
	log.Printf("Bucket: %s", bucketName)
	log.Printf("Endpoint: %s", endpoint)

	var cfg aws.Config
	var err error

	if enableRoleAuth == "true" {
		log.Printf("AWS Role Auth is enabled, using default AWS credentials chain")
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     accessKey,
					SecretAccessKey: secretKey,
				}, nil
			})),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	if endpoint != "" {
		cfg.BaseEndpoint = aws.String(endpoint)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	uploader := manager.NewUploader(client)

	return &AWSStorageService{
		client:     client,
		uploader:   uploader,
		bucketName: bucketName,
	}, nil
}

func (s *AWSStorageService) UploadFile(path string, content io.Reader) error {
	_, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
		Body:   content,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return nil
}

func (s *AWSStorageService) DownloadFile(path string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}
	return out.Body, nil
}
