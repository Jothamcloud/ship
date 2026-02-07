package aws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jothamarinze/shipfe/internal/config"
	"github.com/jothamarinze/shipfe/internal/ui"
)

// ProvisionResult contains the results of provisioning
type ProvisionResult struct {
	BucketName     string
	DistributionID string
	CloudFrontURL  string
	OACID          string
}

// Provisioner orchestrates AWS resource creation
type Provisioner struct {
	client  *Client
	printer *ui.Printer
}

// NewProvisioner creates a new Provisioner
func NewProvisioner(client *Client, printer *ui.Printer) *Provisioner {
	return &Provisioner{
		client:  client,
		printer: printer,
	}
}

// Init creates AWS resources (no waiting - CDN deploys in background on AWS)
func (p *Provisioner) Init(ctx context.Context, projectName string) (*ProvisionResult, error) {
	bucketName := generateBucketName(projectName)
	oacName := "shipfe-" + bucketName

	// Step 1: Create storage
	p.printer.Step("Creating storage...")
	if err := p.client.S3().CreateBucket(ctx, bucketName); err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	p.printer.Success("Storage created")

	// Step 2: Create access control
	oac, err := p.client.CloudFront().CreateOAC(ctx, oacName)
	if err != nil {
		return nil, fmt.Errorf("failed to configure access: %w", err)
	}

	// Step 3: Create CDN distribution
	bucketDomain := p.client.S3().GetBucketRegionalDomain(bucketName)
	callerRef := fmt.Sprintf("shipfe-%s-%d", projectName, time.Now().Unix())

	p.printer.Step("Setting up CDN...")
	dist, err := p.client.CloudFront().CreateDistribution(ctx, bucketDomain, oac.ID, callerRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDN: %w", err)
	}

	// Step 4: Configure access policy
	if err := p.client.S3().SetBucketPolicy(ctx, bucketName, dist.ARN); err != nil {
		return nil, fmt.Errorf("failed to configure access: %w", err)
	}
	p.printer.Success("CDN configured")

	// CDN will finish deploying in background on AWS side

	return &ProvisionResult{
		BucketName:     bucketName,
		DistributionID: dist.ID,
		CloudFrontURL:  dist.DomainName,
		OACID:          oac.ID,
	}, nil
}

// Provision creates AWS resources and uploads files (no waiting)
func (p *Provisioner) Provision(ctx context.Context, projectName, outputDir string) (*ProvisionResult, error) {
	bucketName := generateBucketName(projectName)
	oacName := "shipfe-" + bucketName

	// Step 1: Create storage
	p.printer.Step("Creating storage...")
	if err := p.client.S3().CreateBucket(ctx, bucketName); err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	p.printer.Success("Storage created")

	// Step 2: Create access control
	oac, err := p.client.CloudFront().CreateOAC(ctx, oacName)
	if err != nil {
		return nil, fmt.Errorf("failed to configure access: %w", err)
	}

	// Step 3: Create CDN distribution
	bucketDomain := p.client.S3().GetBucketRegionalDomain(bucketName)
	callerRef := fmt.Sprintf("shipfe-%s-%d", projectName, time.Now().Unix())

	p.printer.Step("Setting up CDN...")
	dist, err := p.client.CloudFront().CreateDistribution(ctx, bucketDomain, oac.ID, callerRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDN: %w", err)
	}

	// Step 4: Configure access policy
	if err := p.client.S3().SetBucketPolicy(ctx, bucketName, dist.ARN); err != nil {
		return nil, fmt.Errorf("failed to configure access: %w", err)
	}
	p.printer.Success("CDN configured")

	// Step 5: Upload files (no waiting for CDN)
	p.printer.Step("Uploading files...")
	if err := p.client.S3().SyncDirectory(ctx, outputDir, bucketName); err != nil {
		return nil, fmt.Errorf("failed to upload files: %w", err)
	}
	p.printer.Success("Files uploaded")

	return &ProvisionResult{
		BucketName:     bucketName,
		DistributionID: dist.ID,
		CloudFrontURL:  dist.DomainName,
		OACID:          oac.ID,
	}, nil
}

// Destroy removes AWS resources quickly and returns (no waiting)
func (p *Provisioner) Destroy(ctx context.Context, cfg *config.Config) error {
	// Delete S3 contents and bucket (quick)
	p.client.S3().DeleteBucketContents(ctx, cfg.BucketName)
	p.client.S3().DeleteBucket(ctx, cfg.BucketName)

	// Disable CDN (quick API call, returns immediately)
	// The distribution stays disabled - doesn't cost anything
	p.client.CloudFront().DisableDistribution(ctx, cfg.DistributionID)

	return nil
}

// generateBucketName creates a unique bucket name from the project name
func generateBucketName(projectName string) string {
	// Sanitize project name for S3 bucket naming rules
	name := strings.ToLower(projectName)
	name = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")

	// Truncate if too long (max 63 chars for S3, we reserve space for prefix and suffix)
	if len(name) > 30 {
		name = name[:30]
	}

	// Generate random suffix
	suffix := randomHex(6)

	return fmt.Sprintf("shipfe-%s-%s", name, suffix)
}

// randomHex generates a random hex string of the given length
func randomHex(n int) string {
	bytes := make([]byte, n/2+1)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:n]
}
