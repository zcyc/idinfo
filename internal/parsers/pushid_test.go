package parsers

import (
	"testing"
)

func TestPushIDParser(t *testing.T) {
	parser := &PushIDParser{}
	
	// Test valid PushIDs (exactly 20 characters)
	validIDs := []string{
		"-M_Kj4tZ8vZjdHi6B-my",
		"0123456789ABCDEFGHIJ",
		"-NeAmG2h7zNj8vZjdHi6",
		"_abcdefghijklmnopqr1",
	}
	
	for _, id := range validIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid PushID: %s", id)
		}
		
		info, err := parser.Parse(id)
		if err != nil {
			t.Errorf("Failed to parse PushID %s: %v", id, err)
		}
		
		if info.IDType != "Firebase PushID" {
			t.Errorf("Expected IDType to be Firebase PushID, got %s", info.IDType)
		}
	}
	
	// Test invalid IDs
	invalidIDs := []string{
		"123456789012345678901", // 21 characters
		"1234567890123456789",   // 19 characters
		"",                      // empty
		"@#$%^&*()1234567890",   // invalid characters
	}
	
	for _, id := range invalidIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid PushID: %s", id)
		}
	}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Errorf("Failed to generate PushID: %v", err)
	}
	
	if len(generated) != 20 {
		t.Errorf("Generated PushID should be 20 characters, got %d", len(generated))
	}
	
	if !parser.CanParse(generated) {
		t.Errorf("Generated PushID is not valid: %s", generated)
	}
}

// Benchmark tests for PushID parser
func BenchmarkPushIDParser(b *testing.B) {
	parser := &PushIDParser{}
	testID := "-M_Kj4tZ8vZjdHi6B-my"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(testID)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}