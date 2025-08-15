package parsers

import (
	"testing"

	"github.com/nats-io/nuid"
)

func TestNUIDParser_Name(t *testing.T) {
	parser := &NUIDParser{}
	if parser.Name() != "NUID" {
		t.Errorf("Expected name 'NUID', got '%s'", parser.Name())
	}
}

func TestNUIDParser_CanParse(t *testing.T) {
	parser := &NUIDParser{}
	
	// Test valid NUIDs (generate some using the official library)
	validNUIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		validNUIDs[i] = nuid.Next()
	}
	
	for _, id := range validNUIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid NUID: %s", id)
		}
	}
	
	// Test invalid NUIDs
	invalidNUIDs := []string{
		"",                           // empty
		"abc",                        // too short
		"1234567890123456789012345",  // too long (25 chars)
		"123456789012345678901",      // too short (21 chars)
		"12345678901234567890123",    // too long (23 chars)
		"!@#$%^&*()1234567890ab",     // invalid characters
		"abcdefghijklmnopqr-+==",     // contains invalid chars
		"0123456789012345678901",     // starts with 0 (heuristic)
		"abcdefghijklmnopqr st u",     // contains space
	}
	
	for _, id := range invalidNUIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid NUID: %s", id)
		}
	}
}

func TestNUIDParser_Parse(t *testing.T) {
	parser := &NUIDParser{}
	
	// Generate a valid NUID for testing
	validNUID := nuid.Next()
	
	info, err := parser.Parse(validNUID)
	if err != nil {
		t.Fatalf("Failed to parse valid NUID: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "NUID (NATS Unique Identifier)" {
		t.Errorf("Expected IDType 'NUID (NATS Unique Identifier)', got '%s'", info.IDType)
	}
	
	if info.Standard != validNUID {
		t.Errorf("Expected Standard '%s', got '%s'", validNUID, info.Standard)
	}
	
	// Check size (should be around 132 bits)
	if info.Size != 132 {
		t.Errorf("Expected Size 132, got %d", info.Size)
	}
	
	// Check entropy (should be around 132 bits)
	if info.Entropy == nil || *info.Entropy != 132 {
		t.Errorf("Expected Entropy 132, got %v", info.Entropy)
	}
	
	// Check extra information
	if info.Extra["alphabet"] != "Base62 (0-9A-Za-z)" {
		t.Errorf("Expected alphabet info, got '%s'", info.Extra["alphabet"])
	}
	
	if info.Extra["length"] != "22 characters" {
		t.Errorf("Expected length info, got '%s'", info.Extra["length"])
	}
	
	if info.Extra["format"] != "NATS Unique Identifier" {
		t.Errorf("Expected format info, got '%s'", info.Extra["format"])
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_id")
	if err == nil {
		t.Error("Expected error for invalid NUID")
	}
	
	// Test whitespace handling
	info, err = parser.Parse("  " + validNUID + "  ")
	if err != nil {
		t.Errorf("Failed to parse NUID with whitespace: %v", err)
	}
	if info.Standard != validNUID {
		t.Errorf("Expected trimmed input, got '%s'", info.Standard)
	}
}

func TestNUIDParser_Generate(t *testing.T) {
	parser := &NUIDParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate NUID: %v", err)
	}
	
	// Check if generated NUID can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated NUID is not valid: %s", generated)
	}
	
	// Check length
	if len(generated) != 22 {
		t.Errorf("Generated NUID should be 22 characters, got %d: %s", len(generated), generated)
	}
	
	// Check character set
	for _, char := range generated {
		if !((char >= '0' && char <= '9') || 
		     (char >= 'A' && char <= 'Z') || 
		     (char >= 'a' && char <= 'z')) {
			t.Errorf("Generated NUID contains invalid character '%c': %s", char, generated)
		}
	}
	
	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second NUID: %v", err)
	}
	
	if generated == generated2 {
		t.Error("Two consecutive NUID generations should produce different results")
	}
	
	// Test multiple generations for uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate NUID #%d: %v", i, err)
		}
		if seen[id] {
			t.Errorf("Duplicate NUID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestNUIDParser_Performance(t *testing.T) {
	parser := &NUIDParser{}
	
	// Test that we can generate many NUIDs quickly
	const count = 10000
	generated := make([]string, count)
	
	for i := 0; i < count; i++ {
		id, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate NUID #%d: %v", i, err)
		}
		generated[i] = id
	}
	
	// Check for uniqueness
	seen := make(map[string]bool)
	for _, id := range generated {
		if seen[id] {
			t.Errorf("Duplicate NUID found: %s", id)
		}
		seen[id] = true
	}
}

func TestNUIDParser_RealWorldExamples(t *testing.T) {
	parser := &NUIDParser{}
	
	// Test with some real NUID examples (these are actual NUIDs)
	realNUIDs := []string{
		"B5CsmFZRhDg5XQA2fqrCR8",
		"B5CsmFZRhDg5XQA2fqrCR9",
		"B5CsmFZRhDg5XQA2fqrCRA",
	}
	
	for _, nuidStr := range realNUIDs {
		if !parser.CanParse(nuidStr) {
			t.Errorf("Should parse real-world NUID: %s", nuidStr)
			continue
		}
		
		info, err := parser.Parse(nuidStr)
		if err != nil {
			t.Errorf("Failed to parse real-world NUID %s: %v", nuidStr, err)
			continue
		}
		
		if info.IDType != "NUID (NATS Unique Identifier)" {
			t.Errorf("Expected NUID type for %s, got %s", nuidStr, info.IDType)
		}
	}
}

func TestNUIDParser_EdgeCases(t *testing.T) {
	parser := &NUIDParser{}
	
	// Test edge cases
	edgeCases := []struct {
		input    string
		shouldParse bool
		description string
	}{
		{"0234567890123456789012", false, "starts with 0 (heuristic)"},
		{"1234567890123456789012", true, "starts with 1"},
		{"A234567890123456789012", true, "starts with A"},
		{"z234567890123456789012", true, "starts with z"},
		{"9234567890123456789012", true, "starts with 9"},
		{"aBcDeFgHiJkLmNoPqRsTuV", true, "mixed case"},
		{"0000000000000000000000", false, "all zeros"},
		{"AAAAAAAAAAAAAAAAAAAAAA", true, "all A's"},
		{"aaaaaaaaaaaaaaaaaaaaaa", true, "all a's"},
		{"9999999999999999999999", true, "all 9's"},
	}
	
	for _, tc := range edgeCases {
		result := parser.CanParse(tc.input)
		if result != tc.shouldParse {
			t.Errorf("CanParse(\"%s\") = %v, expected %v (%s)", 
				tc.input, result, tc.shouldParse, tc.description)
		}
	}
}