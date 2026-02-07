package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Client wraps AWS SDK clients
type Client struct {
	cfg        aws.Config
	region     string
	s3Client   *S3Client
	cfClient   *CloudFrontClient
}

// NewClient creates a new AWS client with the given configuration
func NewClient(ctx context.Context, region, profile string) (*Client, error) {
	var opts []func(*config.LoadOptions) error

	opts = append(opts, config.WithRegion(region))

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Verify credentials are valid
	if err := verifyCredentials(ctx, cfg); err != nil {
		return nil, err
	}

	client := &Client{
		cfg:    cfg,
		region: region,
	}

	// Initialize S3 client
	s3Svc := s3.NewFromConfig(cfg)
	client.s3Client = &S3Client{
		client: s3Svc,
		region: region,
	}

	// Initialize CloudFront client
	cfSvc := cloudfront.NewFromConfig(cfg)
	client.cfClient = &CloudFrontClient{
		client: cfSvc,
	}

	return client, nil
}

// S3 returns the S3 client
func (c *Client) S3() *S3Client {
	return c.s3Client
}

// CloudFront returns the CloudFront client
func (c *Client) CloudFront() *CloudFrontClient {
	return c.cfClient
}

// Region returns the configured region
func (c *Client) Region() string {
	return c.region
}

// Config returns the underlying AWS config
func (c *Client) Config() aws.Config {
	return c.cfg
}

// ErrNoCredentials is returned when AWS credentials are not configured
var ErrNoCredentials = errors.New("no AWS credentials found")

// verifyCredentials checks if AWS credentials are valid
func verifyCredentials(ctx context.Context, cfg aws.Config) error {
	stsClient := sts.NewFromConfig(cfg)
	_, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("%w: run 'aws configure' or set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables", ErrNoCredentials)
	}
	return nil
}
