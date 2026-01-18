package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Client *s3.Client
var BucketName string

func IsConfigured() bool {
	return S3Client != nil && BucketName != ""
}

func InitS3() error {
	accessKeyID := os.Getenv("TIGRIS_STORAGE_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("TIGRIS_STORAGE_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("TIGRIS_STORAGE_ENDPOINT")
	bucketName := os.Getenv("TIGRIS_BUCKET_NAME")

	if accessKeyID == "" || secretAccessKey == "" || endpoint == "" || bucketName == "" {
		return fmt.Errorf("missing required Tigris S3 environment variables")
	}

	BucketName = bucketName

	creds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(creds),
		config.WithRegion("auto"), // tigris
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	S3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return nil
}

func UploadFile(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	return err
}

func DownloadFile(ctx context.Context, key string) ([]byte, error) {
	result, err := S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func DeleteFile(ctx context.Context, key string) error {
	_, err := S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})
	return err
}

func GetPresignedURL(ctx context.Context, key string, expireSeconds int64) (string, error) {
	presignClient := s3.NewPresignClient(S3Client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(time.Duration(expireSeconds)*time.Second))
	if err != nil {
		return "", err
	}

	return request.URL, nil
}
