package data

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/company/nexus-perf/config"
)

// GenerateFilePath generates a file path based on format and index
func GenerateFilePath(format config.RepositoryFormat, index int) string {
	switch format {
	case config.RepositoryFormatRaw:
		return generateRawFilePath(index)
	case config.RepositoryFormatMaven:
		return generateMavenFilePath(index)
	default:
		return fmt.Sprintf("file-%d", index)
	}
}

func generateRawFilePath(index int) string {
	return filepath.Join("test-files", fmt.Sprintf("file-%03d.bin", index))
}

// generateMavenFilePath: groupId/artifactId/version/artifactId-version.jar
func generateMavenFilePath(index int) string {
	groupID := "com.example.test"
	artifactID := fmt.Sprintf("test-artifact-%03d", index)
	version := "1.0.0"

	groupPath := strings.ReplaceAll(groupID, ".", "/")
	return filepath.Join(groupPath, artifactID, version, fmt.Sprintf("%s-%s.jar", artifactID, version))
}
