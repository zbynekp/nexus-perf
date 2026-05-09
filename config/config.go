package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// RepositoryFormat defines the format of the repository
type RepositoryFormat string

const (
	RepositoryFormatRaw   RepositoryFormat = "RAW"
	RepositoryFormatMaven RepositoryFormat = "MAVEN"
)

// Config holds all configuration for the application
type Config struct {
	// Nexus connection settings
	NexusEndpoint  string
	Username       string
	Password       string
	RepositoryName string
	Format         RepositoryFormat

	// TLS/SSL settings
	CAPath     string
	SkipVerify bool

	// Test parameters
	FileSize   int64 // in bytes
	NumFiles   int
	NumThreads int
	Verbosity  int // 0=error, 1=info, 2=debug, 3=trace

	// Flags
	SkipUpload   bool
	SkipDownload bool
	KeepFiles    bool
}

// New creates a new config from environment variables and Viper
func New() (*Config, error) {
	// Set up Viper to read from environment
	viper.SetEnvPrefix("NEXUS_PERF")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("verbosity", 1)
	viper.SetDefault("num_files", 10)
	viper.SetDefault("num_threads", 4)
	viper.SetDefault("file_size", 1024*1024) // 1MB default
	viper.SetDefault("format", "RAW")
	viper.SetDefault("skip_verify", false)

	cfg := &Config{
		NexusEndpoint:  viper.GetString("nexus_endpoint"),
		Username:       viper.GetString("username"),
		Password:       viper.GetString("password"),
		RepositoryName: viper.GetString("repository_name"),
		Format:         RepositoryFormat(strings.ToUpper(viper.GetString("format"))),
		CAPath:         viper.GetString("ca_path"),
		SkipVerify:     viper.GetBool("skip_verify"),
		FileSize:       viper.GetInt64("file_size"),
		NumFiles:       viper.GetInt("num_files"),
		NumThreads:     viper.GetInt("num_threads"),
		Verbosity:      viper.GetInt("verbosity"),
		SkipUpload:     viper.GetBool("skip_upload"),
		SkipDownload:   viper.GetBool("skip_download"),
		KeepFiles:      viper.GetBool("keep_files"),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.NexusEndpoint == "" {
		return fmt.Errorf("nexus_endpoint is required")
	}

	if c.Username == "" {
		return fmt.Errorf("username is required")
	}

	if c.Password == "" {
		return fmt.Errorf("password is required")
	}

	if c.RepositoryName == "" {
		return fmt.Errorf("repository_name is required")
	}

	if c.Format != RepositoryFormatRaw && c.Format != RepositoryFormatMaven {
		return fmt.Errorf("format must be RAW or MAVEN, got: %s", c.Format)
	}

	if c.FileSize <= 0 {
		return fmt.Errorf("file_size must be greater than 0")
	}

	if c.NumFiles <= 0 {
		return fmt.Errorf("num_files must be greater than 0")
	}

	if c.NumThreads <= 0 {
		return fmt.Errorf("num_threads must be greater than 0")
	}

	if c.Verbosity < 0 || c.Verbosity > 3 {
		return fmt.Errorf("verbosity must be between 0 and 3")
	}

	if c.CAPath != "" {
		// Check if CA file exists
		if _, err := os.Stat(c.CAPath); err != nil {
			return fmt.Errorf("ca_path file not found: %w", err)
		}
	}

	return nil
}

// ApplyFlags applies CLI flags to the configuration
func (c *Config) ApplyFlags(flags map[string]interface{}) error {
	for key, value := range flags {
		switch key {
		case "nexus-endpoint":
			if v, ok := value.(string); ok {
				c.NexusEndpoint = v
			}
		case "username":
			if v, ok := value.(string); ok {
				c.Username = v
			}
		case "password":
			if v, ok := value.(string); ok {
				c.Password = v
			}
		case "repository-name":
			if v, ok := value.(string); ok {
				c.RepositoryName = v
			}
		case "format":
			if v, ok := value.(string); ok {
				c.Format = RepositoryFormat(strings.ToUpper(v))
			}
		case "ca-path":
			if v, ok := value.(string); ok {
				c.CAPath = v
			}
		case "skip-verify":
			if v, ok := value.(bool); ok {
				c.SkipVerify = v
			}
		case "file-size":
			if v, ok := value.(int64); ok {
				c.FileSize = v
			}
		case "num-files":
			if v, ok := value.(int); ok {
				c.NumFiles = v
			}
		case "num-threads":
			if v, ok := value.(int); ok {
				c.NumThreads = v
			}
		case "verbosity":
			if v, ok := value.(int); ok {
				c.Verbosity = v
			}
		case "skip-upload":
			if v, ok := value.(bool); ok {
				c.SkipUpload = v
			}
		case "skip-download":
			if v, ok := value.(bool); ok {
				c.SkipDownload = v
			}
		case "keep-files":
			if v, ok := value.(bool); ok {
				c.KeepFiles = v
			}
		}
	}

	return c.Validate()
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf(
		"NexusEndpoint: %s, Repository: %s (%s), Files: %d, FileSize: %d, Threads: %d, Verbosity: %d",
		c.NexusEndpoint, c.RepositoryName, c.Format, c.NumFiles, c.FileSize, c.NumThreads, c.Verbosity,
	)
}
