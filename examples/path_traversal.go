// Example vulnerable Go code - Path Traversal
// CWE-22: Improper Limitation of a Pathname to a Restricted Directory

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func readUserFile(filename string) (string, error) {
	baseDir := "/var/www/uploads"

	// VULNERABLE: No validation of filename
	// Attacker could use: ../../etc/passwd
	fullPath := filepath.Join(baseDir, filename)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func main() {
	var userFile string
	fmt.Print("Enter filename: ")
	fmt.Scanln(&userFile)

	content, err := readUserFile(userFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(content)
}
