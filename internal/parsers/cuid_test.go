package parsers

import (
	"testing"

	"github.com/nrednav/cuid2"
)

func TestCUIDParser_Name(t *testing.T) {
	parser := &CUIDParser{}
	if parser.Name() != "CUID" {
		t.Errorf("Expected name 'CUID', got '%s'", parser.Name())
	}
}

func TestCUIDParser_CanParse(t *testing.T) {
	parser := &CUIDParser{}
	
	// Test valid CUIDs (generate some using the official library)
	validCUIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		validCUIDs[i] = cuid2.Generate()
	}
	
	for _, id := range validCUIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid CUID: %s", id)
		}
	}
	
	// Test some known valid CUID2 patterns
	additionalValid := []string{
		"ckj0123456789abcdef",  // typical CUID2 pattern
		"abcd1234",             // short valid CUID2
		"x123456789abcdefghij", // medium length
	}
	
	for _, id := range additionalValid {
		// Only test if the official library considers it valid
		if cuid2.IsCuid(id) && !parser.CanParse(id) {
			t.Errorf("Expected to parse valid CUID: %s", id)
		}
	}
	
	// Test invalid CUIDs
	invalidCUIDs := []string{
		"",                       // empty
		"abc",                    // too short (< 4 chars)
		"123456789012345678901234567890123456789", // too long (> 32 chars)
		"ABC123",                 // contains uppercase
		"abc-123",                // contains hyphen
		"abc_123",                // contains underscore
		"abc 123",                // contains space
		"abc!123",                // contains special character
		"123456789012345678901234567890123", // exactly 33 chars (too long)
	}
	
	for _, id := range invalidCUIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid CUID: %s", id)
		}
	}
}

func TestCUIDParser_Parse(t *testing.T) {
	parser := &CUIDParser{}
	
	// Generate a valid CUID for testing
	validCUID := cuid2.Generate()
	
	info, err := parser.Parse(validCUID)
	if err != nil {
		t.Fatalf("Failed to parse valid CUID: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "CUID v2 (Collision-resistant Unique Identifier)" {
		t.Errorf("Expected IDType 'CUID v2 (Collision-resistant Unique Identifier)', got '%s'", info.IDType)
	}
	
	if info.Standard != validCUID {
		t.Errorf("Expected Standard '%s', got '%s'", validCUID, info.Standard)
	}
	
	// Check size (should be length * 6 bits approximately)
	expectedSize := len(validCUID) * 6
	if info.Size != expectedSize {
		t.Errorf("Expected Size %d, got %d", expectedSize, info.Size)
	}
	
	// Check entropy (should be around length * 5.2 bits)
	expectedEntropy := int(float64(len(validCUID)) * 5.2)
	if info.Entropy == nil || *info.Entropy != expectedEntropy {
		t.Errorf("Expected Entropy %d, got %v", expectedEntropy, info.Entropy)
	}
	
	// Check extra information
	if info.Extra["version"] != "2" {
		t.Errorf("Expected version '2', got '%s'", info.Extra["version"])
	}
	
	if info.Extra["encoding"] != "Base36 (lowercase)" {
		t.Errorf("Expected encoding 'Base36 (lowercase)', got '%s'", info.Extra["encoding"])
	}
	
	if info.Extra["collision_resistant"] != "Yes" {
		t.Errorf("Expected collision_resistant 'Yes', got '%s'", info.Extra["collision_resistant"])
	}
	
	if info.Extra["cryptographically_secure"] != "Yes" {
		t.Errorf("Expected cryptographically_secure 'Yes', got '%s'", info.Extra["cryptographically_secure"])
	}
	
	if info.Extra["url_safe"] != "Yes" {
		t.Errorf("Expected url_safe 'Yes', got '%s'", info.Extra["url_safe"])
	}
	
	// Check that length is set and is reasonable
	if len(info.Extra["length"]) == 0 {
		t.Errorf("Expected length to be set, got empty")
	}
	
	// Check hex representation exists
	if len(info.Hex) == 0 {
		t.Error("Expected Hex representation to be set")
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_id")
	if err == nil {
		t.Error("Expected error for invalid CUID")
	}
}

func TestCUIDParser_Generate(t *testing.T) {
	parser := &CUIDParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate CUID: %v", err)
	}
	
	// Check if generated CUID can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated CUID is not valid: %s", generated)
	}
	
	// Check length is within valid range
	if len(generated) < 4 || len(generated) > 32 {
		t.Errorf("Generated CUID length should be 4-32 characters, got %d: %s", len(generated), generated)
	}
	
	// Check character set (lowercase letters and digits only)
	for _, char := range generated {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'z')) {
			t.Errorf("Generated CUID contains invalid character '%c': %s", char, generated)
		}
	}
	
	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second CUID: %v", err)
	}
	
	if generated == generated2 {
		t.Error("Two consecutive CUID generations should produce different results")
	}
	
	// Test multiple generations for uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		gen, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate CUID #%d: %v", i, err)
		}
		if seen[gen] {
			t.Errorf("Generated duplicate CUID: %s", gen)
		}
		seen[gen] = true
	}
}

