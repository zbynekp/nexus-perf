package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/company/nexus-perf/config"
	"github.com/company/nexus-perf/logger"
	"github.com/company/nexus-perf/nexus"
	"github.com/company/nexus-perf/suite"
)

var (
	appVersion = "dev" // overridden at link time via -X github.com/company/nexus-perf/cmd.appVersion

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
	logFile        string
	verboseMetrics bool

	rootCmd = &cobra.Command{
		Use:     "nexus-perf",
		Short:   "Nexus Repository Performance Test Tool",
		Long:    "Enterprise-grade performance testing tool for Sonatype Nexus Repository",
		Version: appVersion,
	}

	testCmd = &cobra.Command{
		Use:   "test",
		Short: "Run performance tests",
		Long:  "Execute upload and download performance tests against Nexus repository",
		RunE:  runTest,
	}

	uploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "Run upload test only",
		RunE:  runUploadTest,
	}

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
	testCmd.Flags().StringVar(&logFile, "log-file", "", "Path to log file (logs to stdout if not specified)")
	testCmd.Flags().BoolVar(&verboseMetrics, "verbose-metrics", false, "Print per-file metrics (time and speed for each file)")

	// Bind testCmd flags to viper so NEXUS_PERF_* env vars take precedence
	viper.SetEnvPrefix("NEXUS_PERF")
	viper.AutomaticEnv()
	_ = viper.BindPFlag("nexus_endpoint", testCmd.Flags().Lookup("nexus-endpoint"))
	_ = viper.BindPFlag("username", testCmd.Flags().Lookup("username"))
	_ = viper.BindPFlag("password", testCmd.Flags().Lookup("password"))
	_ = viper.BindPFlag("repository_name", testCmd.Flags().Lookup("repository-name"))
	_ = viper.BindPFlag("format", testCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("ca_path", testCmd.Flags().Lookup("ca-path"))
	_ = viper.BindPFlag("skip_verify", testCmd.Flags().Lookup("skip-verify"))
	_ = viper.BindPFlag("file_size", testCmd.Flags().Lookup("file-size"))
	_ = viper.BindPFlag("num_files", testCmd.Flags().Lookup("num-files"))
	_ = viper.BindPFlag("num_threads", testCmd.Flags().Lookup("num-threads"))
	_ = viper.BindPFlag("verbosity", testCmd.Flags().Lookup("verbosity"))
	_ = viper.BindPFlag("skip_upload", testCmd.Flags().Lookup("skip-upload"))
	_ = viper.BindPFlag("skip_download", testCmd.Flags().Lookup("skip-download"))
	_ = viper.BindPFlag("keep_files", testCmd.Flags().Lookup("keep-files"))
	_ = viper.BindPFlag("log_file", testCmd.Flags().Lookup("log-file"))
	_ = viper.BindPFlag("verbose_metrics", testCmd.Flags().Lookup("verbose-metrics"))

	// Copy flags to upload and download subcommands
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
		cmd.Flags().IntVar(&verbosity, "verbosity", 1, "Verbosity level (0=error, 1=info, 2=debug, 3=trace)")
		cmd.Flags().BoolVar(&keepFiles, "keep-files", false, "Keep uploaded files after test")
		cmd.Flags().StringVar(&logFile, "log-file", "", "Path to log file (logs to stdout if not specified)")
		cmd.Flags().BoolVar(&verboseMetrics, "verbose-metrics", false, "Print per-file metrics")
	}
}

