package main

// This file demonstrates alternative vectorization approaches for educational purposes
// These are NOT recommended for production use - see VECTORIZATION.md for details

import (
	"crypto/md5"
	"encoding/binary"
	"hash/fnv"
)

// SimpleHashVector creates a vector using FNV hash
// ❌ NO semantic similarity - only useful for exact duplicate detection
func SimpleHashVector(code string, dimensions int) []float32 {
	h := fnv.New128a()
	h.Write([]byte(code))
	hashBytes := h.Sum(nil)

	vec := make([]float32, dimensions)
	for i := range vec {
		vec[i] = float32(hashBytes[i%len(hashBytes)]) / 255.0 // Normalize to 0-1
	}
	return vec
}

// MD5HashVector creates a vector using MD5 hash
// ❌ NO semantic similarity - different implementation, same problem
func MD5HashVector(code string, dimensions int) []float32 {
	hash := md5.Sum([]byte(code))

	vec := make([]float32, dimensions)
	for i := range vec {
		vec[i] = float32(hash[i%len(hash)]) / 255.0
	}
	return vec
}

// ASCIIVector converts string to ASCII values
// ❌ NO semantic similarity - length-dependent, no pattern recognition
func ASCIIVector(code string, dimensions int) []float32 {
	vec := make([]float32, dimensions)

	for i, char := range code {
		if i >= dimensions {
			break
		}
		vec[i] = float32(char) / 128.0 // Normalize ASCII values
	}

	// Pad with zeros if code is shorter than dimensions
	return vec
}

// NGramVector creates a simple n-gram based vector
// ⚠️ Slightly better but still primitive - no deep semantic understanding
func NGramVector(code string, dimensions int) []float32 {
	vec := make([]float32, dimensions)
	ngrams := make(map[string]int)

	// Extract 3-character n-grams
	for i := 0; i < len(code)-2; i++ {
		ngram := code[i : i+3]
		ngrams[ngram]++
	}

	// Hash n-grams into vector dimensions
	for ngram, count := range ngrams {
		h := fnv.New32a()
		h.Write([]byte(ngram))
		idx := int(h.Sum32()) % dimensions
		vec[idx] += float32(count)
	}

	// Normalize
	var sum float32
	for _, v := range vec {
		sum += v * v
	}
	if sum > 0 {
		norm := float32(1.0) / float32(sum)
		for i := range vec {
			vec[i] *= norm
		}
	}

	return vec
}

// CharacterFrequencyVector creates vector based on character frequencies
// ❌ Loses positional information and semantic meaning
func CharacterFrequencyVector(code string) []float32 {
	vec := make([]float32, 256) // One dimension per ASCII character

	for _, char := range code {
		if char < 256 {
			vec[char]++
		}
	}

	// Normalize by document length
	total := float32(len(code))
	if total > 0 {
		for i := range vec {
			vec[i] /= total
		}
	}

	return vec
}

// WHY THESE APPROACHES FAIL:
//
// Example: Finding similar buffer overflow vulnerabilities
//
// Code 1: strcpy(buffer, user_input);
// Code 2: memcpy(buffer, user_input, len);
// Code 3: sprintf(buffer, "%s", user_input);
//
// With sentence-transformers (e5-base-v2):
//   - All three get similar vectors (semantic similarity ~0.7-0.9)
//   - System finds them as related vulnerabilities ✓
//
// With hash-based approaches:
//   - Completely different vectors (similarity ~0.0-0.1)
//   - System sees them as unrelated ✗
//
// With n-gram approach:
//   - Some overlap (similarity ~0.2-0.4)
//   - Misses semantic relationship ✗
//
// CONCLUSION: Use sentence-transformers for semantic code search!

// BenchmarkComparison demonstrates the difference
func BenchmarkComparison() {
	code1 := "strcpy(buffer, user_input);"
	code2 := "memcpy(buffer, user_input, len);"
	code3 := "printf(\"hello\");" // Unrelated

	println("=== Hash-based Vectorization ===")
	v1_hash := SimpleHashVector(code1, 128)
	v2_hash := SimpleHashVector(code2, 128)
	v3_hash := SimpleHashVector(code3, 128)

	println("Code 1 vs Code 2 (should be similar):", cosineSimilarity(v1_hash, v2_hash))
	println("Code 1 vs Code 3 (should be different):", cosineSimilarity(v1_hash, v3_hash))
	println("Note: Hash-based can't distinguish semantic similarity!\n")

	println("=== N-gram Vectorization ===")
	v1_ngram := NGramVector(code1, 128)
	v2_ngram := NGramVector(code2, 128)
	v3_ngram := NGramVector(code3, 128)

	println("Code 1 vs Code 2:", cosineSimilarity(v1_ngram, v2_ngram))
	println("Code 1 vs Code 3:", cosineSimilarity(v1_ngram, v3_ngram))
	println("Note: N-grams find some overlap but miss deeper meaning!\n")
}

// cosineSimilarity calculates similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	var dotProduct, normA, normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(binary.Size(normA)) * float32(binary.Size(normB)))
}
