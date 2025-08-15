package parsers

import (
	"strings"
	"testing"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

func TestNanoIDParser_Name(t *testing.T) {
	parser := &NanoIDParser{}
	if parser.Name() != "NanoID" {
		t.Errorf("Expected name 'NanoID', got '%s'", parser.Name())
	}
}

func TestNanoIDParser_CanParse(t *testing.T) {
	parser := &NanoIDParser{}
	
	// Test valid NanoIDs (generate some using the official library)
	validNanoIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		generated, err := gonanoid.New()
		if err != nil {
			t.Fatalf("Failed to generate NanoID for test: %v", err)
		}
		validNanoIDs[i] = generated
	}
	
	for _, id := range validNanoIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid NanoID: %s", id)
		}
	}
	
	// Test some valid NanoID patterns with different lengths
	additionalValid := []string{
		"V1StGXR8_Z5jdHi6B-myT",       // Standard 21-char NanoID
		"4f90d13a42",                   // Shorter NanoID
		"ABCDEF123456789",              // Mixed case
		"_-abc123XYZ",                  // All valid characters
		"123456",                       // Minimum length
	}
	
	for _, id := range additionalValid {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid NanoID pattern: %s", id)
		}
	}
	
	// Test invalid NanoIDs
	invalidNanoIDs := []string{
		"",                                      // empty
		"12345",                                 // too short (< 6 chars)
		strings.Repeat("a", 256),                // too long (> 255 chars)
		"abc@123",                               // contains invalid character '@'
		"abc 123",                               // contains space
		"abc+123",                               // contains '+'
		"abc=123",                               // contains '='
		"abc/123",                               // contains '/'
		"abc\\123",                              // contains backslash
		"abc!123",                               // contains '!'
		"abc#123",                               // contains '#'
	}
	
	for _, id := range invalidNanoIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid NanoID: %s", id)
		}
	}
}

func TestNanoIDParser_Parse(t *testing.T) {
	parser := &NanoIDParser{}
	
	// Generate a valid NanoID for testing
	validNanoID, err := gonanoid.New()
	if err != nil {
		t.Fatalf("Failed to generate NanoID for test: %v", err)
	}
	
	info, err := parser.Parse(validNanoID)
	if err != nil {
		t.Fatalf("Failed to parse valid NanoID: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "Nano ID" {
		t.Errorf("Expected IDType 'Nano ID', got '%s'", info.IDType)
	}
	
	if info.Standard != validNanoID {
		t.Errorf("Expected Standard '%s', got '%s'", validNanoID, info.Standard)
	}
	
	// Check size (should be length * 8 bits approximately)
	expectedSize := len(validNanoID) * 8
	if info.Size != expectedSize {
		t.Errorf("Expected Size %d, got %d", expectedSize, info.Size)
	}
	
	// Check entropy (should be reasonable for the alphabet size)
	if info.Entropy == nil {
		t.Error("Expected Entropy to be set")
	} else {
		// With 64 character alphabet and typical 21 char length, entropy should be around 126 bits
		if *info.Entropy < 100 || *info.Entropy > 200 {
			t.Errorf("Expected reasonable entropy (100-200), got %d", *info.Entropy)
		}
	}
	
	// Check extra information
	if info.Extra["alphabet"] != nanoIDAlphabet {
		t.Errorf("Expected alphabet '%s', got '%s'", nanoIDAlphabet, info.Extra["alphabet"])
	}
	
	if info.Extra["alphabet_size"] != "64" {
		t.Errorf("Expected alphabet_size '64', got '%s'", info.Extra["alphabet_size"])
	}
	
	actualLengthStr := info.Extra["length"]
	if actualLengthStr == "" {
		t.Error("Expected length to be set")
	} else {
		// Just check it's a reasonable number representation
		if len(actualLengthStr) == 0 {
			t.Error("Expected length to be a non-empty string")
		}
	}
	
	if info.Extra["url_safe"] != "true" {
		t.Errorf("Expected url_safe 'true', got '%s'", info.Extra["url_safe"])
	}
	
	if info.Extra["collision_resistant"] != "true" {
		t.Errorf("Expected collision_resistant 'true', got '%s'", info.Extra["collision_resistant"])
	}
	
	// Check collision probability for standard 21-char NanoID
	if len(validNanoID) == 21 {
		if info.Extra["collision_probability"] != "~1% in 4 years (1 ID/hour)" {
			t.Errorf("Expected collision probability info for 21-char NanoID, got '%s'", info.Extra["collision_probability"])
		}
	}
	
	// Check hex representation exists
	if len(info.Hex) == 0 {
		t.Error("Expected Hex representation to be set")
	}
	
	// Check binary representation exists
	if len(info.Binary) == 0 {
		t.Error("Expected Binary representation to be set")
	}
	
	// Test invalid input
	_, err = parser.Parse("abc@123")
	if err == nil {
		t.Error("Expected error for invalid NanoID")
	}
}

func TestNanoIDParser_Generate(t *testing.T) {
	parser := &NanoIDParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate NanoID: %v", err)
	}
	
	// Check if generated NanoID can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated NanoID is not valid: %s", generated)
	}
	
	// Check length (default NanoID is 21 characters)
	if len(generated) != 21 {
		t.Errorf("Generated NanoID should be 21 characters, got %d: %s", len(generated), generated)
	}
	
	// Check character set (should only contain valid NanoID characters)
	for _, char := range generated {
		if !containsChar(nanoIDAlphabet, char) {
			t.Errorf("Generated NanoID contains invalid character '%c': %s", char, generated)
		}
	}
	
	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second NanoID: %v", err)
	}
	
	if generated == generated2 {
		t.Error("Two consecutive NanoID generations should produce different results")
	}
	
	// Test multiple generations for uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		gen, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate NanoID #%d: %v", i, err)
		}
		if seen[gen] {
			t.Errorf("Generated duplicate NanoID: %s", gen)
		}
		seen[gen] = true
	}
}

