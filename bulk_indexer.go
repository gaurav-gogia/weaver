package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
)

// VulnerabilityMetadata contains metadata about a code vulnerability
type VulnerabilityMetadata struct {
	Language         string
	VulnType         string
	CWE              string
	CWEDescription   string
	CVE              string
	CVSSScore        float64
	CVSSVector       string
	Function         string
	FilePath         string
	Library          string
	Severity         string
	ExploitAvailable bool
	PatchAvailable   bool
	AffectedVersion  string
	FixedVersion     string
	SourceRepo       string
	CommitHash       string
	AuditTool        string
	Auditor          string
}

// IndexCodeFile reads a code file, generates its vector embedding, and stores it in Weaviate
func IndexCodeFile(client *weaviate.Client, filePath string, schema string, metadata VulnerabilityMetadata) error {
	// Read the code file
	code, err := readCodeFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	// Generate vector embedding
	vector, err := getVectorFromPython(code)
	if err != nil {
		return fmt.Errorf("failed to generate vector for %s: %w", filePath, err)
	}

	// Index in Weaviate
	_, err = client.Data().Creator().
		WithClassName(schema).
		WithProperties(map[string]interface{}{
			"code":             code,
			"language":         metadata.Language,
			"vulnType":         metadata.VulnType,
			"cwe":              metadata.CWE,
			"cweDescription":   metadata.CWEDescription,
			"cve":              metadata.CVE,
			"cvssScore":        metadata.CVSSScore,
			"cvssVector":       metadata.CVSSVector,
			"function":         metadata.Function,
			"filePath":         metadata.FilePath,
			"library":          metadata.Library,
			"severity":         metadata.Severity,
			"exploitAvailable": metadata.ExploitAvailable,
			"patchAvailable":   metadata.PatchAvailable,
			"affectedVersion":  metadata.AffectedVersion,
			"fixedVersion":     metadata.FixedVersion,
			"sourceRepo":       metadata.SourceRepo,
			"commitHash":       metadata.CommitHash,
			"auditTool":        metadata.AuditTool,
			"auditor":          metadata.Auditor,
		}).
		WithVector(vector).
		Do(context.Background())

	if err != nil {
		return fmt.Errorf("failed to index %s: %w", filePath, err)
	}

	fmt.Printf("✓ Indexed: %s (%d bytes, %s)\n", filePath, len(code), metadata.VulnType)
	return nil
}

// IndexDirectory recursively indexes all code files in a directory
func IndexDirectory(client *weaviate.Client, dirPath string, schema string) error {
	var indexedCount, errorCount int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Detect language from extension
		ext := strings.ToLower(filepath.Ext(path))
		language := detectLanguage(ext)
		if language == "" {
			return nil // Skip non-code files
		}

		// Infer vulnerability type from filename (simple heuristic)
		vulnType := inferVulnType(filepath.Base(path))

		// Create metadata
		metadata := VulnerabilityMetadata{
			Language:         language,
			VulnType:         vulnType,
			CWE:              "CWE-Unknown",
			CWEDescription:   "Automatically indexed code snippet",
			Severity:         "Unknown",
			FilePath:         path,
			ExploitAvailable: false,
			PatchAvailable:   false,
			AuditTool:        "automated-indexer",
			Auditor:          "weaver-system",
		}

		// Index the file
		if err := IndexCodeFile(client, path, schema, metadata); err != nil {
			fmt.Printf("✗ Error indexing %s: %v\n", path, err)
			errorCount++
			return nil // Continue processing other files
		}

		indexedCount++
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("\nIndexing complete: %d files indexed, %d errors\n", indexedCount, errorCount)
	return nil
}

// detectLanguage returns the programming language based on file extension
func detectLanguage(ext string) string {
	languageMap := map[string]string{
		".c":    "C",
		".h":    "C",
		".cpp":  "C++",
		".cc":   "C++",
		".cxx":  "C++",
		".hpp":  "C++",
		".py":   "Python",
		".go":   "Go",
		".java": "Java",
		".js":   "JavaScript",
		".ts":   "TypeScript",
		".rs":   "Rust",
		".rb":   "Ruby",
		".php":  "PHP",
		".cs":   "C#",
		".sh":   "Shell",
		".sql":  "SQL",
	}

	return languageMap[ext]
}

// inferVulnType attempts to infer vulnerability type from filename
func inferVulnType(filename string) string {
	lower := strings.ToLower(filename)

	if strings.Contains(lower, "buffer") || strings.Contains(lower, "overflow") {
		return "Buffer Overflow"
	}
	if strings.Contains(lower, "sql") || strings.Contains(lower, "injection") {
		return "SQL Injection"
	}
	if strings.Contains(lower, "xss") || strings.Contains(lower, "cross_site") {
		return "Cross-Site Scripting (XSS)"
	}
	if strings.Contains(lower, "path") || strings.Contains(lower, "traversal") {
		return "Path Traversal"
	}
	if strings.Contains(lower, "csrf") {
		return "Cross-Site Request Forgery (CSRF)"
	}
	if strings.Contains(lower, "rce") || strings.Contains(lower, "remote_exec") {
		return "Remote Code Execution"
	}

	return "Unknown Vulnerability"
}
