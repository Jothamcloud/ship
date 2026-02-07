package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// CloudFrontClient wraps CloudFront operations
type CloudFrontClient struct {
	client *cloudfront.Client
}

// OACResult contains the created OAC info
type OACResult struct {
	ID   string
	Name string
}

// DistributionResult contains the created distribution info
type DistributionResult struct {
	ID         string
	DomainName string
	ARN        string
}

// CreateOAC creates an Origin Access Control for S3
func (c *CloudFrontClient) CreateOAC(ctx context.Context, name string) (*OACResult, error) {
	input := &cloudfront.CreateOriginAccessControlInput{
		OriginAccessControlConfig: &types.OriginAccessControlConfig{
			Name:                          aws.String(name),
			Description:                   aws.String("OAC for " + name),
			OriginAccessControlOriginType: types.OriginAccessControlOriginTypesS3,
			SigningBehavior:               types.OriginAccessControlSigningBehaviorsAlways,
			SigningProtocol:               types.OriginAccessControlSigningProtocolsSigv4,
		},
	}

	result, err := c.client.CreateOriginAccessControl(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAC: %w", err)
	}

	return &OACResult{
		ID:   *result.OriginAccessControl.Id,
		Name: *result.OriginAccessControl.OriginAccessControlConfig.Name,
	}, nil
}

// CreateDistribution creates a CloudFront distribution for an S3 bucket
func (c *CloudFrontClient) CreateDistribution(ctx context.Context, bucketDomain, oacID, callerRef string) (*DistributionResult, error) {
	originID := "S3Origin"

	input := &cloudfront.CreateDistributionInput{
		DistributionConfig: &types.DistributionConfig{
			CallerReference: aws.String(callerRef),
			Comment:         aws.String("Created by shipfe"),
			Enabled:         aws.Bool(true),
			DefaultRootObject: aws.String("index.html"),
			Origins: &types.Origins{
				Quantity: aws.Int32(1),
				Items: []types.Origin{
					{
						Id:         aws.String(originID),
						DomainName: aws.String(bucketDomain),
						OriginAccessControlId: aws.String(oacID),
						S3OriginConfig: &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String(originID),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps,
				AllowedMethods: &types.AllowedMethods{
					Quantity: aws.Int32(2),
					Items:    []types.Method{types.MethodGet, types.MethodHead},
					CachedMethods: &types.CachedMethods{
						Quantity: aws.Int32(2),
						Items:    []types.Method{types.MethodGet, types.MethodHead},
					},
				},
				Compress: aws.Bool(true),
				// Use CachingOptimized managed policy
				CachePolicyId: aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
				// ForwardedValues is not needed when using CachePolicyId
			},
			// Custom error responses for SPA routing (404 -> index.html)
			CustomErrorResponses: &types.CustomErrorResponses{
				Quantity: aws.Int32(1),
				Items: []types.CustomErrorResponse{
					{
						ErrorCode:          aws.Int32(404),
						ResponseCode:       aws.String("200"),
						ResponsePagePath:   aws.String("/index.html"),
						ErrorCachingMinTTL: aws.Int64(10),
					},
				},
			},
			PriceClass: types.PriceClassPriceClass100,
			HttpVersion: types.HttpVersionHttp2,
		},
	}

	result, err := c.client.CreateDistribution(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create distribution: %w", err)
	}

	return &DistributionResult{
		ID:         *result.Distribution.Id,
		DomainName: *result.Distribution.DomainName,
		ARN:        *result.Distribution.ARN,
	}, nil
}

// WaitForDistributionDeployed waits for a distribution to be fully deployed
func (c *CloudFrontClient) WaitForDistributionDeployed(ctx context.Context, distributionID string) error {
	waiter := cloudfront.NewDistributionDeployedWaiter(c.client)

	err := waiter.Wait(ctx, &cloudfront.GetDistributionInput{
		Id: aws.String(distributionID),
	}, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("timeout waiting for deployment: %w", err)
	}

	return nil
}

// CreateInvalidation invalidates the CloudFront cache
func (c *CloudFrontClient) CreateInvalidation(ctx context.Context, distributionID string) error {
	input := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(distributionID),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(fmt.Sprintf("shipfe-%d", time.Now().Unix())),
			Paths: &types.Paths{
				Quantity: aws.Int32(1),
				Items:    []string{"/*"},
			},
		},
	}

	_, err := c.client.CreateInvalidation(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create invalidation: %w", err)
	}

	return nil
}

// DisableDistribution disables a CloudFront distribution
func (c *CloudFrontClient) DisableDistribution(ctx context.Context, distributionID string) error {
	// Get current config
	getResult, err := c.client.GetDistributionConfig(ctx, &cloudfront.GetDistributionConfigInput{
		Id: aws.String(distributionID),
	})
	if err != nil {
		return fmt.Errorf("failed to get distribution config: %w", err)
	}

	// Check if already disabled
	if !*getResult.DistributionConfig.Enabled {
		return nil
	}

	// Disable it
	getResult.DistributionConfig.Enabled = aws.Bool(false)

	_, err = c.client.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
		Id:                 aws.String(distributionID),
		DistributionConfig: getResult.DistributionConfig,
		IfMatch:            getResult.ETag,
	})
	if err != nil {
		return fmt.Errorf("failed to disable distribution: %w", err)
	}

	return nil
}

// WaitForDistributionDisabled waits for a distribution to be fully disabled
func (c *CloudFrontClient) WaitForDistributionDisabled(ctx context.Context, distributionID string) error {
	waiter := cloudfront.NewDistributionDeployedWaiter(c.client)

	err := waiter.Wait(ctx, &cloudfront.GetDistributionInput{
		Id: aws.String(distributionID),
	}, 15*time.Minute)
	if err != nil {
		return fmt.Errorf("timeout waiting for distribution to be disabled: %w", err)
	}

	return nil
}

// DeleteDistribution deletes a CloudFront distribution (must be disabled first)
func (c *CloudFrontClient) DeleteDistribution(ctx context.Context, distributionID string) error {
	// Get ETag
	getResult, err := c.client.GetDistribution(ctx, &cloudfront.GetDistributionInput{
		Id: aws.String(distributionID),
	})
	if err != nil {
		return fmt.Errorf("failed to get distribution: %w", err)
	}

	_, err = c.client.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
		Id:      aws.String(distributionID),
		IfMatch: getResult.ETag,
	})
	if err != nil {
		return fmt.Errorf("failed to delete distribution: %w", err)
	}

	return nil
}

// DeleteOAC deletes an Origin Access Control
func (c *CloudFrontClient) DeleteOAC(ctx context.Context, oacID string) error {
	// Get ETag
	getResult, err := c.client.GetOriginAccessControl(ctx, &cloudfront.GetOriginAccessControlInput{
		Id: aws.String(oacID),
	})
	if err != nil {
		return fmt.Errorf("failed to get OAC: %w", err)
	}

	_, err = c.client.DeleteOriginAccessControl(ctx, &cloudfront.DeleteOriginAccessControlInput{
		Id:      aws.String(oacID),
		IfMatch: getResult.ETag,
	})
	if err != nil {
		return fmt.Errorf("failed to delete OAC: %w", err)
	}

	return nil
}
