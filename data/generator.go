package data

import (
	"crypto/rand"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/company/nexus-perf/config"
)

// RandomData generates random bytes
func RandomData(size int64) []byte {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return data
}

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

// generateRawFilePath generates a raw repository file path
func generateRawFilePath(index int) string {
	// Example: test-files/file-001.bin
	return filepath.Join("test-files", fmt.Sprintf("file-%03d.bin", index))
}

// generateMavenFilePath generates a Maven repository file path
// Maven format: groupId/artifactId/version/artifactId-version.jar
func generateMavenFilePath(index int) string {
	groupID := "com.example.test"
	artifactID := fmt.Sprintf("test-artifact-%03d", index)
	version := "1.0.0"

	groupPath := strings.ReplaceAll(groupID, ".", "/")
	return filepath.Join(groupPath, artifactID, version, fmt.Sprintf("%s-%s.jar", artifactID, version))
}

// GenerateRandomFile creates a file with random content and returns its content
func GenerateRandomFile(size int64) []byte {
	return RandomData(size)
}

// GenerateMavenMetadata generates basic Maven metadata XML
// This would be a pom.xml or similar for Maven repos
func GenerateMavenMetadata(groupID, artifactID, version string) string {
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>%s</groupId>
  <artifactId>%s</artifactId>
  <version>%s</version>
  <packaging>jar</packaging>
  <name>Test Artifact</name>
  <description>Test artifact for Nexus performance testing</description>
</project>`, groupID, artifactID, version)
	return xml
}

// GetContentType returns the appropriate content type based on format and file extension
func GetContentType(format config.RepositoryFormat, filename string) string {
	switch format {
	case config.RepositoryFormatRaw:
		return "application/octet-stream"
	case config.RepositoryFormatMaven:
		if strings.HasSuffix(filename, ".jar") {
			return "application/java-archive"
		}
		if strings.HasSuffix(filename, ".pom") || strings.HasSuffix(filename, ".xml") {
			return "application/xml"
		}
		return "application/octet-stream"
	default:
		return "application/octet-stream"
	}
}
