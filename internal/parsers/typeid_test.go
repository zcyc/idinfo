package parsers

import (
	"strings"
	"testing"
)

func TestTypeIDParser_Name(t *testing.T) {
	parser := &TypeIDParser{}
	if parser.Name() != "TypeID" {
		t.Errorf("Expected name 'TypeID', got '%s'", parser.Name())
	}
}

func TestTypeIDParser_Generate(t *testing.T) {
	parser := &TypeIDParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate TypeID: %v", err)
	}
	
	// Check if generated ID can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated TypeID is not valid: %s", generated)
	}
	
	// Check that it contains an underscore
	if !strings.Contains(generated, "_") {
		t.Errorf("Generated TypeID should contain underscore: %s", generated)
	}
	
	// Check that ULID part is 26 characters
	parts := strings.Split(generated, "_")
	if len(parts) != 2 {
		t.Errorf("Generated TypeID should have exactly one underscore: %s", generated)
	}
	
	if len(parts[1]) != 26 {
		t.Errorf("Generated TypeID ULID part should be 26 characters, got %d: %s", len(parts[1]), parts[1])
	}
	
	// Check that prefix is valid
	prefix := parts[0]
	if prefix != "demo" {
		t.Errorf("Expected prefix 'demo', got '%s'", prefix)
	}
}

func TestTypeIDParser_CanParse(t *testing.T) {
	parser := &TypeIDParser{}
	
	// Generate a valid TypeID for testing
	validID, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate TypeID for testing: %v", err)
	}
	
	// Test valid TypeID
	if !parser.CanParse(validID) {
		t.Errorf("Expected to parse valid TypeID: %s", validID)
	}
	
	// Test invalid TypeIDs
	invalidIDs := []string{
		"",                                      // empty
		"01h455vkjcc9jyvvnt6f3p8xz7",          // no prefix (just ULID)
		"_01h455vkjcc9jyvvnt6f3p8xz7",         // prefix starts with underscore
		"user_",                               // no ULID part
		"user_01h455vkjcc9jyvvnt6f3p8x",       // ULID too short
		"user_01h455vkjcc9jyvvnt6f3p8xz7a",    // ULID too long
		"USER_01h455vkjcc9jyvvnt6f3p8xz7",     // uppercase prefix
		"user-01h455vkjcc9jyvvnt6f3p8xz7",     // hyphen instead of underscore
		"user 01h455vkjcc9jyvvnt6f3p8xz7",     // space instead of underscore
		"user__01h455vkjcc9jyvvnt6f3p8xz7",    // double underscore
		"123user_01h455vkjcc9jyvvnt6f3p8xz7",  // prefix starts with number
		"invalid_id",                          // completely invalid
	}
	
	for _, id := range invalidIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid TypeID: %s", id)
		}
	}
}

func TestTypeIDParser_Parse(t *testing.T) {
	parser := &TypeIDParser{}
	
	// Generate a valid TypeID for testing
	validID, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate TypeID for testing: %v", err)
	}
	
	info, err := parser.Parse(validID)
	if err != nil {
		t.Fatalf("Failed to parse valid TypeID: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "TypeID" {
		t.Errorf("Expected IDType 'TypeID', got '%s'", info.IDType)
	}
	
	if info.Standard != validID {
		t.Errorf("Expected Standard '%s', got '%s'", validID, info.Standard)
	}
	
	// Check size (TypeID has same as ULID)
	if info.Size != 128 {
		t.Errorf("Expected Size 128, got %d", info.Size)
	}
	
	// Check entropy (TypeID has 128 bits of entropy)
	if info.Entropy == nil || *info.Entropy != 128 {
		t.Errorf("Expected Entropy 128, got %v", info.Entropy)
	}
	
	// Check extra information
	if info.Extra["type_prefix"] != "demo" {
		t.Errorf("Expected type_prefix 'demo', got '%s'", info.Extra["type_prefix"])
	}
	
	// Check that format is present
	if info.Extra["format"] != "TypeID (type prefix + ULID)" {
		t.Errorf("Expected format info, got '%s'", info.Extra["format"])
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_id")
	if err == nil {
		t.Error("Expected error for invalid TypeID")
	}
	
	// Test whitespace handling
	info, err = parser.Parse("  " + validID + "  ")
	if err != nil {
		t.Errorf("Failed to parse TypeID with whitespace: %v", err)
	}
	if info.Standard != validID {
		t.Errorf("Expected trimmed input, got '%s'", info.Standard)
	}
}