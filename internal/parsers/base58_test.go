package parsers

import (
	"testing"
)

func TestBase58Parser(t *testing.T) {
	parser := &Base58Parser{}
	
	// Test valid Base58 IDs
	validIDs := []string{
		"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy", // Bitcoin address
		"7WDKdXBqai6cGhNM9HXpQEF2vWgf5Bb2xn",
		"4HUtbDSKF7CKgkd2RSsWsq",
		"StV1DL6CwTryKyV",
		"5HueCGU8rMjxEXxiPuD5BDku4MkFqeZyd4dZ1jvhTVqvbTLvyTJ", // Private key format
	}
	
	for _, id := range validIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid Base58 ID: %s", id)
		}
		
		info, err := parser.Parse(id)
		if err != nil {
			t.Errorf("Failed to parse Base58 ID %s: %v", id, err)
		}
		
		if info.IDType != "Base58" {
			t.Errorf("Expected IDType to be Base58, got %s", info.IDType)
		}
	}
	
	// Test invalid IDs
	invalidIDs := []string{
		"0OIl",     // contains confusing characters
		"abc123",   // too short
		"",         // empty
		"12345678", // all numbers, too short
	}
	
	for _, id := range invalidIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid Base58 ID: %s", id)
		}
	}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Errorf("Failed to generate Base58 ID: %v", err)
	}
	
	if !parser.CanParse(generated) {
		t.Errorf("Generated Base58 ID is not valid: %s", generated)
	}
}

// Benchmark tests for Base58 parser
func BenchmarkBase58Parser(b *testing.B) {
	parser := &Base58Parser{}
	testID := "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(testID)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}