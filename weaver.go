package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

// Request body
type EmbedRequest struct {
	Text string `json:"text"`
}

// Response body
type EmbedResponse struct {
	Vector []float64 `json:"vector"` // Use float64 because JSON unmarshals floats as float64
}

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

func main() {
	fmt.Println("starting...")

	config := weaviate.Config{
		Scheme: "http",
		Host:   "localhost:8080",
	}

	client, err := weaviate.NewClient(config)
	handle(err)

	metaGetter := client.Misc().MetaGetter()
	meta, err := metaGetter.Do(context.Background())
	handle(err)

	fmt.Printf("Weaviate meta information:\n")
	fmt.Printf("Hostname: %s\nVersion: %s\n", meta.Hostname, meta.Version)

	const SCHEMA = "CodeSnippet10"
	ok, err := classExists(client, SCHEMA)
	handle(err)
	if ok {
		fmt.Println(SCHEMA + " exists")
	}
	if !ok {
		// Create schema (only once)
		schema := &models.Class{
			Class:      SCHEMA,
			Vectorizer: "none", // Don't use text2vec-openai
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
	}

	code := "strcpy(buffer, user_input);"
	floaters, err := getVectorFromPython(code)
	handle(err)

	_, err = client.Data().Creator().
		WithClassName(SCHEMA).
		WithProperties(map[string]interface{}{
			"code":     `strcpy(buffer, user_input);`,
			"language": "C",
			"vulnType": "Buffer Overflow",

			"cwe":            "CWE-120, CWE-119",
			"cweDescription": "Classic and memory-based buffer overflows",
			"cve":            "CVE-2022-12345",
			"cvssScore":      8.6,
			"cvssVector":     "AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",

			"function": "copy_user_input",
			"filePath": "src/main.c",
			"library":  "libutils",

			"severity":         "High",
			"exploitAvailable": true,
			"patchAvailable":   true,
			"affectedVersion":  "v1.0.0 - v1.2.3",
			"fixedVersion":     "v1.2.4",

			"sourceRepo": "https://github.com/example/libutils",
			"commitHash": "a1b2c3d4e5f6g7h8i9j0",
			"auditTool":  "semgrep",
			"auditor":    "internal-audit-team",
		}).
		WithVector(floaters).
		Do(context.Background())
	handle(err)

	nearVector := client.GraphQL().NearVectorArgBuilder()
	nearVector.WithVector(floaters)
	result, err := client.GraphQL().Get().
		WithClassName(SCHEMA).
		WithFields(
			graphql.Field{Name: "code"},
			graphql.Field{Name: "language"},
			graphql.Field{Name: "vulnType"},
			graphql.Field{Name: "cwe"},
		).
		WithNearVector(nearVector).
		Do(context.Background())

	handle(err)

	fmt.Printf("Search result:\n%+v\n", result.Data)
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
