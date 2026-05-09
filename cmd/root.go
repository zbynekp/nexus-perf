package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/company/nexus-perf/config"
	"github.com/company/nexus-perf/logger"
	"github.com/company/nexus-perf/nexus"
	"github.com/company/nexus-perf/suite"
)

var (
	// Global flags
	nexusEndpoint  string
	username       string
	password       string
	repositoryName string
	format         string
	caPath         string
	skipVerify     bool
	fileSize       int64
	numFiles       int
	numThreads     int
	verbosity      int
	skipUpload     bool
	skipDownload   bool
	keepFiles      bool

	// Root command
	rootCmd = &cobra.Command{
		Use:     "nexus-perf",
		Short:   "Nexus Repository Performance Test Tool",
		Long:    "Enterprise-grade performance testing tool for Sonatype Nexus Repository",
		Version: "1.0.0",
	}

	// Test command
	testCmd = &cobra.Command{
		Use:   "test",
		Short: "Run performance tests",
		Long:  "Execute upload and download performance tests against Nexus repository",
		RunE:  runTest,
	}

	// Upload only command
	uploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Run upload test only",
		RunE:  runUploadTest,
	}

	// Download only command
	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Run download test only",
		RunE:  runDownloadTest,
	}
)

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(testCmd, uploadCmd, downloadCmd)

	// Define flags
	testCmd.Flags().StringVar(&nexusEndpoint, "nexus-endpoint", "", "Nexus repository endpoint URL")
	testCmd.Flags().StringVar(&username, "username", "", "Nexus username")
	testCmd.Flags().StringVar(&password, "password", "", "Nexus password")
	testCmd.Flags().StringVar(&repositoryName, "repository-name", "", "Repository name")
	testCmd.Flags().StringVar(&format, "format", "RAW", "Repository format (RAW|MAVEN)")
	testCmd.Flags().StringVar(&caPath, "ca-path", "", "Path to custom CA certificate")
	testCmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Skip TLS certificate verification")
	testCmd.Flags().Int64Var(&fileSize, "file-size", 1024*1024, "Test file size in bytes (default: 1MB)")
	testCmd.Flags().IntVar(&numFiles, "num-files", 10, "Number of test files")
	testCmd.Flags().IntVar(&numThreads, "num-threads", 4, "Number of concurrent threads")
	testCmd.Flags().IntVar(&verbosity, "verbosity", 1, "Verbosity level (0=error, 1=info, 2=debug, 3=trace)")
	testCmd.Flags().BoolVar(&skipUpload, "skip-upload", false, "Skip upload test")
	testCmd.Flags().BoolVar(&skipDownload, "skip-download", false, "Skip download test")
	testCmd.Flags().BoolVar(&keepFiles, "keep-files", false, "Keep uploaded files after test")

	// Make environment variables override CLI flags
	viper.SetEnvPrefix("NEXUS_PERF")
	viper.AutomaticEnv()
	viper.BindPFlag("nexus_endpoint", testCmd.Flags().Lookup("nexus-endpoint"))
	viper.BindPFlag("username", testCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", testCmd.Flags().Lookup("password"))
	viper.BindPFlag("repository_name", testCmd.Flags().Lookup("repository-name"))
	viper.BindPFlag("format", testCmd.Flags().Lookup("format"))
	viper.BindPFlag("ca_path", testCmd.Flags().Lookup("ca-path"))
	viper.BindPFlag("skip_verify", testCmd.Flags().Lookup("skip-verify"))
	viper.BindPFlag("file_size", testCmd.Flags().Lookup("file-size"))
	viper.BindPFlag("num_files", testCmd.Flags().Lookup("num-files"))
	viper.BindPFlag("num_threads", testCmd.Flags().Lookup("num-threads"))
	viper.BindPFlag("verbosity", testCmd.Flags().Lookup("verbosity"))
	viper.BindPFlag("skip_upload", testCmd.Flags().Lookup("skip-upload"))
	viper.BindPFlag("skip_download", testCmd.Flags().Lookup("skip-download"))
	viper.BindPFlag("keep_files", testCmd.Flags().Lookup("keep-files"))

	// Copy flags to other commands
	for _, cmd := range []*cobra.Command{uploadCmd, downloadCmd} {
		cmd.Flags().StringVar(&nexusEndpoint, "nexus-endpoint", "", "Nexus repository endpoint URL")
		cmd.Flags().StringVar(&username, "username", "", "Nexus username")
		cmd.Flags().StringVar(&password, "password", "", "Nexus password")
		cmd.Flags().StringVar(&repositoryName, "repository-name", "", "Repository name")
		cmd.Flags().StringVar(&format, "format", "RAW", "Repository format (RAW|MAVEN)")
		cmd.Flags().StringVar(&caPath, "ca-path", "", "Path to custom CA certificate")
		cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Skip TLS certificate verification")
		cmd.Flags().Int64Var(&fileSize, "file-size", 1024*1024, "Test file size in bytes")
		cmd.Flags().IntVar(&numFiles, "num-files", 10, "Number of test files")
		cmd.Flags().IntVar(&numThreads, "num-threads", 4, "Number of concurrent threads")
		cmd.Flags().IntVar(&verbosity, "verbosity", 1, "Verbosity level (0=error, 1=info, 2=debug)")
		cmd.Flags().BoolVar(&keepFiles, "keep-files", false, "Keep uploaded files after test")
	}
}

