package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jothamarinze/shipfe/internal/aws"
	"github.com/jothamarinze/shipfe/internal/builder"
	"github.com/jothamarinze/shipfe/internal/config"
	"github.com/jothamarinze/shipfe/internal/detector"
	"github.com/jothamarinze/shipfe/internal/ui"
	"github.com/spf13/cobra"
)

var (
	skipBuild bool
	dryRun    bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy frontend app to AWS",
	Long:  `Build and deploy your frontend application to AWS.`,
	RunE:  runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().BoolVar(&skipBuild, "skip-build", false, "Skip the build step")
	deployCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate deployment without creating AWS resources")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	printer := ui.NewPrinter()

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectName := filepath.Base(cwd)
	if dryRun {
		printer.Info("[DRY RUN] Simulating deployment of %s...", projectName)
	} else {
		printer.Info("Deploying %s...", projectName)
	}

	// Check for existing config (redeploy case)
	cfg, err := config.Load(cwd)
	isRedeploy := err == nil && cfg.BucketName != ""

	if isRedeploy {
		printer.Info("Existing deployment found, redeploying...")
	}

	// Detect framework
	printer.Step("Detecting framework...")
	detection, err := detector.Detect(cwd)
	if err != nil {
		return fmt.Errorf("framework detection failed: %w", err)
	}
	printer.Success("Detected: %s (output: %s)", detection.Framework, detection.OutputDir)

	// Run build if not skipped and not static
	if !skipBuild && detection.Framework != detector.FrameworkStatic {
		printer.Step("Building project...")
		b := builder.New(cwd)
		if err := b.Build(detection.BuildCmd); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
		printer.Success("Build completed")
	}

	// Verify output directory exists
	outputPath := filepath.Join(cwd, detection.OutputDir)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("build output directory not found: %s", outputPath)
	}

	// Count files to upload
	fileCount, totalSize, err := countFiles(outputPath)
	if err != nil {
		return fmt.Errorf("failed to count files: %w", err)
	}
	printer.Success("Found %d files (%s) to upload", fileCount, formatSize(totalSize))

	// Dry run: show what would happen and exit
	if dryRun {
		printer.Divider()
		printer.Info("[DRY RUN] Would perform the following actions:")
		fmt.Println()
		if isRedeploy {
			fmt.Printf("  1. Upload %d files\n", fileCount)
			fmt.Println("  2. Invalidate CDN cache")
			fmt.Println()
			printer.Info("Existing URL: https://%s", cfg.CloudFrontURL)
		} else {
			fmt.Println("  1. Create storage")
			fmt.Println("  2. Set up CDN")
			fmt.Printf("  3. Upload %d files\n", fileCount)
			fmt.Println()
			printer.Info("Region: %s", GetRegion())
			printer.Hint("Tip: Run 'shipfe init' first for faster deploys")
		}
		printer.Divider()
		printer.Success("Dry run complete - no resources created")
		return nil
	}

	// Initialize AWS client
	printer.Step("Connecting to AWS...")
	awsClient, err := aws.NewClient(ctx, GetRegion(), GetProfile())
	if err != nil {
		return fmt.Errorf("failed to initialize AWS client: %w", err)
	}
	printer.Success("Connected to AWS (%s)", GetRegion())

	// Create provisioner
	provisioner := aws.NewProvisioner(awsClient, printer)

	var deployURL string
	if isRedeploy {
		// Redeploy: sync files and invalidate cache
		printer.Step("Uploading files...")
		if err := awsClient.S3().SyncDirectory(ctx, outputPath, cfg.BucketName); err != nil {
			return fmt.Errorf("failed to upload files: %w", err)
		}
		printer.Success("Files uploaded")

		printer.Step("Invalidating cache...")
		if err := awsClient.CloudFront().CreateInvalidation(ctx, cfg.DistributionID); err != nil {
			return fmt.Errorf("failed to invalidate cache: %w", err)
		}
		printer.Success("Cache invalidated")

		deployURL = cfg.CloudFrontURL
	} else {
		// First deploy: provision all resources
		result, err := provisioner.Provision(ctx, projectName, outputPath)
		if err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		// Save config
		cfg = &config.Config{
			Region:         GetRegion(),
			BucketName:     result.BucketName,
			DistributionID: result.DistributionID,
			CloudFrontURL:  result.CloudFrontURL,
			OACID:          result.OACID,
			Framework:      string(detection.Framework),
			BuildOutputDir: detection.OutputDir,
		}
		if err := config.Save(cwd, cfg); err != nil {
			printer.Warning("Failed to save config: %v", err)
		}

		deployURL = result.CloudFrontURL
	}

	printer.Divider()
	printer.Success("Deployment complete!")
	printer.URL("URL: ", "https://"+deployURL)
	if !isRedeploy {
		printer.Hint("First deploy - URL may take a minute to be ready")
	}

	return nil
}

// countFiles counts files and total size in a directory
func countFiles(dir string) (int, int64, error) {
	var count int
	var size int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
			size += info.Size()
		}
		return nil
	})

	return count, size, err
}

// formatSize formats bytes into human-readable format
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)

	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