func TestNanoIDParser_Alphabet(t *testing.T) {
	// Test that alphabet is correct
	expectedAlphabet := "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if nanoIDAlphabet != expectedAlphabet {
		t.Errorf("Expected alphabet '%s', got '%s'", expectedAlphabet, nanoIDAlphabet)
	}
	
	// Test that alphabet has the right size
	if len(nanoIDAlphabet) != 64 {
		t.Errorf("Expected alphabet size 64, got %d", len(nanoIDAlphabet))
	}
	
	// Test that all characters in the alphabet are unique
	seen := make(map[rune]bool)
	for _, char := range nanoIDAlphabet {
		if seen[char] {
			t.Errorf("Duplicate character in alphabet: %c", char)
		}
		seen[char] = true
	}
}

func TestNanoIDParser_HelperFunctions(t *testing.T) {
	// Test containsChar function
	if !containsChar("abc", 'b') {
		t.Error("containsChar should find 'b' in 'abc'")
	}
	
	if containsChar("abc", 'd') {
		t.Error("containsChar should not find 'd' in 'abc'")
	}
	
	// Test indexOfChar function
	if indexOfChar("abc", 'b') != 1 {
		t.Errorf("indexOfChar should return 1 for 'b' in 'abc', got %d", indexOfChar("abc", 'b'))
	}
	
	if indexOfChar("abc", 'd') != -1 {
		t.Errorf("indexOfChar should return -1 for 'd' in 'abc', got %d", indexOfChar("abc", 'd'))
	}
}

func TestNanoIDParser_Performance(t *testing.T) {
	parser := &NanoIDParser{}
	
	// Test parsing performance
	testNanoID, err := gonanoid.New()
	if err != nil {
		t.Fatalf("Failed to generate test NanoID: %v", err)
	}
	
	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testNanoID)
		if err != nil {
			t.Fatalf("Parse failed during performance test: %v", err)
		}
	}
	
	// Test generation performance
	for i := 0; i < 1000; i++ {
		_, err := parser.Generate()
		if err != nil {
			t.Fatalf("Generate failed during performance test: %v", err)
		}
	}
}

func TestNanoIDParser_OfficialLibraryIntegration(t *testing.T) {
	parser := &NanoIDParser{}
	
	// Test that our parser works with the official library
	officialNanoID, err := gonanoid.New()
	if err != nil {
		t.Fatalf("Failed to generate NanoID with official library: %v", err)
	}
	
	// Parse with our parser
	info, err := parser.Parse(officialNanoID)
	if err != nil {
		t.Fatalf("Failed to parse NanoID from official library: %v", err)
	}
	
	// Basic validation
	if info.IDType != "Nano ID" {
		t.Errorf("Expected correct IDType, got: %s", info.IDType)
	}
	
	if info.Standard != officialNanoID {
		t.Errorf("Expected Standard to match input: %s vs %s", info.Standard, officialNanoID)
	}
	
	// Test different lengths if the library supports them
	customNanoID, err := gonanoid.Generate("0123456789", 10)
	if err != nil {
		t.Fatalf("Failed to generate custom NanoID: %v", err)
	}
	
	// This might not be parseable by our parser since it uses a different alphabet
	// But we can check that our parser handles it gracefully
	canParse := parser.CanParse(customNanoID)
	t.Logf("Custom NanoID with different alphabet: %s, can parse: %v", customNanoID, canParse)
}

func TestNanoIDParser_EdgeCases(t *testing.T) {
	parser := &NanoIDParser{}
	
	// Test edge cases for length
	edgeCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"123456", true, "minimum length (6 chars)"},
		{"12345", false, "too short (5 chars)"},
		{strings.Repeat("a", 255), true, "maximum length (255 chars)"},
		{strings.Repeat("a", 256), false, "too long (256 chars)"},
		{"V1StGXR8_Z5jdHi6B-myT", true, "standard 21-char NanoID"},
		{"_-abc123XYZ", true, "mixed valid characters"},
		{"___________", true, "all underscores"},
		{"-----------", true, "all hyphens"},
		{"00000000000", true, "all zeros"},
		{"AAAAAAAAAAA", true, "all uppercase"},
		{"aaaaaaaaaaa", true, "all lowercase"},
	}
	
	for _, tc := range edgeCases {
		result := parser.CanParse(tc.input)
		if result != tc.shouldParse {
			t.Errorf("CanParse(\"%s\") = %v, expected %v (%s)", 
				tc.input, result, tc.shouldParse, tc.description)
		}
		
		// If it should parse, try parsing it
		if tc.shouldParse && result {
			_, err := parser.Parse(tc.input)
			if err != nil {
				t.Errorf("Parse(\"%s\") failed but CanParse returned true: %v (%s)", 
					tc.input, err, tc.description)
			}
		}
	}
}