package aws

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client wraps S3 operations
type S3Client struct {
	client *s3.Client
	region string
}

// CreateBucket creates a new S3 bucket
func (c *S3Client) CreateBucket(ctx context.Context, bucketName string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	// us-east-1 doesn't accept LocationConstraint
	if c.region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(c.region),
		}
	}

	_, err := c.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	// Block all public access
	_, err = c.client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to block public access: %w", err)
	}

	return nil
}

// SetBucketPolicy sets the bucket policy for CloudFront OAC access
func (c *S3Client) SetBucketPolicy(ctx context.Context, bucketName, distributionArn string) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "AllowCloudFrontServicePrincipalReadOnly",
				"Effect": "Allow",
				"Principal": {
					"Service": "cloudfront.amazonaws.com"
				},
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/*",
				"Condition": {
					"StringEquals": {
						"AWS:SourceArn": "%s"
					}
				}
			}
		]
	}`, bucketName, distributionArn)

	_, err := c.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(policy),
	})
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	return nil
}

// SyncDirectory uploads all files from a local directory to S3
func (c *S3Client) SyncDirectory(ctx context.Context, localDir, bucketName string) error {
	uploader := manager.NewUploader(c.client)

	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Calculate the S3 key (relative path)
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		// Use forward slashes for S3 keys
		key := strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer file.Close()

		// Detect content type
		contentType := detectContentType(path)

		// Upload
		_, err = uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(key),
			Body:        file,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", key, err)
		}

		return nil
	})
}

// DeleteBucketContents deletes all objects in a bucket
func (c *S3Client) DeleteBucketContents(ctx context.Context, bucketName string) error {
	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		var objects []types.ObjectIdentifier
		for _, obj := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		_, err = c.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}
	}

	return nil
}

// DeleteBucket deletes an S3 bucket (must be empty)
func (c *S3Client) DeleteBucket(ctx context.Context, bucketName string) error {
	_, err := c.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}

// GetBucketRegionalDomain returns the regional domain name for an S3 bucket
func (c *S3Client) GetBucketRegionalDomain(bucketName string) string {
	return fmt.Sprintf("%s.s3.%s.amazonaws.com", bucketName, c.region)
}

// detectContentType returns the MIME type for a file based on its extension
func detectContentType(path string) string {
	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)

	if contentType == "" {
		// Default fallback
		contentType = "application/octet-stream"

		// Handle common web extensions that mime might miss
		switch ext {
		case ".js":
			contentType = "application/javascript"
		case ".mjs":
			contentType = "application/javascript"
		case ".css":
			contentType = "text/css"
		case ".html":
			contentType = "text/html"
		case ".json":
			contentType = "application/json"
		case ".svg":
			contentType = "image/svg+xml"
		case ".woff":
			contentType = "font/woff"
		case ".woff2":
			contentType = "font/woff2"
		case ".ttf":
			contentType = "font/ttf"
		case ".ico":
			contentType = "image/x-icon"
		case ".webp":
			contentType = "image/webp"
		case ".avif":
			contentType = "image/avif"
		}
	}

	return contentType
}

// Ensure file is closed properly
var _ io.Closer = (*os.File)(nil)
