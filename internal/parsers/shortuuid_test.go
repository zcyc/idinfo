package parsers

import (
	"strings"
	"testing"
)

func TestShortUUIDParser_Name(t *testing.T) {
	parser := &ShortUUIDParser{}
	if parser.Name() != "ShortUUID" {
		t.Errorf("Expected name 'ShortUUID', got '%s'", parser.Name())
	}
}

func TestShortUUIDParser_CanParse(t *testing.T) {
	parser := &ShortUUIDParser{}
	
	// Test valid ShortUUIDs (22 characters, base57 alphabet)
	validIDs := []string{
		"23456789ABCDEFGHJKLMNt", // 22 chars, valid alphabet
		"PQRSTUVWXYZabcdefghijk", // 22 chars, valid alphabet
		"mnpqrstuvwxyz23456789a", // 22 chars, valid alphabet
		"ABCDEFGHJKLMNPQRSTUVWx", // 22 chars, valid alphabet
	}
	
	for _, id := range validIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid ShortUUID: %s", id)
		}
	}
	
	// Test invalid IDs
	invalidIDs := []string{
		"",                        // empty
		"short",                   // too short (5 chars)
		"this_is_way_too_long_for_shortuuid_format", // too long
		"23456789ABCDEFGHJKLMNt0", // 23 chars (too long)
		"23456789ABCDEFGHJKLMN",   // 21 chars (too short)
		"23456789ABCDEFGHJKLMNOt", // contains confusing character 'O'
		"23456789ABCDEFGHJKLMNIt", // contains confusing character 'I'
		"23456789ABCDEFGHJKLMNlt", // contains confusing character 'l'
		"23456789ABCDEFGHJKLMn1t", // contains confusing character '1'
		"23456789ABCDEFGHJKLMn0t", // contains confusing character '0'
		"@#$%^&*()ABCDEFGHJKLMN",  // invalid characters
	}
	
	for _, id := range invalidIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid ShortUUID: %s", id)
		}
	}
}

func TestShortUUIDParser_Parse(t *testing.T) {
	parser := &ShortUUIDParser{}
	
	validID := "23456789ABCDEFGHJKLMNt"
	
	info, err := parser.Parse(validID)
	if err != nil {
		t.Fatalf("Failed to parse valid ShortUUID: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "ShortUUID" {
		t.Errorf("Expected IDType 'ShortUUID', got '%s'", info.IDType)
	}
	
	if info.Standard != validID {
		t.Errorf("Expected Standard '%s', got '%s'", validID, info.Standard)
	}
	
	if info.Size != 128 {
		t.Errorf("Expected Size 128, got %d", info.Size)
	}
	
	if info.Entropy == nil || *info.Entropy != 122 {
		t.Errorf("Expected Entropy 122, got %v", info.Entropy)
	}
	
	// Check extra information
	if info.Extra["alphabet"] != "Base57 (no ambiguous chars)" {
		t.Errorf("Expected alphabet info, got '%s'", info.Extra["alphabet"])
	}
	
	if info.Extra["length"] != "22 characters" {
		t.Errorf("Expected length info, got '%s'", info.Extra["length"])
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid")
	if err == nil {
		t.Error("Expected error for invalid ShortUUID")
	}
	
	// Test whitespace handling
	info, err = parser.Parse("  " + validID + "  ")
	if err != nil {
		t.Errorf("Failed to parse ShortUUID with whitespace: %v", err)
	}
	if info.Standard != validID {
		t.Errorf("Expected trimmed input, got '%s'", info.Standard)
	}
}

func TestShortUUIDParser_Generate(t *testing.T) {
	parser := &ShortUUIDParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate ShortUUID: %v", err)
	}
	
	// Check generated ID length
	if len(generated) != 22 {
		t.Errorf("Expected generated ShortUUID to be 22 characters, got %d", len(generated))
	}
	
	// Check if generated ID can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated ShortUUID is not valid: %s", generated)
	}
	
	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second ShortUUID: %v", err)
	}
	
	if generated == generated2 {
		t.Error("Generated ShortUUIDs should be different (very unlikely to be same)")
	}
}

func TestShortUUIDParser_Coverage(t *testing.T) {
	parser := &ShortUUIDParser{}
	
	// Test edge cases for full coverage
	testCases := []struct {
		input    string
		shouldParse bool
		description string
	}{
		{"", false, "empty string"},
		{"a", false, "single character"},
		{"abcde", false, "too short"},
		{strings.Repeat("a", 21), false, "21 characters"},
		{strings.Repeat("a", 22), true, "exactly 22 characters"},
		{strings.Repeat("a", 23), false, "23 characters"},
		{"23456789ABCDEFGHJKLMNa", true, "valid base57 chars"},
		{"23456789ABCDEFGHJKLMNaO", false, "contains O"},
		{"23456789ABCDEFGHJKLMNaI", false, "contains I"},
		{"23456789ABCDEFGHJKLMNal", false, "contains l"},
		{"23456789ABCDEFGHJKLMNa1", false, "contains 1"},
		{"23456789ABCDEFGHJKLMNa0", false, "contains 0"},
	}
	
	for _, tc := range testCases {
		result := parser.CanParse(tc.input)
		if result != tc.shouldParse {
			t.Errorf("CanParse(%s) = %v, expected %v (%s)", 
				tc.input, result, tc.shouldParse, tc.description)
		}
	}
}