func runTest(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load base config from env vars (no validation yet — CLI flags may supply missing values).
	cfg := config.Load()

	// Apply only flags the user explicitly passed. Skipping unset flags prevents cobra
	// defaults from overriding NEXUS_PERF_* environment variables.
	if cmd.Flags().Changed("nexus-endpoint") {
		cfg.NexusEndpoint = nexusEndpoint
	}
	if cmd.Flags().Changed("username") {
		cfg.Username = username
	}
	if cmd.Flags().Changed("password") {
		cfg.Password = password
	}
	if cmd.Flags().Changed("repository-name") {
		cfg.RepositoryName = repositoryName
	}
	if cmd.Flags().Changed("format") {
		cfg.Format = config.RepositoryFormat(strings.ToUpper(format))
	}
	if cmd.Flags().Changed("ca-path") {
		cfg.CAPath = caPath
	}
	if cmd.Flags().Changed("skip-verify") {
		cfg.SkipVerify = skipVerify
	}
	if cmd.Flags().Changed("file-size") {
		cfg.FileSize = fileSize
	}
	if cmd.Flags().Changed("num-files") {
		cfg.NumFiles = numFiles
	}
	if cmd.Flags().Changed("num-threads") {
		cfg.NumThreads = numThreads
	}
	if cmd.Flags().Changed("verbosity") {
		cfg.Verbosity = verbosity
	}
	if cmd.Flags().Changed("keep-files") {
		cfg.KeepFiles = keepFiles
	}
	if cmd.Flags().Changed("log-file") {
		cfg.LogFile = logFile
	}
	if cmd.Flags().Changed("verbose-metrics") {
		cfg.VerboseMetrics = verboseMetrics
	}
	// skip-upload/skip-download are also set programmatically by the upload/download
	// subcommands, so always apply them (OR with any env-var value already loaded).
	cfg.SkipUpload = cfg.SkipUpload || skipUpload
	cfg.SkipDownload = cfg.SkipDownload || skipDownload

	if strings.HasPrefix(cfg.NexusEndpoint, "http://") {
		fmt.Fprintln(os.Stderr, "WARNING: endpoint uses plain HTTP — credentials will be transmitted unencrypted")
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	// Create logger with optional file output
	var (
		log *logger.Logger
		err error
	)
	if cfg.LogFile != "" {
		log, err = logger.NewWithFile(cfg.Verbosity, cfg.LogFile)
	} else {
		log, err = logger.New(cfg.Verbosity)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		return err
	}
	defer func() { _ = log.Sync() }()

	printBanner(cfg)

	// Create Nexus client
	client, err := nexus.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  [FAIL] Could not initialise client: %v\n\n", err)
		return err
	}
	defer client.Close()

	// Pre-flight: connectivity
	fmt.Print("  Checking connectivity ... ")
	if err := client.Ping(ctx); err != nil {
		fmt.Printf("FAIL\n\n  Error: %v\n\n", err)
		return err
	}
	fmt.Println("OK")

	// Pre-flight: credentials + repository
	fmt.Print("  Checking credentials & repository ... ")
	if err := client.CheckRepository(ctx); err != nil {
		fmt.Printf("FAIL\n\n  Error: %v\n\n", err)
		return err
	}
	fmt.Println("OK")

	fmt.Println()
	log.Infof("Pre-flight checks passed, starting test")

	testSuite := suite.NewTestSuite(cfg, client, log)

	if err := testSuite.Prepare(); err != nil {
		log.Fatalf("Failed to prepare test files: %v", err)
		return err
	}

	if !cfg.SkipUpload {
		uploadMetrics, err := testSuite.RunUploadTest(ctx)
		if err != nil {
			log.Errorf("Upload test failed: %v", err)
		} else if cfg.LogFile != "" {
			fmt.Print(uploadMetrics.FormatAsTable())
		} else {
			log.Infof(uploadMetrics.String())
		}
	}

	if !cfg.SkipDownload {
		downloadMetrics, err := testSuite.RunDownloadTest(ctx)
		if err != nil {
			log.Errorf("Download test failed: %v", err)
		} else if cfg.LogFile != "" {
			fmt.Print(downloadMetrics.FormatAsTable())
		} else {
			log.Infof(downloadMetrics.String())
		}
	}

	// Fresh context for cleanup so it always runs even after a Ctrl+C.
	if err := testSuite.Cleanup(context.Background()); err != nil {
		log.Warnf("Cleanup error: %v", err)
	}

	log.Infof("Test completed successfully")
	return nil
}

func printBanner(cfg *config.Config) {
	tests := []string{}
	if !cfg.SkipUpload {
		tests = append(tests, "upload")
	}
	if !cfg.SkipDownload {
		tests = append(tests, "download")
	}
	testList := "upload + download"
	if len(tests) == 1 {
		testList = tests[0]
	}

	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Println("                     Nexus Repository Performance Test")
	fmt.Println("================================================================================")
	fmt.Printf("  Endpoint:      %s\n", cfg.NexusEndpoint)
	fmt.Printf("  Repository:    %s (%s)\n", cfg.RepositoryName, cfg.Format)
	fmt.Printf("  Username:      %s\n", cfg.Username)
	fmt.Printf("  Files:         %d x %.2f MB  (total %.2f MB)\n",
		cfg.NumFiles,
		float64(cfg.FileSize)/1_000_000,
		float64(cfg.FileSize)*float64(cfg.NumFiles)/1_000_000,
	)
	fmt.Printf("  Threads:       %d\n", cfg.NumThreads)
	fmt.Printf("  Tests:         %s\n", testList)
	fmt.Println("================================================================================")
	fmt.Println()
	fmt.Println("Pre-flight checks:")
}

func runUploadTest(cmd *cobra.Command, args []string) error {
	skipDownload = true
	return runTest(cmd, args)
}

func runDownloadTest(cmd *cobra.Command, args []string) error {
	skipUpload = true
	return runTest(cmd, args)
}