func runTest(cmd *cobra.Command, args []string) error {
	// Load configuration from environment and flags
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid configuration: %v\n", err)
		return err
	}

	// Apply CLI flags if provided
	flags := map[string]interface{}{
		"nexus-endpoint":  nexusEndpoint,
		"username":        username,
		"password":        password,
		"repository-name": repositoryName,
		"format":          format,
		"ca-path":         caPath,
		"skip-verify":     skipVerify,
		"file-size":       fileSize,
		"num-files":       numFiles,
		"num-threads":     numThreads,
		"verbosity":       verbosity,
		"skip-upload":     skipUpload,
		"skip-download":   skipDownload,
		"keep-files":      keepFiles,
	}

	// Filter out empty flags
	for k, v := range flags {
		switch v := v.(type) {
		case string:
			if v == "" {
				delete(flags, k)
			}
		}
	}

	if err := cfg.ApplyFlags(flags); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid configuration: %v\n", err)
		return err
	}

	// Create logger
	log := logger.New(cfg.Verbosity)
	defer log.Sync()

	log.Infof("Starting Nexus Performance Test")
	log.Infof("Configuration: %s", cfg)

	// Create Nexus client
	client, err := nexus.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Nexus client: %v", err)
		return err
	}
	defer client.Close()

	// Create test suite
	testSuite := suite.NewTestSuite(cfg, client, log)

	// Prepare test files
	if err := testSuite.Prepare(); err != nil {
		log.Fatalf("Failed to prepare test files: %v", err)
		return err
	}

	// Run upload test
	if !cfg.SkipUpload {
		uploadMetrics, err := testSuite.RunUploadTest()
		if err != nil {
			log.Errorf("Upload test failed: %v", err)
		} else {
			log.Infof(uploadMetrics.String())
		}
	}

	// Run download test
	if !cfg.SkipDownload {
		downloadMetrics, err := testSuite.RunDownloadTest()
		if err != nil {
			log.Errorf("Download test failed: %v", err)
		} else {
			log.Infof(downloadMetrics.String())
		}
	}

	// Cleanup
	if err := testSuite.Cleanup(); err != nil {
		log.Warnf("Cleanup error: %v", err)
	}

	log.Infof("Test completed successfully")
	return nil
}

func runUploadTest(cmd *cobra.Command, args []string) error {
	skipDownload = true
	return runTest(cmd, args)
}

func runDownloadTest(cmd *cobra.Command, args []string) error {
	skipUpload = true
	return runTest(cmd, args)
}
