package cmd

import (
	"fmt"
	"os"

	"github.com/jothamarinze/shipfe/internal/config"
	"github.com/jothamarinze/shipfe/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show deployment status",
	Long:  `Display information about the current deployment.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	printer := ui.NewPrinter()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		printer.Warning("No deployment found in this directory")
		printer.Hint("Run 'shipfe deploy' to deploy your app")
		return nil
	}

	printer.Divider()
	printer.Info("Deployment Status")
	printer.Divider()
	printer.URL("  URL: ", "https://"+cfg.CloudFrontURL)
	fmt.Printf("  Framework:    %s\n", cfg.Framework)
	fmt.Printf("  Build Output: %s\n", cfg.BuildOutputDir)
	fmt.Printf("  Region:       %s\n", cfg.Region)
	if cfg.CreatedAt != "" {
		fmt.Printf("  Created:      %s\n", cfg.CreatedAt)
	}
	printer.Divider()

	return nil
}