func TestCUIDParser_Performance(t *testing.T) {
	parser := &CUIDParser{}
	
	// Test parsing performance
	testCuid := cuid2.Generate()
	
	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testCuid)
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

func TestCUIDParser_OfficialLibraryIntegration(t *testing.T) {
	parser := &CUIDParser{}
	
	// Test that our parser works with the official library
	officialCuid := cuid2.Generate()
	
	// Verify it's considered valid by both libraries
	if !cuid2.IsCuid(officialCuid) {
		t.Fatalf("Official library generated invalid CUID: %s", officialCuid)
	}
	
	if !parser.CanParse(officialCuid) {
		t.Errorf("Our parser should accept CUID from official library: %s", officialCuid)
	}
	
	// Parse with our parser
	info, err := parser.Parse(officialCuid)
	if err != nil {
		t.Fatalf("Failed to parse CUID from official library: %v", err)
	}
	
	// Basic validation
	if info.IDType != "CUID v2 (Collision-resistant Unique Identifier)" {
		t.Errorf("Expected correct IDType, got: %s", info.IDType)
	}
	
	if info.Standard != officialCuid {
		t.Errorf("Expected Standard to match input: %s vs %s", info.Standard, officialCuid)
	}
}

func TestCUIDParser_EdgeCases(t *testing.T) {
	parser := &CUIDParser{}
	
	// Test edge cases for length
	edgeCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"abcd", true, "minimum length (4 chars)"},  // Will test with cuid2.IsCuid
		{"12345678901234567890123456789012", true, "maximum length (32 chars)"}, // Will test with cuid2.IsCuid
		{"abc", false, "too short (3 chars)"},
		{"123456789012345678901234567890123", false, "too long (33 chars)"},
		{"a1b2", true, "mixed letters and numbers (4 chars)"}, // Will test with cuid2.IsCuid
		{"0000", true, "all numbers (4 chars)"}, // Will test with cuid2.IsCuid
		{"aaaa", true, "all letters (4 chars)"}, // Will test with cuid2.IsCuid
	}
	
	for _, tc := range edgeCases {
		// For short inputs that we expect to parse, first check if the official library accepts them
		expectedResult := tc.shouldParse
		if tc.shouldParse && len(tc.input) >= 4 && len(tc.input) <= 32 {
			// Use the official library as the source of truth for valid CUIDs
			expectedResult = cuid2.IsCuid(tc.input)
		}
		
		result := parser.CanParse(tc.input)
		if result != expectedResult {
			t.Errorf("CanParse(\"%s\") = %v, expected %v (%s)", 
				tc.input, result, expectedResult, tc.description)
		}
		
		// If it should parse, try parsing it
		if expectedResult && result {
			_, err := parser.Parse(tc.input)
			if err != nil {
				t.Errorf("Parse(\"%s\") failed but CanParse returned true: %v (%s)", 
					tc.input, err, tc.description)
			}
		}
	}
}