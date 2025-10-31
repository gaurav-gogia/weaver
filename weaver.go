package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

// Request body for embedding service
type EmbedRequest struct {
	Text string `json:"text"`
}

// Response body from embedding service
type EmbedResponse struct {
	Vector []float64 `json:"vector"` // Use float64 because JSON unmarshals floats as float64
}

// getVectorFromPython sends code to the Python embedding service (sentence-transformers)
// We use intfloat/e5-base-v2 which provides:
// - Semantic understanding of code (not just keyword matching)
// - 768-dimensional vectors optimized for similarity search
// - Efficient CPU-based inference (~400MB model)
// - Better than simple string encoding while avoiding LLM overhead
func getVectorFromPython(code string) ([]float32, error) {
	reqBody := EmbedRequest{Text: code}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:5005/embed", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	vec := make([]float32, len(result.Vector))
	for i, v := range result.Vector {
		vec[i] = float32(v)
	}

	return vec, nil
}

// readCodeFromFile reads the entire content of a file
func readCodeFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return string(content), nil
}

func main() {
	// Command-line flags
	fileFlag := flag.String("file", "", "Index a single code file")
	dirFlag := flag.String("dir", "examples", "Index all code files in a directory")
	searchFlag := flag.String("search", "", "Search for similar vulnerabilities using a code file")
	modeFlag := flag.String("mode", "dir", "Operation mode: 'file', 'dir', or 'search'")
	flag.Parse()

	fmt.Println("🚀 Weaver - Code Vulnerability Vector Database")
	fmt.Println("================================================")

	config := weaviate.Config{
		Scheme: "http",
		Host:   "localhost:8080",
	}

	client, err := weaviate.NewClient(config)
	handle(err)

	metaGetter := client.Misc().MetaGetter()
	meta, err := metaGetter.Do(context.Background())
	handle(err)

	fmt.Printf("\n📊 Weaviate: %s (v%s)\n\n", meta.Hostname, meta.Version)

	const SCHEMA = "CodeSnippet10"

	// Ensure schema exists
	ok, err := classExists(client, SCHEMA)
	handle(err)
	if !ok {
		fmt.Println("📝 Creating schema...")
		schema := &models.Class{
			Class:      SCHEMA,
			Vectorizer: "none", // Using custom sentence-transformers (intfloat/e5-base-v2)
			Properties: []*models.Property{
				{Name: "code", DataType: []string{"text"}},     // the actual code
				{Name: "language", DataType: []string{"text"}}, // e.g., C, Python, etc.
				{Name: "vulnType", DataType: []string{"text"}}, // e.g., Buffer Overflow

				{Name: "cwe", DataType: []string{"text"}},            // e.g., CWE-119
				{Name: "cweDescription", DataType: []string{"text"}}, // brief CWE explanation
				{Name: "cve", DataType: []string{"text"}},            // e.g., CVE-2023-XXXX
				{Name: "cvssScore", DataType: []string{"number"}},    // e.g., 7.8
				{Name: "cvssVector", DataType: []string{"text"}},     // e.g., AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H

				{Name: "function", DataType: []string{"text"}}, // function name
				{Name: "filePath", DataType: []string{"text"}}, // location in source
				{Name: "library", DataType: []string{"text"}},  // e.g., libc, openssl

				{Name: "severity", DataType: []string{"text"}},            // e.g., High, Medium, Low
				{Name: "exploitAvailable", DataType: []string{"boolean"}}, // true/false
				{Name: "patchAvailable", DataType: []string{"boolean"}},   // true/false
				{Name: "affectedVersion", DataType: []string{"text"}},     // e.g., <1.2.3
				{Name: "fixedVersion", DataType: []string{"text"}},        // e.g., >=1.2.4

				{Name: "sourceRepo", DataType: []string{"text"}}, // GitHub/GitLab/etc.
				{Name: "commitHash", DataType: []string{"text"}}, // reference commit
				{Name: "auditTool", DataType: []string{"text"}},  // e.g., semgrep, snyk
				{Name: "auditor", DataType: []string{"text"}},    // analyst or system name
			},
		}
		err = client.Schema().ClassCreator().WithClass(schema).Do(context.Background())
		handle(err)
		fmt.Println("✓ Schema created")
	}

	// Execute based on mode
	switch *modeFlag {
	case "file":
		if *fileFlag == "" {
			fmt.Println("❌ Error: --file flag required for 'file' mode")
			os.Exit(1)
		}
		fmt.Printf("📂 Indexing single file: %s\n\n", *fileFlag)

		metadata := VulnerabilityMetadata{
			Language:         "C",
			VulnType:         "Buffer Overflow",
			CWE:              "CWE-120, CWE-119",
			CWEDescription:   "Classic and memory-based buffer overflows",
			CVE:              "CVE-2022-12345",
			CVSSScore:        8.6,
			CVSSVector:       "AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
			Function:         "copy_user_input",
			FilePath:         *fileFlag,
			Library:          "libutils",
			Severity:         "High",
			ExploitAvailable: true,
			PatchAvailable:   true,
			AffectedVersion:  "v1.0.0 - v1.2.3",
			FixedVersion:     "v1.2.4",
			SourceRepo:       "https://github.com/example/libutils",
			CommitHash:       "a1b2c3d4e5f6g7h8i9j0",
			AuditTool:        "semgrep",
			Auditor:          "internal-audit-team",
		}

		err = IndexCodeFile(client, *fileFlag, SCHEMA, metadata)
		handle(err)

	case "dir":
		dirPath := *dirFlag
		fmt.Printf("📁 Indexing directory: %s\n\n", dirPath)
		err = IndexDirectory(client, dirPath, SCHEMA)
		handle(err)

	case "search":
		if *searchFlag == "" {
			fmt.Println("❌ Error: --search flag required for 'search' mode")
			os.Exit(1)
		}
		fmt.Printf("🔍 Searching for vulnerabilities similar to: %s\n\n", *searchFlag)

		code, err := readCodeFromFile(*searchFlag)
		handle(err)

		vector, err := getVectorFromPython(code)
		handle(err)

		nearVector := client.GraphQL().NearVectorArgBuilder()
		nearVector.WithVector(vector)

		result, err := client.GraphQL().Get().
			WithClassName(SCHEMA).
			WithFields(
				graphql.Field{Name: "code"},
				graphql.Field{Name: "language"},
				graphql.Field{Name: "vulnType"},
				graphql.Field{Name: "cwe"},
				graphql.Field{Name: "severity"},
				graphql.Field{Name: "filePath"},
			).
			WithNearVector(nearVector).
			WithLimit(5).
			Do(context.Background())

		handle(err)
		fmt.Printf("Results:\n%+v\n", result.Data)

	default:
		fmt.Printf("❌ Unknown mode: %s\n", *modeFlag)
		fmt.Println("Available modes: file, dir, search")
		os.Exit(1)
	}

	fmt.Println("\n✅ Done!")
}

func classExists(client *weaviate.Client, className string) (bool, error) {
	schema, err := client.Schema().Getter().Do(context.Background())
	if err != nil {
		return false, err
	}
	for _, cls := range schema.Classes {
		if strings.EqualFold(cls.Class, className) {
			return true, nil
		}
	}
	return false, nil
}

func handle(err error) {
	if err != nil {
		fmt.Printf("\n\n%v\n\n", err)
		os.Exit(1)
	}
}
