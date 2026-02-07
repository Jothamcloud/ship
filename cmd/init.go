package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jothamarinze/shipfe/internal/aws"
	"github.com/jothamarinze/shipfe/internal/config"
	"github.com/jothamarinze/shipfe/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AWS resources for deployment",
	Long:  `Set up AWS resources (storage, CDN) for your project. Run this once, then use 'deploy' for fast uploads.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	printer := ui.NewPrinter()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectName := filepath.Base(cwd)

	// Check if already initialized
	if config.Exists(cwd) {
		cfg, err := config.Load(cwd)
		if err == nil {
			printer.Success("Already initialized")
			printer.URL("URL: ", "https://"+cfg.CloudFrontURL)
			printer.Hint("Run 'shipfe deploy' to upload your build")
			return nil
		}
	}

	printer.Info("Initializing %s...", projectName)

	// Initialize AWS client
	printer.Step("Connecting to AWS...")
	awsClient, err := aws.NewClient(ctx, GetRegion(), GetProfile())
	if err != nil {
		return fmt.Errorf("failed to connect to AWS: %w", err)
	}
	printer.Success("Connected to AWS (%s)", GetRegion())

	// Provision resources
	provisioner := aws.NewProvisioner(awsClient, printer)
	result, err := provisioner.Init(ctx, projectName)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Save config
	cfg := &config.Config{
		Region:         GetRegion(),
		BucketName:     result.BucketName,
		DistributionID: result.DistributionID,
		CloudFrontURL:  result.CloudFrontURL,
		OACID:          result.OACID,
	}
	if err := config.Save(cwd, cfg); err != nil {
		printer.Warning("Failed to save config: %v", err)
	}

	printer.Divider()
	printer.Success("Initialization complete!")
	printer.URL("URL: ", "https://"+result.CloudFrontURL)
	printer.Hint("URL will be ready in ~2 minutes. Run 'shipfe deploy' to upload your build.")

	return nil
}
