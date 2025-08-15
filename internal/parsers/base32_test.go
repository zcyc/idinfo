package parsers

import (
	"testing"
)

func TestBase32Parser(t *testing.T) {
	parser := &Base32Parser{}

	// Test valid Base32 IDs
	validIDs := []string{
		"JBSWY3DPEHPK3PXP",
		"GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
		"MFRGG43FMZRW6ZDJNZTSA43FMFQQ====",
		"NBSWY3DP",
	}

	for _, id := range validIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid Base32 ID: %s", id)
		}

		info, err := parser.Parse(id)
		if err != nil {
			t.Errorf("Failed to parse Base32 ID %s: %v", id, err)
		}

		if info.IDType != "Base32" {
			t.Errorf("Expected IDType to be Base32, got %s", info.IDType)
		}
	}

	// Test invalid IDs
	invalidIDs := []string{
		"189",     // contains invalid characters for Base32
		"abc",     // too short
		"",        // empty
		"ABCDEFG", // invalid Base32 (contains characters not in alphabet)
	}

	for _, id := range invalidIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid Base32 ID: %s", id)
		}
	}

	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Errorf("Failed to generate Base32 ID: %v", err)
	}

	if !parser.CanParse(generated) {
		t.Errorf("Generated Base32 ID is not valid: %s", generated)
	}
}
