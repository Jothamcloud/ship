package cmd

import (
	"fmt"
	"os"

	"github.com/jothamarinze/shipfe/pkg/version"
	"github.com/spf13/cobra"
)

var (
	region  string
	profile string
)

var rootCmd = &cobra.Command{
	Use:     "shipfe",
	Short:   "Deploy frontend apps to AWS with minimal setup",
	Long:    `Shipfe is a CLI tool that deploys frontend applications to AWS with a single command.`,
	Version: version.Version,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&region, "region", "us-east-1", "AWS region")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS profile name")
}

// GetRegion returns the configured AWS region
func GetRegion() string {
	return region
}

// GetProfile returns the configured AWS profile
func GetProfile() string {
	return profile
}
