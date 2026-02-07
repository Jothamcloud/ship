package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jothamarinze/shipfe/internal/aws"
	"github.com/jothamarinze/shipfe/internal/config"
	"github.com/jothamarinze/shipfe/internal/ui"
	"github.com/spf13/cobra"
)

var (
	yesFlag           bool
	backgroundCleanup bool
	cleanupDistID     string
	cleanupOACID      string
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy deployed resources",
	Long:  `Remove all AWS resources created by shipfe for this project.`,
	RunE:  runDestroy,
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
	destroyCmd.Flags().BoolVar(&backgroundCleanup, "background-cleanup", false, "Run background cleanup (internal use)")
	destroyCmd.Flags().StringVar(&cleanupDistID, "cleanup-dist-id", "", "Distribution ID for background cleanup")
	destroyCmd.Flags().StringVar(&cleanupOACID, "cleanup-oac-id", "", "OAC ID for background cleanup")
	destroyCmd.Flags().MarkHidden("background-cleanup")
	destroyCmd.Flags().MarkHidden("cleanup-dist-id")
	destroyCmd.Flags().MarkHidden("cleanup-oac-id")
}

func runDestroy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	printer := ui.NewPrinter()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Background cleanup mode: handle slow CloudFront deletion
	if backgroundCleanup {
		return runBackgroundCleanup(ctx)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		printer.Warning("No deployment found in this directory")
		return nil
	}

	// Confirmation prompt
	if !yesFlag {
		printer.Warning("This will permanently delete your deployment at:")
		fmt.Printf("  https://%s\n", cfg.CloudFrontURL)
		fmt.Println()
		fmt.Print("Are you sure? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			printer.Info("Destroy cancelled")
			return nil
		}
	}

	// Initialize AWS client
	awsClient, err := aws.NewClient(ctx, cfg.Region, GetProfile())
	if err != nil {
		return fmt.Errorf("failed to connect to AWS: %w", err)
	}

	// Quick cleanup: delete S3 contents and bucket
	awsClient.S3().DeleteBucketContents(ctx, cfg.BucketName)
	awsClient.S3().DeleteBucket(ctx, cfg.BucketName)

	// Disable CloudFront (quick API call)
	awsClient.CloudFront().DisableDistribution(ctx, cfg.DistributionID)

	// Spawn background process for slow CloudFront cleanup
	spawnBackgroundCleanup(cfg.Region, GetProfile(), cfg.DistributionID, cfg.OACID)

	// Remove config directory immediately
	config.Remove(cwd)

	printer.Success("Deployment destroyed")

	return nil
}

// spawnBackgroundCleanup spawns a detached process to handle slow CloudFront deletion
func spawnBackgroundCleanup(region, profile, distID, oacID string) {
	exe, err := os.Executable()
	if err != nil {
		return
	}

	args := []string{"destroy", "--background-cleanup", "--region", region, "--cleanup-dist-id", distID}
	if oacID != "" {
		args = append(args, "--cleanup-oac-id", oacID)
	}
	if profile != "" {
		args = append(args, "--profile", profile)
	}

	cmd := exec.Command(exe, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Start detached - don't wait for it
	cmd.Start()
}

// runBackgroundCleanup handles the slow CloudFront deletion in background
func runBackgroundCleanup(ctx context.Context) error {
	if cleanupDistID == "" {
		return nil
	}

	awsClient, err := aws.NewClient(ctx, GetRegion(), GetProfile())
	if err != nil {
		return err
	}

	// Wait for distribution to be disabled, then delete it
	if err := awsClient.CloudFront().WaitForDistributionDisabled(ctx, cleanupDistID); err == nil {
		awsClient.CloudFront().DeleteDistribution(ctx, cleanupDistID)
	}

	// Clean up OAC
	if cleanupOACID != "" {
		awsClient.CloudFront().DeleteOAC(ctx, cleanupOACID)
	}

	return nil
}
